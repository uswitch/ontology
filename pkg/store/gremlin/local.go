package gremlin

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/kr/pretty"
	"github.com/qasaur/gremgo"

	"github.com/uswitch/ontology/pkg/store"
)

const DATE_LAYOUT = "2006-01-02T15:04:05.000Z"

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

func (s *localStore) execute(ctx context.Context, statement Statements) ([]interface{}, error) {
	log.Println(statement.String())

	out, err := s.client.Execute(statement.String(), nil, nil)
	//pretty.Println(out, err)
	if err != nil {
		return nil, err
	}

	results, ok := out.([]interface{})
	if !ok {
		return nil, fmt.Errorf("Failed to get results from data: %v", out)
	}

	pretty.Println(results)

	if values, ok := results[0].([]interface{}); ok {
		return values, nil
	} else if err, ok := results[0].(error); ok {
		return nil, err
	} else if results[0] == nil {
		return []interface{}{}, nil
	} else {
		return nil, fmt.Errorf("Failed to get values from result: %v", results[0])
	}
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

		jsonProperties, err := json.Marshal(thing.Properties)
		if err != nil {
			return err
		}

		st = st.AddV(id).As(id).
			Property("name", thing.Metadata.Name).
			Property("updated_at", thing.Metadata.UpdatedAt.Format(DATE_LAYOUT)).
			Property("properties", string(jsonProperties)).
			AddE(store.TypeOfType.ID().String()).To(Var("g").V().HasLabel(thing.Metadata.Type.String()))

		if parentID, hasParent := thing.Properties["parent"].(string); hasParent && thing.Metadata.Type == store.TypeType.ID() {
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

	values, err := s.execute(ctx, query)

	value := values[0].(map[string]interface{})

	return int(value["@value"].(float64)), err
}

func (s *localStore) Types(ctx context.Context, thingable store.Thingable) ([]*store.Type, error) {
	/*thing := thingable.Thing()

	data, err := s.execute(ctx, Statements{
		Graph().V().
			HasLabel(thing.Thing().ID().String()).
			Repeat(
				BothE(store.TypeOfType.ID().String()).OtherV().SimplePath(),
			).
			Emit(),
	})*/

	return nil, store.ErrUnimplemented
}

func (s *localStore) TypeHierarchy(ctx context.Context, typ *store.Type) ([]*store.Type, error) {
	return s.typeHierarchy(ctx, typ.ID())
}

func thingQuery(s Statement) Statement {
	return s.As("thing").OutE(store.TypeOfType.ID().String()).OtherV().As("type").Select("thing", "type")
}

func propertiesLoader(rawProps map[string]interface{}) (map[string]interface{}, error) {
	props := map[string]interface{}{}

	for k, rawProp := range rawProps {
		randomArray, ok := rawProp.([]interface{})
		if !ok {
			return nil, fmt.Errorf("Failed to cast random array")
		}

		vertexProp, ok := randomArray[0].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("Failed to cast vertex prop")
		}

		vpValue, ok := vertexProp["@value"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("Failed to cast vpCalue")
		}

		props[k] = vpValue["value"]
	}

	return props, nil
}

func vertexLoader(v map[string]interface{}) (*store.Thing, error) {
	values, ok := v["@value"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Failed to get @value: %v", v["@value"])
	}

	id, ok := values["label"].(string)
	if !ok {
		return nil, fmt.Errorf("Failed to cast label: %v", values["label"])
	}

	// TODO: pull in name, updated at and properties
	rawProps, ok := values["properties"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("failed to cast properties: %v", values["properties"])
	}

	props, err := propertiesLoader(rawProps)
	if err != nil {
		return nil, err
	}

	name, ok := props["name"].(string)
	if !ok {
		return nil, fmt.Errorf("failed to cast name: %v", props["name"])
	}
	updatedAtString, ok := props["updated_at"].(string)
	if !ok {
		return nil, fmt.Errorf("failed to cast updated_at: %v", props["updated_at"])
	}
	jsonProperties, ok := props["properties"].(string)
	if !ok {
		return nil, fmt.Errorf("failed to cast json properties: %v", props["properties"])
	}

	updatedAt, err := time.Parse(DATE_LAYOUT, updatedAtString)
	if err != nil {
		return nil, err
	}

	thing := &store.Thing{
		Metadata: store.Metadata{
			ID:        store.ID(id),
			Name:      name,
			UpdatedAt: updatedAt,
		},
		Properties: store.Properties{},
	}

	if err := json.Unmarshal([]byte(jsonProperties), &thing.Properties); err != nil {
		return nil, err
	}

	return thing, nil
}

func thingLoader(datum map[string]interface{}) (*store.Thing, error) {
	rawTyp, ok := datum["type"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Failed to get type: %v", datum["type"])
	}
	typ, err := vertexLoader(rawTyp)
	if err != nil {
		return nil, err
	}

	rawThing, ok := datum["thing"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Failed to get thing: %v", datum["type"])
	}
	thing, err := vertexLoader(rawThing)
	if err != nil {
		return nil, err
	}

	thing.Metadata.Type = typ.Metadata.ID

	return thing, nil
}

func (s *localStore) typeHierarchy(ctx context.Context, typID store.ID) ([]*store.Type, error) {
	results, err := s.execute(ctx, Statements{
		thingQuery(Graph().V().
			HasLabel(typID.String()).
			As("a").
			Union(
				Select("a"),
				Repeat(
					OutE(store.SubtypeOfType.ID().String()).OtherV(),
				).
					Until(
						InE(store.SubtypeOfType.ID().String()).Count().Is(0),
					).
					Emit(),
			)),
	})

	if err != nil {
		return nil, err
	}

	typs := make([]*store.Type, len(results))

	for idx, rawEntry := range results {
		rawMap := rawEntry.(map[string]interface{})
		thing, err := thingLoader(rawMap)
		if err != nil {
			return nil, err
		}

		typs[idx] = (*store.Type)(thing)
	}

	return typs, nil
}

func (s *localStore) Inherits(context.Context, *store.Type, *store.Type) (bool, error) {
	return false, store.ErrUnimplemented
}

func (s *localStore) IsA(ctx context.Context, thingable store.Thingable, t *store.Type) (bool, error) {
	if t == store.TypeType {
		return thingable.Thing().Metadata.Type == t.Metadata.ID, nil
	}

	types, err := s.typeHierarchy(ctx, thingable.Thing().Metadata.Type)
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

func (s *localStore) Validate(ctx context.Context, t store.Thingable, opts store.ValidateOptions) ([]store.ValidationError, error) {
	return store.Validate(ctx, s, t, opts)
}

func (s *localStore) GetByID(ctx context.Context, idable store.IDable) (*store.Thing, error) {
	results, err := s.execute(ctx, Statements{
		thingQuery(Graph().V().
			HasLabel(idable.ID().String())),
	})

	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, store.ErrNotFound
	}

	rawMap, ok := results[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Failed to cast result")
	}

	return thingLoader(rawMap)
}

func (s *localStore) GetEntityByID(ctx context.Context, idable store.IDable) (*store.Entity, error) {
	if thing, err := s.GetByID(ctx, idable); err != nil {
		return nil, err
	} else if ok, err := s.IsA(ctx, thing, store.EntityType); !ok {
		return nil, store.ErrNotFound
	} else if err != nil {
		return nil, err
	} else {
		return (*store.Entity)(thing), nil
	}
}

func (s *localStore) GetRelationByID(ctx context.Context, idable store.IDable) (*store.Relation, error) {
	if thing, err := s.GetByID(ctx, idable); err != nil {
		return nil, err
	} else if ok, err := s.IsA(ctx, thing, store.RelationType); !ok {
		return nil, store.ErrNotFound
	} else if err != nil {
		return nil, err
	} else {
		return (*store.Relation)(thing), nil
	}
}

func (s *localStore) GetTypeByID(ctx context.Context, idable store.IDable) (*store.Type, error) {
	if thing, err := s.GetByID(ctx, idable); err != nil {
		return nil, err
	} else if ok, err := s.IsA(ctx, thing, store.TypeType); !ok {
		return nil, store.ErrNotFound
	} else if err != nil {
		return nil, err
	} else {
		return (*store.Type)(thing), nil
	}
}

func (s *localStore) List(ctx context.Context, options store.ListOptions) ([]*store.Thing, error) {
	return s.ListByType(ctx, nil, options)
}

func (s *localStore) ListByType(ctx context.Context, typ *store.Type, options store.ListOptions) ([]*store.Thing, error) {
	if options.NumberOfResults == 0 {
		options.NumberOfResults = store.DefaultNumberOfResults
	}

	var order string

	switch options.SortOrder {
	case store.SortAscending:
		order = "asc"
	case store.SortDescending:
		order = "desc"
	default:
		return []*store.Thing{}, store.ErrUnimplemented
	}

	query := Graph().V()

	if typ != nil {
		typID := typ.ID()
		query = Graph().V().
			HasLabel(typID.String()).
			As("a").
			Union(
				Select("a"),
				Repeat(
					OutE(store.SubtypeOfType.ID().String()).OtherV(),
				).
					Until(
						InE(store.SubtypeOfType.ID().String()).Count().Is(0),
					).
					Emit(),
			).
			InE(store.TypeOfType.ID().String()).OtherV()
	}

	results, err := s.execute(ctx, Statements{
		thingQuery(
			query.Order().By("label", order).
				Range(options.Offset, options.Offset+options.NumberOfResults),
		),
	})

	if err != nil {
		return nil, err
	}

	things := make([]*store.Thing, len(results))

	for idx, rawEntry := range results {
		rawMap := rawEntry.(map[string]interface{})
		thing, err := thingLoader(rawMap)
		if err != nil {
			return nil, err
		}

		things[idx] = thing
	}

	return things, nil
}

func (s *localStore) ListEntities(ctx context.Context, options store.ListOptions) ([]*store.Entity, error) {
	things, err := s.ListByType(ctx, store.EntityType, options)
	entities := make([]*store.Entity, len(things))

	if err != nil {
		return entities, nil
	}

	for idx, thing := range things {
		entities[idx] = (*store.Entity)(thing)
	}

	return entities, nil
}

func (s *localStore) ListRelations(ctx context.Context, options store.ListOptions) ([]*store.Relation, error) {
	things, err := s.ListByType(ctx, store.RelationType, options)
	relations := make([]*store.Relation, len(things))

	if err != nil {
		return relations, nil
	}

	for idx, thing := range things {
		relations[idx] = (*store.Relation)(thing)
	}

	return relations, nil
}

func (s *localStore) ListTypes(ctx context.Context, options store.ListOptions) ([]*store.Type, error) {
	things, err := s.ListByType(ctx, store.TypeType, options)
	types := make([]*store.Type, len(things))

	if err != nil {
		return types, nil
	}

	for idx, thing := range things {
		types[idx] = (*store.Type)(thing)
	}

	return types, nil
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
