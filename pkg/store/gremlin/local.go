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

	dialer := gremtune.NewDialer(url, gremtune.SetBufferSize(1_000_000, 1_000_000)) // Returns a WebSocket dialer to connect to Gremlin Server
	g, err := gremtune.Dial(dialer, errs)                                           // Returns a gremgo client to interact with
	if err != nil {
		return nil, err
	}

	return &local{
		client: g,
	}, err
}

func (l *local) execute(ctx context.Context, statement Statement) ([]GenericValue, error) {
	//log.Println(statement.String())

	responseCh := make(chan gremtune.AsyncResponse, 10)

	if err := l.client.ExecuteAsync(statement.String(), responseCh); err != nil {
		return nil, err
	}

	items := []GenericValue{}

	for {
		select {
		case response, ok := <-responseCh:
			if !ok {
				return items, nil
			}

			if response.ErrorMessage != "" {
				log.Printf("response.ErrorMessage: %v", response.ErrorMessage)
			}

			if response.Response.Status.Code == 204 {
				return items, nil
			}

			var listValue GenericValue
			if err := json.Unmarshal(response.Response.Result.Data, &listValue); err != nil {
				return nil, fmt.Errorf("listValue unmarshal: %w", err)
			}

			if listValue.Type != "g:List" {
				return nil, fmt.Errorf("expecting a list, got %s %v", listValue.Type, response)
			}

			responseItems := []GenericValue{}
			if err := json.Unmarshal(listValue.Value, &responseItems); err != nil {
				return nil, fmt.Errorf("responseItems unmarshal: %w", err)
			}

			items = append(items, responseItems...)
		case <-ctx.Done():
			return nil, ctx.Err()
		}
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
			return nil, fmt.Errorf("vertex value unmarshal: %w", err)
		}

		rawSerialized = vertex.Properties["_serialized"][0].Value.Value
	case "g:Edge":
		var edge EdgeValue

		if err := json.Unmarshal(val.Value, &edge); err != nil {
			return nil, fmt.Errorf("edge value unmarshal: %w", err)
		}

		rawSerialized = edge.Properties["_serialized"].Value.Value
	default:
		return nil, fmt.Errorf("unknown value type: %s", val.Type)
	}

	var serialized string

	if err := json.Unmarshal(rawSerialized, &serialized); err != nil {
		return nil, fmt.Errorf("serialied unmarshal: %w", err)
	}

	if parsed, err := types.Parse(serialized); err != nil {
		return nil, fmt.Errorf("error parsing serialized form: %w", err)
	} else {
		return parsed, nil
	}
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
			return nil, fmt.Errorf("instance loader error: %w", err)
		}

		instances[idx] = instance
	}

	return instances, nil
}

func expandTypes(typeList []types.ID) ([]types.ID, error) {
	typeMap := map[types.ID]bool{}

	for _, typ := range typeList {
		subclasses := types.SubclassesOf(typ)

		typeMap[typ] = true
		for _, subclass := range subclasses {
			typeMap[subclass] = true
		}
	}

	firstIsAnEntity := types.InheritsFrom(typeList[0], entity.ID)

	outList := make([]types.ID, len(typeMap))
	idx := 0
	for typ, _ := range typeMap {
		if isAnEntity := types.InheritsFrom(typ, entity.ID); isAnEntity != firstIsAnEntity {
			return nil, fmt.Errorf("all types need to be entity or relation types, cannot be mixed")
		}

		outList[idx] = typ
		idx = idx + 1
	}

	return outList, nil
}

func typesToStatements(typeList []types.ID) []Statement {
	out := make([]Statement, len(typeList))
	for idx, typ := range typeList {
		out[idx] = String(typ)
	}
	return out
}

func (l *local) ListByType(ctx context.Context, typeIDs []types.ID, options store.ListByTypeOptions) ([]types.Instance, error) {
	var err error

	if options.IncludeSubclasses {
		if typeIDs, err = expandTypes(typeIDs); err != nil {
			return nil, fmt.Errorf("failed to expand sub types: %w", err)
		}
	}

	typeStatement := Within(typesToStatements(typeIDs)...)

	if types.InheritsFrom(typeIDs[0], entity.ID) {
		return l.listByStatement(ctx, G.V().Has(String("entity"), String("type"), typeStatement))
	} else if types.InheritsFrom(typeIDs[0], relation.ID) {
		return l.listByStatement(ctx, G.E().Has(String("type"), typeStatement))
	}

	return nil, fmt.Errorf("type '%s' isn't an entity or relation", typeIDs[0])
}

func (l *local) ListFromByType(ctx context.Context, rootID types.ID, typeIDs []types.ID, options store.ListFromByTypeOptions) ([]types.Instance, error) {
	var err error

	if options.IncludeSubclasses {
		if typeIDs, err = expandTypes(typeIDs); err != nil {
			return nil, fmt.Errorf("failed to expand sub types: %w", err)
		}
	}

	typeStatement := Within(typesToStatements(typeIDs)...)

	maxDepth := 2
	if options.MaxDepth > 0 {
		maxDepth = options.MaxDepth
	}

	constraintPredicate := Without()

	if numConstraints := len(options.ConstrainByType); numConstraints > 0 {
		constraintTypes := map[types.ID]bool{}

		for _, constraint := range options.ConstrainByType {
			constraintTypes[constraint] = true

			for _, subclass := range types.SubclassesOf(constraint) {
				constraintTypes[subclass] = true
			}
		}

		constraintStatements := []Statement{}

		for typ, _ := range constraintTypes {
			constraintStatements = append(constraintStatements, String(typ.String()))
		}

		constraintPredicate = Within(constraintStatements...)
	}

	var repeatStatement Statement

	switch options.Direction {
	case store.OutTraverseDirection:
		repeatStatement = OutE().Has(String("type"), constraintPredicate).As("edge").
			InV().Has(String("type"), constraintPredicate).As("vertex")
	case store.InTraverseDirection:
		repeatStatement = InE().Has(String("type"), constraintPredicate).As("edge").
			OutV().Has(String("type"), constraintPredicate).As("vertex")
	default:
		return nil, fmt.Errorf("unknown traverse direction: %v", options.Direction)
	}

	var selectType string

	if types.InheritsFrom(typeIDs[0], entity.ID) {
		selectType = "vertex"
	} else if types.InheritsFrom(typeIDs[0], relation.ID) {
		selectType = "edge"
	} else {
		return nil, fmt.Errorf("type '%s' isn't an entity or relation", typeIDs[0])
	}

	return l.listByStatement(
		ctx,
		G.V().Has(String("id"), String(rootID)).
			Repeat(repeatStatement).Times(Int(maxDepth)).Emit().Select(selectType).Dedup().
			Has(String("type"), typeStatement),
	)
}
