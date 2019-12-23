package gremlin

import (
	"context"
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

func (l *local) execute(ctx context.Context, statement Statements) ([]interface{}, error) {
	//log.Println(statement.String())

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

	for _, instance := range instances {
		st := Var("g")

		if types.IsA(instance, entity.ID) {
			log.Printf("is an entity: %s", instance.ID())

			st = st.AddV(String(instance.ID())).
				Property(String("name"), String(instance.Name())).
				Property(String("updated_at"), String(instance.UpdatedAt().Format(DATE_LAYOUT)))
		} else if types.IsA(instance, relation.ID) {
			if instance, ok := instance.(relation.Instance); ok {
				log.Printf("is a relation: %s [%s -> %s]", instance.ID(), instance.A(), instance.B())

				st = st.AddE(String(instance.ID())).
					From(Keyword("g").V().HasLabel(String(instance.A()))).
					To(Keyword("g").V().HasLabel(String(instance.B()))).
					Property(String("name"), String(instance.Name())).
					Property(String("type"), String(instance.Type())).
					Property(String("updated_at"), String(instance.UpdatedAt().Format(DATE_LAYOUT)))
			} else {
				log.Printf("instance doesn't conform to relation.Instance: %T", instance)
			}

		} else {
			log.Printf("a rather strange type of instance: %s", instance.Type())
		}

		// we can only execute so many at a time. frame size is 65536.
		/*if i%1 == 0 {
			if _, err := l.execute(ctx, Statements{st}); err != nil {
				return err
			} else {
				st = Var("g")
			}
		}*/
		if _, err := l.execute(ctx, Statements{st}); err != nil {
			return err
		}
	}

	return nil
}
func (l *local) Get(ctx context.Context, id types.ID) (types.Instance, error) {
	return nil, nil
}
func (l *local) ListByType(ctx context.Context, id types.ID) ([]types.Instance, error) {
	return nil, nil
}
