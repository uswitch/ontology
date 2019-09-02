package graphql

import (
	"fmt"
	"log"
	"time"

	"github.com/graphql-go/graphql"

	"github.com/uswitch/ontology/pkg/store"
)

var (

	metadataType = graphql.NewObject(graphql.ObjectConfig{
		Name:        "Metadata",
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
		Type:        metadataType,
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
		Name:        "Thing",
		Fields: graphql.Fields{
			"metadata": metadataField,
		},
		ResolveType: func(p graphql.ResolveTypeParams) *graphql.Object {
			thing, ok := p.Value.(*store.Thing)
			if !ok {
				log.Println("wasn't a thing")
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

		},
	})
)



type relatedThing struct {
	relation *store.Relation
	thing    *store.Thing
}


var (

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
			"thing": &graphql.Field{
				Type: thingInterface,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					relEnt, ok := p.Source.(*relatedThing)
					if !ok {
						return nil, fmt.Errorf("Not a relatedThing: %v", p.Source)
					}

					return relEnt.thing, nil
				},
			},
		},
	})

	relatedThingPageType = NewPaginatedList(relatedThingType)

	relatedThingField = &graphql.Field{
		Type: relatedThingPageType,
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
				return nil, err
			}

			relatedThings := make([]*relatedThing, len(relations))

			for idx, relation := range relations {
				otherID, err := relation.OtherID(entity)
				if err != nil {
					return nil, err
				}

				otherEntity, err := s.GetEntityByID(p.Context, otherID)
				if err != nil {
					return nil, err
				}

				relatedThings[idx] = &relatedThing{
					relation: relation,
					thing:   otherEntity.Thing(),
				}
			}

			return (interface{})(relatedThings), nil
		}),
	}

)
