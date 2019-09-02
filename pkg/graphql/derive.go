package graphql

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/graphql-go/graphql"

	"github.com/uswitch/ontology/pkg/store"
)

func splitAndTitle(str, delim string) string {
	parts := strings.Split(str, delim)

	out := ""
	for _, part := range parts {

		out = out + strings.Title(part)
	}

	return out
}

func nameFromID(id string) string {
	titleDash := splitAndTitle(id, "/")
	titleUnderscore := splitAndTitle(titleDash, "_")

	return titleUnderscore
}

func fieldCase(id string) string {
	return strings.ToLower(string(id[0])) + id[1:]
}

func plural(str string) string {
	lastCharacter := string(str[len(str) - 1])

	if lastCharacter == "y" {
		return str[0:len(str)-1] + "ies"
	} else {
		return str + "s"
	}
}

func objectFromType(ctx context.Context, s store.Store, typ *store.Type) (*graphql.Object, error) {

	fields := graphql.Fields{
		"metadata": metadataField,
	}

	// if it's an entity type then we should attach the related field
	if isAnEntity, err := s.Inherits(ctx, typ, store.EntityType); err != nil {
		return nil, err
	} else if isAnEntity {
		fields["related"] = relatedThingField
	}

	for currType := typ;; {
		if spec, ok := currType.Properties["spec"].(map[string]interface{}); ok {
			for k, v := range spec {
				fieldSpec, ok := v.(map[string]interface{})
				if !ok {
					log.Printf("%v#%v has no spec", currType.Metadata.ID, k)
					continue
				}

				var gqlTyp graphql.Type

				if typeString, ok := fieldSpec["type"].(string); !ok {
					return nil, fmt.Errorf("not type in field spec: %v", fieldSpec)
				} else {
					switch typeString {
					case "string":
						gqlTyp = graphql.String
					case "boolean":
						gqlTyp = graphql.Boolean
					case "number":
						gqlTyp = graphql.Float
					case "integer":
						gqlTyp = graphql.Int
					case "object":
						log.Printf("%v#%v is an object, which is currently unsupported", currType.Metadata.ID, k)
						continue
					case "array":
						log.Printf("%v#%v is an array, which is currently unsupported", currType.Metadata.ID, k)
						continue
					default:
						return nil, fmt.Errorf("%v#%v is of unknown type '%v'", currType.Metadata.ID, k, typeString)
					}
				}

				// the leaf most types field should take precedence
				// types are ordered with leaf at 0
				if _, ok := fields[k]; !ok {
					propKey := k
					fields[k] = &graphql.Field{
						Type: gqlTyp,
						Resolve: func(p graphql.ResolveParams) (interface{}, error) {
							thingable, ok := p.Source.(store.Thingable)
							if !ok {
								return nil, fmt.Errorf("Not thingable")
							}

							thing := thingable.Thing()

							return thing.Properties[propKey], nil
						},
					}
				}
			}
		}

		if nextTypeID, ok := currType.Properties["parent"]; !ok {
			break
		} else {
			nextType, err := s.GetTypeByID(ctx, store.ID(nextTypeID.(string)))
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
			thingInterface,
		},
	})

	return obj, nil
}
