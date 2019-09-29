package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
)

type OnOperationFunc func(context.Context, OperationParams) (chan *OperationResult, error)

type ServerChannel struct {
	OnOperation OnOperationFunc

	ch    MessageReaderWriter
	read  <-chan *OperationMessage
	write chan<- *OperationMessage
}

func NewServerChannel(ctx context.Context, ch MessageReaderWriter, fn OnOperationFunc) *ServerChannel {
	read, write := OperationStream(ctx, ch)

	return &ServerChannel{fn, ch, read, write}
}

func (s *ServerChannel) Accept(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("receiving init message: %v", ctx.Err())
	case op := <-s.read:
		if op.Type != GQL_CONNECTION_INIT {
			return fmt.Errorf("Expecting GQL_CONNECTION_INIT, got %s", op.Type)
		}
	}

	select {
	case <-ctx.Done():
		return fmt.Errorf("Sending ack message: %v", ctx.Err())
	case s.write <- &OperationMessage{GQL_CONNECTION_ACK, nil, nil}:
	}

	return nil
}

func (s *ServerChannel) Listen(ctx context.Context) error {
	idToCancel := map[MessageID]context.CancelFunc{}
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("server listen context done: %v", ctx.Err())
		case op := <-s.read:
			if op == nil {
				return fmt.Errorf("server has closed")
			}

			switch op.Type {
			case GQL_CONNECTION_TERMINATE:
				return nil
			case GQL_START:
				if op.ID == nil {
					log.Printf("server received a start with no id, discarding.")
					continue
				}

				var params OperationParams

				pipeR, pipeW := io.Pipe()

				enc := json.NewEncoder(pipeW)
				dec := json.NewDecoder(pipeR)

				go enc.Encode(op.Payload)

				// TODO catch the encoding errors!
				/*if err := enc.Encode(op.Payload); err != nil {
					log.Printf("Failed to encode start payload: %v", err)
					continue
				}*/

				if err := dec.Decode(&params); err != nil {
					log.Printf("Failed to decode start payload: %v", err)
					continue
				}

				if ch, err := s.OnOperation(ctx, params); err != nil {
					log.Printf("failed to start operation: %v", err)
				} else {
					streamCtx, cancel := context.WithCancel(ctx)
					idToCancel[*op.ID] = cancel
					go s.streamResults(streamCtx, *op.ID, ch)
				}
			case GQL_STOP:
				if op.ID == nil {
					log.Printf("server received a stop with no id, discarding.")
					continue
				}

				if cancel, ok := idToCancel[*op.ID]; !ok {
					log.Printf("server received a stop for id with no cancel, discarding")
					continue
				} else {
					cancel()
					delete(idToCancel, *op.ID)
				}

			default:
				log.Printf("server got unknown operation type %s", op.Type)
			}
		}
	}
}

func (s *ServerChannel) streamResults(ctx context.Context, id MessageID, ch <-chan *OperationResult) {
	log.Printf("result streaming starting for %s", id)
	defer func() {
		log.Printf("result streaming done for %s", id)
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case result := <-ch:
			if result == nil {
				select {
				case <-ctx.Done():
					return
				case s.write <- &OperationMessage{GQL_COMPLETE, &id, nil}:
				}

				return
			} else {
				select {
				case <-ctx.Done():
					return
				case s.write <- &OperationMessage{GQL_DATA, &id, result}:
				}
			}
		}
	}
}
