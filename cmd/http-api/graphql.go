package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"math/bits"
	"strconv"

	"github.com/graphql-go/graphql"

	"github.com/uswitch/ontology/pkg/store"
)

func NewGraphQLSchema(s store.Store) (*graphql.Schema, error) {
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

	entityType = graphql.NewObject(graphql.ObjectConfig{
		Name:        "Entity",
		Description: "An entity",
		Fields: graphql.Fields{
			"metadata": metadataField,
			"relations": &graphql.Field{
				Type: graphql.NewList(relationType),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					thing, ok := p.Source.(*store.Thing)
					if !ok {
						return nil, fmt.Errorf("Not an thing: %v", p.Source)
					}

					entity := (*store.Entity)(thing)

					return s.ListRelationsForEntity(entity, store.ListOptions{})
				},
			},
		},
		Interfaces: []*graphql.Interface{
			thingInterface,
		},
	})

	type Page struct {
		Cursor string
		Limit int
	}

	pageType := graphql.NewObject(graphql.ObjectConfig{
		Name:        "Page",
		Description: "Information about a page",
		Fields: graphql.Fields{
			"cursor": &graphql.Field{
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					page, ok := p.Source.(Page)
					if !ok {
						return nil, fmt.Errorf("Not a page")
					}

					return page.Cursor, nil
				},
			},
			"limit": &graphql.Field{
				Type: graphql.Int,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					page, ok := p.Source.(Page)
					if !ok {
						return nil, fmt.Errorf("Not a page")
					}

					return page.Limit, nil
				},
			},
		},
	})

	type ThingPage struct {
		Page
		List []*store.Thing
	}

	thingPageType := graphql.NewObject(graphql.ObjectConfig{
		Name:        "ThingPage",
		Description: "A page of things",
		Fields: graphql.Fields{
			"list": &graphql.Field{
				Type: graphql.NewList(thingInterface),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					thingPage, ok := p.Source.(*ThingPage)
					if !ok {
						return nil, fmt.Errorf("Not a ThingPage: %v", p.Source)
					}

					return thingPage.List, nil
				},
			},
			"page": &graphql.Field{
				Type: pageType,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					thingPage, ok := p.Source.(*ThingPage)
					if !ok {
						return nil, fmt.Errorf("Not a ThingPage: %v", p.Source)
					}

					return thingPage.Page, nil
				},
			},
		},
	})

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
				Args: graphql.FieldConfigArgument{
					"type": &graphql.ArgumentConfig{
						Type: graphql.ID,
					},
					"limit": &graphql.ArgumentConfig{
						Type:         graphql.Int,
						DefaultValue: int(store.DefaultNumberOfResults),
					},
					"cursor": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					typeID, ok := p.Args["type"].(string)
					limit := p.Args["limit"].(int)
					cursor, cursorOk := p.Args["cursor"].(string)

					offset := uint(0)

					if cursorOk {
						decodedCursor, err := base64.StdEncoding.DecodeString(cursor)
						if err != nil {
							return nil, err
						}

						offset64, err := strconv.ParseUint(string(decodedCursor), 10, bits.UintSize)
						if err != nil {
							return nil, err
						}

						offset = uint(offset64)
					}

					listOptions := store.ListOptions{
						SortOrder:       store.SortAscending,
						SortField:       store.SortByID,
						Offset:          offset,
						NumberOfResults: uint(limit),
					}

					var things []*store.Thing
					var err error

					if !ok {
						things, err = s.List(listOptions)
					} else {
						typ, err := s.GetTypeByID(store.ID(typeID))
						if err != nil {
							return nil, err
						}

						things, err = s.ListByType(typ, listOptions)
					}

					if err != nil {
						return nil, err
					}

					newOffset := offset + uint(len(things))
					encodedCursor := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d", newOffset)))

					return &ThingPage{
						Page: Page{Cursor: encodedCursor, Limit: limit},
						List: things,
					}, nil
				},
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
