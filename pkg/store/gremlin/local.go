package gremlin

import (
	"context"
	"log"

	"github.com/qasaur/gremgo"

	"github.com/uswitch/ontology/pkg/store"
)

type localStore struct {
	typeBroadcast *store.Broadcast
	idBroadcast   *store.Broadcast

	client gremgo.Client
}

func NewLocalServer(url string) (store.Store, error) {
	errs := make(chan error)
	go func(chan error) {
		err := <-errs
		log.Fatal("Lost connection to the database: " + err.Error())
	}(errs) // Example of connection error handling logic

	dialer := gremgo.NewDialer("ws://127.0.0.1:8182") // Returns a WebSocket dialer to connect to Gremlin Server
	g, err := gremgo.Dial(dialer, errs)               // Returns a gremgo client to interact with
	if err != nil {
		return nil, err
	}

	s := &localStore{
		typeBroadcast: store.NewBroadcast(),
		idBroadcast:   store.NewBroadcast(),

		client: g,
	}

	return s, nil
}

func (s *localStore) Add(ctx context.Context, things ...store.Thingable) error {
	return s.AddAll(ctx, things)
}
func (s *localStore) AddAll(context.Context, []store.Thingable) error {
	return store.ErrUnimplemented
}

func (s *localStore) Len(context.Context) (int, error) {
	return 0, store.ErrUnimplemented
}

func (s *localStore) Types(context.Context, store.Thingable) ([]*store.Type, error) {
	return nil, store.ErrUnimplemented
}

func (s *localStore) TypeHierarchy(context.Context, *store.Type) ([]*store.Type, error) {
	return nil, store.ErrUnimplemented
}

func (s *localStore) Inherits(context.Context, *store.Type, *store.Type) (bool, error) {
	return false, store.ErrUnimplemented
}

func (s *localStore) IsA(context.Context, store.Thingable, *store.Type) (bool, error) {
	return false, store.ErrUnimplemented
}

func (s *localStore) Validate(context.Context, store.Thingable, store.ValidateOptions) ([]store.ValidationError, error) {
	return nil, store.ErrUnimplemented
}

func (s *localStore) GetByID(context.Context, store.IDable) (*store.Thing, error) {
	return nil, store.ErrUnimplemented
}

func (s *localStore) GetEntityByID(context.Context, store.IDable) (*store.Entity, error) {
	return nil, store.ErrUnimplemented
}

func (s *localStore) GetRelationByID(context.Context, store.IDable) (*store.Relation, error) {
	return nil, store.ErrUnimplemented
}

func (s *localStore) GetTypeByID(context.Context, store.IDable) (*store.Type, error) {
	return nil, store.ErrUnimplemented
}

func (s *localStore) List(context.Context, store.ListOptions) ([]*store.Thing, error) {
	return nil, store.ErrUnimplemented
}

func (s *localStore) ListByType(context.Context, *store.Type, store.ListOptions) ([]*store.Thing, error) {
	return nil, store.ErrUnimplemented
}

func (s *localStore) ListEntities(context.Context, store.ListOptions) ([]*store.Entity, error) {
	return nil, store.ErrUnimplemented
}

func (s *localStore) ListRelations(context.Context, store.ListOptions) ([]*store.Relation, error) {
	return nil, store.ErrUnimplemented
}

func (s *localStore) ListTypes(context.Context, store.ListOptions) ([]*store.Type, error) {
	return nil, store.ErrUnimplemented
}

func (s *localStore) ListRelationsForEntity(context.Context, *store.Type, *store.Entity, store.ListOptions) ([]*store.Relation, error) {
	return nil, store.ErrUnimplemented
}

func (s *localStore) WatchByID(context.Context, store.IDable) (chan *store.Thing, error) {
	return nil, store.ErrUnimplemented
}

func (s *localStore) WatchByType(context.Context, store.IDable) (chan *store.Thing, error) {
	return nil, store.ErrUnimplemented
}
