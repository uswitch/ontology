package store

import (
	"context"
	"sync"
	"testing"
)

func appendFromChannel(ch chan *Thing, list *[]*Thing, wg *sync.WaitGroup) {
	for ;; {
		if thing := <- ch; thing == nil {
			wg.Done()
			return
		} else {
			*list = append(*list, thing)
		}
	}
}

func TestBroadcastSendRecieve(t *testing.T) {
	b := newBroadcast()

	closedWG := sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())

	closedWG.Add(2)

	list1 := []*Thing{}
	if ch, err := b.Register(ctx, ID("/foo")); err != nil {
		t.Fatalf("Couldn't watch type: %v", err)
	} else {
		go appendFromChannel(ch, &list1, &closedWG)
	}

	list2 := []*Thing{}
	if ch, err := b.Register(ctx, ID("/bar")); err != nil {
		t.Fatalf("Couldn't watch type: %v", err)
	} else {
		go appendFromChannel(ch, &list2, &closedWG)
	}

	b.Send(ctx, entity("1"), ID("/foo"))
	b.Send(ctx, entity("2"), ID("/foo"), ID("/bar"))

	cancel()
	closedWG.Wait()

	if expected := 2; len(list1) != expected {
		t.Errorf("Expected there to be %d types, but it was %d", expected, len(list1))
	}
	if expected := 1; len(list2) != expected {
		t.Errorf("Expected there to be %d types, but it was %d", expected, len(list2))
	}

	b.Send(ctx, entity("3"), ID("/foo"))
	b.Send(ctx, entity("4"), ID("/foo"), ID("/bar"))

	if expected := 2; len(list1) != expected {
		t.Errorf("Expected there to be %d types, but it was %d", expected, len(list1))
	}
	if expected := 1; len(list2) != expected {
		t.Errorf("Expected there to be %d types, but it was %d", expected, len(list2))
	}

}
