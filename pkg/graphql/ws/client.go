package ws

import (
	"context"
	"fmt"
	"log"
	"sync"
)

type ClientChannel struct {
	ch    MessageReaderWriter
	read  <-chan *OperationMessage
	write chan<- *OperationMessage

	idRead     map[MessageID]chan<- *OperationMessage
	idReadLock sync.RWMutex

	ready bool
}

func NewClientChannel(ctx context.Context, ch MessageReaderWriter) *ClientChannel {
	rawRead, write := OperationStream(ctx, ch)
	nonIDRead := make(chan *OperationMessage)

	c := &ClientChannel{
		ch, nonIDRead, write,
		map[MessageID]chan<- *OperationMessage{},
		sync.RWMutex{},
		false,
	}

	go c.filterDataMessages(ctx, rawRead, nonIDRead)

	return c
}

func (c *ClientChannel) filterDataMessages(ctx context.Context, rawRead <-chan *OperationMessage, nonIDRead chan<- *OperationMessage) {
	for {
		select {
		case <-ctx.Done():
			return
		case op := <-rawRead:
			if op.ID == nil {
				select {
				case <-ctx.Done():
					return
				case nonIDRead <- op:
				}
			} else {
				id := *op.ID

				c.idReadLock.RLock()
				ch, ok := c.idRead[id]
				c.idReadLock.RUnlock()

				if ok {
					select {
					case <-ctx.Done():
						return
					case ch <- op:
					}
				} else {
					log.Printf("No read for id %s dropping %v", id, op)
				}
			}
		}
	}
}

func (c *ClientChannel) registerIDReader(id MessageID, in chan<- *OperationMessage) error {
	c.idReadLock.Lock()
	defer c.idReadLock.Unlock()

	if _, ok := c.idRead[id]; ok {
		return fmt.Errorf("There is already a registered reader for %v", id)
	}

	c.idRead[id] = in

	return nil
}
func (c *ClientChannel) unregisterIDReader(id MessageID) error {
	c.idReadLock.Lock()
	defer c.idReadLock.Unlock()

	if _, ok := c.idRead[id]; !ok {
		return fmt.Errorf("There isn't a registered reader for %v", id)
	}

	delete(c.idRead, id)

	return nil
}

func (c *ClientChannel) Connect(ctx context.Context, _ ConnectionParams) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("Sending init message: %v", ctx.Err())
	case c.write <- &OperationMessage{GQL_CONNECTION_INIT, nil, nil}:
	}

	select {
	case <-ctx.Done():
		return fmt.Errorf("receiving ack message: %v", ctx.Err())
	case op := <-c.read:
		if op.Type != GQL_CONNECTION_ACK {
			return fmt.Errorf("Expected GQL_CONNECTION_ACK, but got %s", op.Type)
		}
	}

	c.ready = true

	return nil
}

func (c *ClientChannel) Operation(ctx context.Context, params OperationParams) (<-chan interface{}, error) {
	if !c.ready {
		return nil, fmt.Errorf("The client isn't connected")
	}

	id := MessageID("wibble") // TODO: generate this value

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("Sending start message: %v", ctx.Err())
	case c.write <- &OperationMessage{GQL_START, &id, params}:
	}

	in := make(chan *OperationMessage)
	out := make(chan interface{})

	go func() {
		for {
			select {
			case <-ctx.Done():
				c.unregisterIDReader(id)
				close(out)
				close(in)
				return
			case msg := <-in:
				if msg == nil {
					close(out)
					return
				}

				switch msg.Type {
				case GQL_COMPLETE:
					close(out)
					return
				case GQL_DATA:
					select {
					case <-ctx.Done():
						c.unregisterIDReader(id)
						close(out)
						close(in)
						return
					case out <- msg.Payload:
					}
				default:
					log.Printf("message of unknown type for %s: %s", id, msg.Type)
				}
			}
		}
	}()

	c.registerIDReader(id, in)

	return out, nil
}

func (c *ClientChannel) Close(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("Sending terminate message: %v", ctx.Err())
	case c.write <- &OperationMessage{GQL_CONNECTION_TERMINATE, nil, nil}:
	}

	return nil
}
