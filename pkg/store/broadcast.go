package store

import (
	"context"
	"log"
	"sync"
)

type broadcast struct {
	registered map[ID][]chan *Thing

	rw sync.RWMutex
}

func newBroadcast() *broadcast {
	return &broadcast{
		registered: map[ID][]chan *Thing{},
	}
}

func (b *broadcast) Register(ctx context.Context, id ID) (chan *Thing, error) {
	b.rw.Lock()
	defer b.rw.Unlock()

	ch := make(chan *Thing)

	b.registered[id] = append(b.registered[id], ch)

	go func(id ID, ch chan *Thing) {
		select {
		case <-ctx.Done():
			b.rw.Lock()
			oldRegistered := b.registered[id]
			newRegistered := []chan*Thing{}
			for _, regCh := range oldRegistered {
				if regCh != ch {
					newRegistered = append(newRegistered, ch)
				}
			}

			b.registered[id] = newRegistered
			b.rw.Unlock()

			close(ch)
		}
	}(id, ch)

	return ch, nil
}

func (b *broadcast) Send(ctx context.Context, thingable Thingable, ids ...ID) error {
	b.rw.RLock()
	defer b.rw.RUnlock()

	var wg sync.WaitGroup
	thing := thingable.Thing()

	for _, id := range ids {
		chans := b.registered[id]
		wg.Add(len(chans))
		for _, ch := range chans {
			go func(wg *sync.WaitGroup, ch chan *Thing) {
				defer func() {
					if r := recover(); r != nil {
						log.Printf("write: error writing %v on channel: %v\n", ch, r)
					}

					wg.Done()
				}()

				ch <- thing
			}(&wg, ch)
		}
	}

	wg.Wait()

	return nil
}
