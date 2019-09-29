package ws

import (
	"context"
	"encoding/json"
	"log"
)

func OperationStream(ctx context.Context, rw MessageReaderWriter) (<-chan *OperationMessage, chan<- *OperationMessage) {
	outCh := make(chan *OperationMessage)

	go func(r MessageReader) {
		for {
			select {
			case <-ctx.Done():
				log.Printf("[READ] context done: %v", ctx.Err())
				close(outCh)
				return
			default:
				if _, data, err := r.ReadMessage(); err != nil {
					// any error from ReadMessage should be interpretted as the underlying
					// channel has failed
					log.Printf("[READ] failed read: %v", err)
					close(outCh)
					return
				} else {
					var op OperationMessage

					if err = json.Unmarshal(data, &op); err != nil {
						log.Printf("[READ] Invalid operation recieved: %v", err)
						continue
					}

					outCh <- &op
				}
			}
		}
	}(rw)

	inCh := make(chan *OperationMessage)

	go func(w MessageWriter) {
		for {
			select {
			case <-ctx.Done():
				log.Printf("[WRITE] context done: %v", ctx.Err())
				close(inCh)
				return
			case op := <-inCh:
				if op == nil {
					return
				}

				bytes, err := json.Marshal(op)
				if err != nil {
					log.Printf("[WRITE] Invalid operation recieved: %v", err)
					continue
				}

				if err := w.WriteMessage(1, bytes); err != nil {
					log.Printf("[WRITE] Error writing message: %v", err)
					continue
				}
			}
		}
	}(rw)

	return outCh, inCh
}
