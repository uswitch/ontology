package graphql

import (
	"fmt"
	"log"

	"github.com/graphql-go/graphql"

	"github.com/uswitch/ontology/pkg/store"
)

func NewSchema(s store.Store) (*graphql.Schema, error) {
	var entityType, relationType, typeType *graphql.Object

	metadataType := graphql.NewObject(graphql.ObjectConfig{
		Name:        "Metadata",
		Description: "Metadata about any thing",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type:        graphql.ID,
				Description: "ID of the thing",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					metadata, ok := p.Source.(store.Metadata)
					if !ok {
						return nil, fmt.Errorf("Not metadata")
					}

					return metadata.ID, nil
				},
			},
			"type": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "Type of the thing",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					metadata, ok := p.Source.(store.Metadata)
					if !ok {
						return nil, fmt.Errorf("Not metadata")
					}

					return metadata.Type, nil
				},
			},
			"name": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "Name of the thing",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					metadata, ok := p.Source.(store.Metadata)
					if !ok {
						return nil, fmt.Errorf("Not metadata")
					}

					return metadata.Name, nil
				},
			},
			"updated_at": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "RFC3339 timestamp of when the thing was last updated",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					metadata, ok := p.Source.(store.Metadata)
					if !ok {
						return nil, fmt.Errorf("Not metadata")
					}

					return metadata.UpdatedAt.String(), nil
				},
			},
		},
	})

	metadataField := &graphql.Field{
		Type:        metadataType,
		Description: "Metadata for this thing",
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			thingable, ok := p.Source.(store.Thingable)
			if !ok {
				return nil, fmt.Errorf("Not thingable")
			}

			thing := thingable.Thing()

			return thing.Metadata, nil
		},
	}

	thingInterface := graphql.NewInterface(graphql.InterfaceConfig{
		Name:        "Thing",
		Description: "A thing",
		Fields: graphql.Fields{
			"metadata": metadataField,
		},
		ResolveType: func(p graphql.ResolveTypeParams) *graphql.Object {
			thing, ok := p.Value.(*store.Thing)
			if !ok {
				log.Println("wasn't a thing")
				return nil
			}

			if match, _ := s.IsA(thing, store.EntityType); match {
				return entityType
			} else if match, _ := s.IsA(thing, store.RelationType); match {
				return relationType
			} else if match, _ := s.IsA(thing, store.TypeType); match {
				return typeType
			} else {
				log.Printf("unknown type: %v", thing)
			}

			return nil

		},
	})

	typeType = graphql.NewObject(graphql.ObjectConfig{
		Name:        "Type",
		Description: "A type",
		Fields: graphql.Fields{
			"metadata": metadataField,
		},
		Interfaces: []*graphql.Interface{
			thingInterface,
		},
	})

	relationType = graphql.NewObject(graphql.ObjectConfig{
		Name:        "Relation",
		Description: "A relation",
		Fields: graphql.Fields{
			"metadata": metadataField,
		},
		Interfaces: []*graphql.Interface{
			thingInterface,
		},
	})

	type relatedEntity struct {
		relation *store.Relation
		entity   *store.Entity
	}

	entityType = graphql.NewObject(graphql.ObjectConfig{
		Name:        "Entity",
		Description: "An entity",
		Fields: graphql.Fields{
			"metadata": metadataField,
		},
		Interfaces: []*graphql.Interface{
			thingInterface,
		},
	})

	relatedEntityType := graphql.NewObject(graphql.ObjectConfig{
		Name: "RelatedEntity",
		Fields: graphql.Fields{
			"metadata": &graphql.Field{
				Type:        metadataType,
				Description: "Metadata for this thing",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					relEnt, ok := p.Source.(*relatedEntity)
					if !ok {
						return nil, fmt.Errorf("Not a relatedEntity: %v", p.Source)
					}

					return relEnt.relation.Metadata, nil
				},
			},
			"entity": &graphql.Field{
				Type: entityType,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					relEnt, ok := p.Source.(*relatedEntity)
					if !ok {
						return nil, fmt.Errorf("Not a relatedEntity: %v", p.Source)
					}

					return relEnt.entity, nil
				},
			},
		},
	})

	relatedEntityPageType := NewPaginatedList(relatedEntityType)

	entityType.AddFieldConfig("relatedEntities", &graphql.Field{
		Type: relatedEntityPageType,
		Args: PageArgs,
		Resolve: ResolvePage(func(listOptions store.ListOptions, p graphql.ResolveParams) (interface{}, error) {
			thing, ok := p.Source.(*store.Thing)
			if !ok {
				return nil, fmt.Errorf("Not an thing: %v", p.Source)
			}

			entity := (*store.Entity)(thing)

			var relations []*store.Relation
			var err error

			relations, err = s.ListRelationsForEntity(entity, listOptions)
			if err != nil {
				return nil, err
			}

			relatedEntities := make([]*relatedEntity, len(relations))

			for idx, relation := range relations {
				otherID, err := relation.OtherID(entity)
				if err != nil {
					return nil, err
				}

				otherEntity, err := s.GetEntityByID(otherID)
				if err != nil {
					return nil, err
				}

				relatedEntities[idx] = &relatedEntity{
					relation: relation,
					entity:   otherEntity,
				}
			}

			return (interface{})(relatedEntities), nil
		}),
	})

	thingPageType := NewPaginatedList(thingInterface)

	rootQuery := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"thing": &graphql.Field{
				Type: thingInterface,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.ID,
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return s.GetByID(store.ID(p.Args["id"].(string)))
				},
			},
			"things": &graphql.Field{
				Type: thingPageType,
				Args: PageArgsWith(graphql.FieldConfigArgument{
					"type": &graphql.ArgumentConfig{
						Type: graphql.ID,
					},
				}),
				Resolve: ResolvePage(func(listOptions store.ListOptions, p graphql.ResolveParams) (interface{}, error) {
					var things []*store.Thing
					var err error

					typeID, ok := p.Args["type"].(string)

					if !ok {
						things, err = s.List(listOptions)
					} else {
						typ, err := s.GetTypeByID(store.ID(typeID))
						if err != nil {
							return nil, err
						}

						things, err = s.ListByType(typ, listOptions)
					}

					return (interface{})(things), err
				}),
			},
		},
	})

	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: rootQuery,
	})

	if err != nil {
		return nil, err
	}

	additionalTypes := []*graphql.Object{entityType, relationType, typeType}

	for _, additionalType := range additionalTypes {
		if err := schema.AppendType(additionalType); err != nil {
			return nil, err
		}
	}

	return &schema, nil
}
