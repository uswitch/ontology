package graphql

import (
	"fmt"
	"strings"

	"github.com/graphql-go/graphql"

	"github.com/uswitch/ontology/pkg/store"
)

func nameFromID(id string) string {
	parts := strings.Split(id, "/")

	out := ""
	for _, part := range parts {
		out = out + strings.Title(part)
	}

	return out
}

func objectFromType(s store.Store, typ *store.Type) (*graphql.Object, error) {

	fields := graphql.Fields{
		"metadata": metadataField,
	}

	for currType := typ;; {
		if spec, ok := currType.Properties["spec"].(map[string]interface{}); ok {
			for k, _ := range spec {
				// the leaf most types field should take precedence
				// types are ordered with leaf at 0
				if _, ok := fields[k]; !ok {
					fields[k] = &graphql.Field{
						Type: graphql.String,
						Resolve: func(p graphql.ResolveParams) (interface{}, error) {
							thingable, ok := p.Source.(store.Thingable)
							if !ok {
								return nil, fmt.Errorf("Not thingable")
							}

							thing := thingable.Thing()

							return thing.Properties[k], nil
						},
					}
				}
			}
		}

		if nextTypeID, ok := currType.Properties["parent"]; !ok {
			break
		} else {
			nextType, err := s.GetTypeByID(store.ID(nextTypeID.(string)))
			if err != nil {
				return nil, err
			}

			currType = nextType
		}
	}

	obj := graphql.NewObject(graphql.ObjectConfig{
		Name: nameFromID(typ.Metadata.ID.String()),
		Fields: fields,
		Interfaces: []*graphql.Interface{
		},
	})

	return obj, nil
}
