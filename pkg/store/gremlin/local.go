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

	err = s.Add(
		context.TODO(),
		store.TypeType.Thing(),
		store.EntityType.Thing(),
		store.RelationType.Thing(),
	)

	return s, err
}

func (s *localStore) execute(ctx context.Context, statement Statements) (interface{}, error) {
	log.Println(statement.String())
	return s.client.Execute(statement.String(), nil, nil)
}

func (s *localStore) Add(ctx context.Context, things ...store.Thingable) error {
	return s.AddAll(ctx, things)
}
func (s *localStore) AddAll(ctx context.Context, things []store.Thingable) error {
	st := Graph()

	for _, thingable := range things {
		thing := thingable.Thing()

		st = st.AddV(thing.Metadata.ID.String()).
			AddE("/relation/type_of").From(thing.Metadata.ID.String()).To(thing.Metadata.Type.String())

		if thing.Metadata.Type == store.TypeType.ID() {
			st = st.AddE("/relation/subtype_of").From(thing.Metadata.ID.String()).To(thing.Properties["parent"].(string))
		}
	}

	_, err := s.execute(ctx, Statements{st})

	return err
}

func (s *localStore) Len(ctx context.Context) (int, error) {
	query := Statements{
		Assign("g", Graph()),
		Var("g").V().Count(),
	}

	data, err := s.execute(ctx, query)

	results := data.([]interface{})
	values := results[0].([]interface{})
	value := values[0].(map[string]interface{})

	return int(value["@value"].(float64)), err
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
