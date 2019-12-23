package gremlin

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/schwartzmx/gremtune"

	"github.com/uswitch/ontology/pkg/store"
	"github.com/uswitch/ontology/pkg/types"
	"github.com/uswitch/ontology/pkg/types/entity"
	"github.com/uswitch/ontology/pkg/types/relation"
)

const DATE_LAYOUT = "2006-01-02T15:04:05.000Z"

type local struct {
	client gremtune.Client
}

func NewLocalServer(url string) (store.Store, error) {
	errs := make(chan error)
	go func(chan error) {
		err := <-errs
		log.Fatal("Lost connection to the database: " + err.Error())
	}(errs) // Example of connection error handling logic

	dialer := gremtune.NewDialer(url)     // Returns a WebSocket dialer to connect to Gremlin Server
	g, err := gremtune.Dial(dialer, errs) // Returns a gremgo client to interact with
	if err != nil {
		return nil, err
	}

	return &local{
		client: g,
	}, err
}

func (l *local) execute(ctx context.Context, statement Statement) ([]GenericValue, error) {
	//log.Println(statement.String())

	out, err := l.client.Execute(statement.String())
	if err != nil {
		return nil, err
	}

	var listValue GenericValue
	if err := json.Unmarshal(out[0].Result.Data, &listValue); err != nil {
		return nil, err
	}

	if listValue.Type != "g:List" {
		return nil, fmt.Errorf("expecting a list, got %s", listValue.Type)
	}

	items := []GenericValue{}
	if err := json.Unmarshal(listValue.Value, &items); err != nil {
		return nil, err
	}

	return items, nil
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

func loader(val GenericValue) (types.Instance, error) {
	var rawSerialized json.RawMessage

	switch val.Type {
	case "g:Vertex":
		var vertex VertexValue

		if err := json.Unmarshal(val.Value, &vertex); err != nil {
			log.Println(string(val.Value))
			return nil, err
		}

		rawSerialized = vertex.Properties["_serialized"][0].Value.Value
	case "g:Edge":
		var edge EdgeValue

		if err := json.Unmarshal(val.Value, &edge); err != nil {
			return nil, err
		}

		rawSerialized = edge.Properties["_serialized"].Value.Value
	default:
		return nil, fmt.Errorf("unknown value type: %s", val.Type)
	}

	var serialized string

	if err := json.Unmarshal(rawSerialized, &serialized); err != nil {
		return nil, err
	}

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

	return loader(results[0])
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

	for idx, result := range results {
		instance, err := loader(result)
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
