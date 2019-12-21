package gremlin

import (
	"log"

	"github.com/qasaur/gremgo"

	"github.com/uswitch/ontology/pkg/store"
	"github.com/uswitch/ontology/pkg/types"
)

const DATE_LAYOUT = "2006-01-02T15:04:05.000Z"

type local struct {
	client gremgo.Client
}

func NewLocalServer(url string) (store.Store, error) {
	errs := make(chan error)
	go func(chan error) {
		err := <-errs
		log.Fatal("Lost connection to the database: " + err.Error())
	}(errs) // Example of connection error handling logic

	dialer := gremgo.NewDialer(url)     // Returns a WebSocket dialer to connect to Gremlin Server
	g, err := gremgo.Dial(dialer, errs) // Returns a gremgo client to interact with
	if err != nil {
		return nil, err
	}

	return &local{
		client: g,
	}, err
}

func (l *local) Add(types.Instance) error {
	return nil
}
func (l *local) Get(types.ID) (types.Instance, error) {
	return nil, nil
}
func (l *local) ListByType(types.ID) ([]types.Instance, error) {
	return nil, nil
}
