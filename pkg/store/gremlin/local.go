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
		store.TypeOfType.Thing(),
		store.SubtypeOfType.Thing(),
	)

	return s, err
}

func (s *localStore) execute(ctx context.Context, statement Statements) (interface{}, error) {
	log.Println(statement.String())
	out, err := s.client.Execute(statement.String(), nil, nil)
	log.Println(out, err)
	return out, err
}

func (s *localStore) Add(ctx context.Context, things ...store.Thingable) error {
	return s.AddAll(ctx, things)
}
func (s *localStore) AddAll(ctx context.Context, things []store.Thingable) error {
	//vStatement := Graph()
	st := Var("g")

	for _, thingable := range things {
		thing := thingable.Thing()
		id := thing.Metadata.ID.String()

		st = st.AddV(id).As(id).AddE(store.TypeOfType.ID().String()).To(Var("g").V().HasLabel(thing.Metadata.Type.String()))

		if parentID, hasParent := thing.Properties["parent"].(string); hasParent && thing.Metadata.Type == store.TypeType.ID() {
			log.Println(parentID)
			st = st.OutV().AddE(store.SubtypeOfType.ID().String()).To(Var("g").V().HasLabel(parentID))
		}
	}

	_, err := s.execute(ctx, Statements{
		Assign("g", Graph()),
		st,
	})

	return err
}

func (s *localStore) Len(ctx context.Context) (int, error) {
	query := Statements{
		Graph().V().Count(),
	}

	data, err := s.execute(ctx, query)

	results := data.([]interface{})
	values := results[0].([]interface{})
	value := values[0].(map[string]interface{})

	return int(value["@value"].(float64)), err
}

func (s *localStore) Types(ctx context.Context, thingable store.Thingable) ([]*store.Type, error) {
	thing := thingable.Thing()

	data, err := s.execute(ctx, Statements{
		Graph().V().
			HasLabel(thing.Thing().ID().String()).
			OutE(store.TypeOfType.ID().String()).
			InV(),
	})
	log.Println(data, err)

	return nil, store.ErrUnimplemented
}

func (s *localStore) TypeHierarchy(context.Context, *store.Type) ([]*store.Type, error) {
	return nil, store.ErrUnimplemented
}

func (s *localStore) Inherits(context.Context, *store.Type, *store.Type) (bool, error) {
	return false, store.ErrUnimplemented
}

func (s *localStore) IsA(ctx context.Context, thingable store.Thingable, t *store.Type) (bool, error) {
	if t == store.TypeType {
		return thingable.Thing().Metadata.Type == t.Metadata.ID, nil
	}

	types, err := s.Types(ctx, thingable)
	if err != nil {
		return false, err
	}

	for _, typ := range types {
		if typ.Metadata.ID == t.Metadata.ID {
			return true, nil
		}
	}

	return false, nil
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
