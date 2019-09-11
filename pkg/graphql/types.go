package graphql

import (
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/graphql-go/graphql"

	"github.com/uswitch/ontology/pkg/store"
)

func resolveThingType(p graphql.ResolveTypeParams) *graphql.Object {
	thing, ok := p.Value.(*store.Thing)
	if !ok {
		log.Printf("wasn't a *store.Thing, was a %s", reflect.TypeOf(p.Value))
		return nil
	}

	provider, ok := p.Context.Value(ProviderContextKey).(*provider)
	if !ok {
		log.Println("Couldn't get a provider instance from the context")
	}

	if typ, ok := provider.TypeFor(thing.Metadata.Type); !ok {
		log.Printf("Couldn't get type for '%v' from provider: %v", thing.Metadata.Type, provider)
	} else {
		log.Printf("type of '%v': %v", thing.Metadata.Type, typ)
		return typ.(*graphql.Object)
	}

	return nil
}

var (
	metadataType = graphql.NewObject(graphql.ObjectConfig{
		Name: "Metadata",
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

					return metadata.UpdatedAt.Format(time.RFC3339), nil
				},
			},
		},
	})

	metadataField = &graphql.Field{
		Type: metadataType,
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			thingable, ok := p.Source.(store.Thingable)
			if !ok {
				return nil, fmt.Errorf("Not thingable")
			}

			thing := thingable.Thing()

			return thing.Metadata, nil
		},
	}

	thingInterface = graphql.NewInterface(graphql.InterfaceConfig{
		Name: "Thing",
		Fields: graphql.Fields{
			"metadata": metadataField,
		},
		ResolveType: resolveThingType,
	})
)

type relatedThing struct {
	relation *store.Relation
	entity    *store.Entity
}

var (
	entityInterface, relationInterface, typeInterface *graphql.Interface

	relatedThingType *graphql.Object

	relatedThingField, relationsField, typeField, typedThingsField, aField, bField *graphql.Field
)

func init() {
	typeInterface = graphql.NewInterface(graphql.InterfaceConfig{
		Name: "IType",
		Fields: graphql.Fields{
			"metadata": metadataField,
			"things":   typedThingsField,
		},
		ResolveType: resolveThingType,
	})

	typeField = &graphql.Field{
		Type: typeInterface,
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			s, ok := p.Context.Value(StoreContextKey).(store.Store)
			if !ok {
				log.Println("Couldn't get a store instance from the context")
			}

			thing, ok := p.Source.(*store.Thing)
			if !ok {
				return nil, fmt.Errorf("Not an thing: %v", p.Source)
			}

			typ, err := s.GetTypeByID(p.Context, thing.Metadata.Type)

			return (interface{})(typ.Thing()), err
		},
	}

	entityInterface = graphql.NewInterface(graphql.InterfaceConfig{
		Name: "IEntity",
		Fields: graphql.Fields{
			"metadata":  metadataField,
			"type":      typeField,
		},

		ResolveType: resolveThingType,
	})

	aField = &graphql.Field{
		Type: entityInterface,
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			s, ok := p.Context.Value(StoreContextKey).(store.Store)
			if !ok {
				log.Println("Couldn't get a store instance from the context")
			}

			thing, ok := p.Source.(*store.Thing)
			if !ok {
				return nil, fmt.Errorf("Not an thing: %v", p.Source)
			}

			id := store.ID(thing.Properties["a"].(string))
			ent, err := s.GetEntityByID(p.Context, id)
			if err == store.ErrNotFound {
				ent = &store.Entity{
					Metadata: store.Metadata{
						ID: id,
					},
				}
			} else if err != nil {
				return nil, err
			}

			return (interface{})(ent.Thing()), nil
		},
	}

	bField = &graphql.Field{
		Type: entityInterface,
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			s, ok := p.Context.Value(StoreContextKey).(store.Store)
			if !ok {
				log.Println("Couldn't get a store instance from the context")
			}

			thing, ok := p.Source.(*store.Thing)
			if !ok {
				return nil, fmt.Errorf("Not an thing: %v", p.Source)
			}

			id := store.ID(thing.Properties["b"].(string))
			ent, err := s.GetEntityByID(p.Context, id)
			if err != nil {
				return nil, err
			}

			return (interface{})(ent.Thing()), err
		},
	}

	relationInterface = graphql.NewInterface(graphql.InterfaceConfig{
		Name: "IRelation",
		Fields: graphql.Fields{
			"metadata": metadataField,
			"type":     typeField,
			"a": aField,
			"b": bField,
		},
		ResolveType: resolveThingType,
	})

	typedThingsField = &graphql.Field{
		Type: NewPaginatedListWithName(thingInterface, "TypedThingPage"),
		Args: PageArgs,
		Resolve: ResolvePage(func(listOptions store.ListOptions, p graphql.ResolveParams) (interface{}, error) {
			s, ok := p.Context.Value(StoreContextKey).(store.Store)
			if !ok {
				log.Println("Couldn't get a store instance from the context")
			}

			thing, ok := p.Source.(*store.Thing)
			if !ok {
				return nil, fmt.Errorf("Not an thing: %v", p.Source)
			}

			typ := (*store.Type)(thing)

			return s.ListByType(p.Context, typ, listOptions)
		}),
	}

	relatedThingType = graphql.NewObject(graphql.ObjectConfig{
		Name: "RelatedThing",
		Fields: graphql.Fields{
			"metadata": &graphql.Field{
				Type:        metadataType,
				Description: "Metadata for this thing",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					relEnt, ok := p.Source.(*relatedThing)
					if !ok {
						return nil, fmt.Errorf("Not a relatedThing: %v", p.Source)
					}

					return relEnt.relation.Metadata, nil
				},
			},
			"entity": &graphql.Field{
				Type: entityInterface,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					relEnt, ok := p.Source.(*relatedThing)
					if !ok {
						return nil, fmt.Errorf("Not a relatedThing: %v", p.Source)
					}

					return relEnt.entity.Thing(), nil
				},
			},
		},
	})

	relatedThingField = &graphql.Field{
		Type: NewPaginatedList(relatedThingType),
		Args: PageArgsWith(graphql.FieldConfigArgument{
			"type": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
		}),
		Resolve: ResolvePage(func(listOptions store.ListOptions, p graphql.ResolveParams) (interface{}, error) {
			s, ok := p.Context.Value(StoreContextKey).(store.Store)
			if !ok {
				log.Println("Couldn't get a store instance from the context")
			}

			thing, ok := p.Source.(*store.Thing)
			if !ok {
				return nil, fmt.Errorf("Not an thing: %v", p.Source)
			}

			entity := (*store.Entity)(thing)

			relTypeID, relTypeOk := p.Args["type"].(string)
			relType := store.RelationType

			if relTypeOk && relTypeID != "" {
				if t, err := s.GetTypeByID(p.Context, store.ID(relTypeID)); err != nil {
					return nil, fmt.Errorf("Can't get type to filter: %v", err)
				} else {
					relType = t
				}
			}

			var relations []*store.Relation
			var err error

			relations, err = s.ListRelationsForEntity(p.Context, relType, entity, listOptions)
			if err != nil {
				return nil, fmt.Errorf("Couldn't list relations for entity: %v", err)
			}

			relatedThings := []*relatedThing{}

			for _, relation := range relations {
				otherID, err := relation.OtherID(entity)
				if err != nil {
					return nil, fmt.Errorf("Failed to resolve other id: %v", err)
				}

				otherEntity, err := s.GetEntityByID(p.Context, otherID)
				if err == store.ErrNotFound {
					continue
				} else if err != nil {
					return nil, fmt.Errorf("Failed to get other id: %v", err)
				}

				relatedThings = append(relatedThings, &relatedThing{
					relation: relation,
					entity:    otherEntity,
				})
			}

			return (interface{})(relatedThings), nil
		}),
	}

	relationsField = &graphql.Field{
		Type: NewPaginatedList(relationInterface),
		Args: PageArgsWith(graphql.FieldConfigArgument{
			"type": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
		}),
		Resolve: ResolvePage(func(listOptions store.ListOptions, p graphql.ResolveParams) (interface{}, error) {
			s, ok := p.Context.Value(StoreContextKey).(store.Store)
			if !ok {
				log.Println("Couldn't get a store instance from the context")
			}

			thing, ok := p.Source.(*store.Thing)
			if !ok {
				return nil, fmt.Errorf("Not an thing: %v", p.Source)
			}

			entity := (*store.Entity)(thing)

			relTypeID, relTypeOk := p.Args["type"].(string)
			relType := store.RelationType

			if relTypeOk && relTypeID != "" {
				if t, err := s.GetTypeByID(p.Context, store.ID(relTypeID)); err != nil {
					return nil, fmt.Errorf("Can't get type to filter: %v", err)
				} else {
					relType = t
				}
			}

			relations, err := s.ListRelationsForEntity(p.Context, relType, entity, listOptions)
			if err != nil {
				return nil, fmt.Errorf("Couldn't list relations for entity: %v", err)
			}

			things := make([]*store.Thing, len(relations))
			for idx, relation := range relations {
				things[idx] = (*store.Thing)(relation)
			}

			return (interface{})(things), nil
		}),
	}

	entityInterface.AddFieldConfig("relations", relationsField)
	entityInterface.AddFieldConfig("related", relatedThingField)

}
