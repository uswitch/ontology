package gremlin

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/qasaur/gremgo"

	"github.com/uswitch/ontology/pkg/store"
	"github.com/uswitch/ontology/pkg/types"
	"github.com/uswitch/ontology/pkg/types/entity"
	"github.com/uswitch/ontology/pkg/types/relation"
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

func (l *local) execute(ctx context.Context, statement Statement) ([]interface{}, error) {
	log.Println(statement.String())

	out, err := l.client.Execute(statement.String(), nil, nil)
	//pretty.Println(out, err)
	if err != nil {
		return nil, err
	}

	results, ok := out.([]interface{})
	if !ok {
		return nil, fmt.Errorf("Failed to get results from data: %v", out)
	}

	//pretty.Println(results)

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

func (l *local) Add(ctx context.Context, instances ...types.Instance) error {
	st := Var("g")

	for i, instance := range instances {
		serialized, err := json.Marshal(instance)
		if err != nil {
			return err
		}

		if types.IsA(instance, entity.ID) {
			st = st.V().Has(String("entity"), String("id"), String(instance.ID())).Fold().
				Coalesce(
					Unfold(),
					AddV(String("entity")),
				)
		} else if types.IsA(instance, relation.ID) {
			if instance, ok := instance.(relation.Instance); ok {
				st = st.V().Has(String("entity"), String("id"), String(instance.A())).Fold().
					Coalesce(
						Unfold(),
						AddV(String("entity")).Property(String("id"), String(instance.A())),
					).As("start").
					Map(V().Has(String("entity"), String("id"), String(instance.B())).Fold()).
					Coalesce(
						Unfold(),
						AddV(String("entity")).Property(String("id"), String(instance.B())),
					).
					Coalesce(
						InE(String(instance.ID())).Where(OutV().As("start")),
						AddE(String(instance.ID())).From(String("start")),
					)
			} else {
				log.Printf("instance doesn't conform to relation.Instance: %T", instance)
			}

		} else {
			log.Printf("a rather strange type of instance: %s", instance.Type())
		}

		st = st.Property(String("id"), String(instance.ID())).
			Property(String("type"), String(instance.Type())).
			Property(String("_serialized"), String(serialized))

		// we can only execute so many at a time. frame size is 65536.
		if i%30 == 0 {
			if _, err := l.execute(ctx, st); err != nil {
				return err
			} else {
				st = Var("g")
			}
		}
		/*if _, err := l.execute(ctx, st); err != nil {
			return err
		}*/
	}

	if _, err := l.execute(ctx, st); err != nil {
		return err
	}

	return nil
}

func propertiesLoader(rawProps map[string]interface{}) (map[string]interface{}, error) {
	props := map[string]interface{}{}

	for k, rawProp := range rawProps {
		var vertexProp map[string]interface{}

		switch typedProp := rawProp.(type) {
		case []interface{}:
			var ok bool

			vertexProp, ok = typedProp[0].(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("Failed to cast vertex prop")
			}
		case map[string]interface{}:
			vertexProp = typedProp
		default:
			return nil, fmt.Errorf("Failed to parse property, unknown type: %T!", rawProp)
		}

		vpValue, ok := vertexProp["@value"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("Failed to cast vpCalue")
		}

		props[k] = vpValue["value"]
	}

	return props, nil
}

func loader(v map[string]interface{}) (types.Instance, error) {
	values, ok := v["@value"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Failed to get @value: %v", v["@value"])
	}

	rawProps, ok := values["properties"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("failed to cast properties: %v", values["properties"])
	}

	props, err := propertiesLoader(rawProps)
	if err != nil {
		return nil, err
	}

	serialized, ok := props["_serialized"].(string)
	if !ok {
		return nil, fmt.Errorf("failed to cast _serialized: %v", props["_serialized"])
	}

	//log.Println(serialized)

	return types.Parse(serialized)
}

func (l *local) getByStatement(ctx context.Context, st Statement) (types.Instance, error) {
	results, err := l.execute(ctx, st)

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

	return loader(rawMap)
}

func (l *local) Get(ctx context.Context, id types.ID) (types.Instance, error) {
	return l.getByStatement(
		ctx,
		G.V().Coalesce(
			Has(String("id"), String(id)),
			G.E().Has(String("id"), String(id)),
		),
	)
}

func (l *local) listByStatement(ctx context.Context, st Statement) ([]types.Instance, error) {
	results, err := l.execute(ctx, st)

	if err != nil {
		return nil, err
	}

	instances := make([]types.Instance, len(results))

	for idx, rawEntry := range results {
		rawMap := rawEntry.(map[string]interface{})
		instance, err := loader(rawMap)
		if err != nil {
			return nil, err
		}

		instances[idx] = instance
	}

	return instances, nil
}

func (l *local) ListByType(ctx context.Context, id types.ID) ([]types.Instance, error) {
	subclasses := types.SubclassesOf(id)
	classStatements := make([]Statement, len(subclasses)+1)

	classStatements[0] = String(id)

	for idx, subclass := range subclasses {
		classStatements[idx+1] = String(subclass)
	}

	if types.InheritsFrom(id, entity.ID) {
		return l.listByStatement(ctx, G.V().Has(String("type"), Within(classStatements...)))
	} else if types.InheritsFrom(id, relation.ID) {
		return l.listByStatement(ctx, G.E().Has(String("type"), Within(classStatements...)))
	}

	return nil, fmt.Errorf("type '%s' isn't an entity or relation", id)
}

func (l *local) ListFromByType(ctx context.Context, rootID types.ID, typeID types.ID) ([]types.Instance, error) {
	subclasses := types.SubclassesOf(typeID)
	classStatements := make([]Statement, len(subclasses)+1)

	classStatements[0] = String(typeID)

	for idx, subclass := range subclasses {
		classStatements[idx+1] = String(subclass)
	}

	if types.InheritsFrom(typeID, entity.ID) {
		return l.listByStatement(
			ctx,
			G.V().Has(String("id"), String(rootID)).
				Repeat(Out()).Times(Int(2)).Emit().Dedup().
				Has(String("type"), Within(classStatements...)),
		)
	} else if types.InheritsFrom(typeID, relation.ID) {
		return nil, nil
	}

	return nil, fmt.Errorf("type '%s' isn't an entity or relation", typeID)
}
