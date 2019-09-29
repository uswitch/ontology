package graphql

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/graphql-go/graphql"

	"github.com/uswitch/ontology/pkg/store"
)

const (
	StoreContextKey    = "graphql-store"
	ProviderContextKey = "graphql-provider"
)

type TypePair struct {
	GraphQL graphql.Type
	Store *store.Type
}

type provider struct {
	s store.Store

	types map[string]TypePair
	rw    sync.RWMutex
}

func NewProvider(s store.Store) (*provider, error) {
	p := &provider{
		s: s,
		types: map[string]TypePair{
			"/": TypePair{thingInterface, nil},
		},
	}

	return p, nil
}

func (p *provider) AddValuesTo(ctx context.Context) context.Context {
	storeCtx := context.WithValue(ctx, StoreContextKey, p.s)
	storeAndProviderCtx := context.WithValue(storeCtx, ProviderContextKey, p)

	return storeAndProviderCtx
}

// list all the types and then start streaming changes
// so we can keep an up to date set of types in the schema
// it will stop streaming when the context is done
func (p *provider) Sync(ctx context.Context) error {
	return p.SyncOnce(ctx)
}

func (p *provider) AddType(ctx context.Context, types ...*store.Type) error {
	return p.AddTypes(ctx, types)
}

func (p *provider) AddTypes(ctx context.Context, types []*store.Type) error {
	typeMap := map[string]TypePair{}

	for _, typ := range types {
		graphqlType, err := objectFromType(ctx, p.s, typ)
		if err != nil {
			log.Printf("failed to generate type %s: %v", typ, err)
			return err
		}

		typeMap[typ.Metadata.ID.String()] = TypePair{GraphQL: graphqlType, Store: typ}
	}

	p.rw.Lock()
	defer p.rw.Unlock()

	for k, pair := range typeMap {
		p.types[k] = pair
	}

	return nil
}

// only looks at the types at a point in time, rather than
// setting up a streaming sync
func (p *provider) SyncOnce(ctx context.Context) error {
	types, err := pageAllTypes(ctx, p.s.ListTypes)
	if err != nil {
		return err
	}

	return p.AddTypes(ctx, types)
}

func (p *provider) TypeFor(id store.ID) (graphql.Type, bool) {
	p.rw.RLock()
	defer p.rw.RUnlock()

	pair, ok := p.types[id.String()]

	return pair.GraphQL, ok
}

func (p *provider) Types() []graphql.Type {
	p.rw.RLock()
	defer p.rw.RUnlock()

	types := make([]graphql.Type, len(p.types))
	idx := 0
	for _, pair := range p.types {
		types[idx] = pair.GraphQL
		idx = idx + 1
	}

	return types
}

func (p *provider) TypePairs() []TypePair {
	p.rw.RLock()
	defer p.rw.RUnlock()

	pairs := make([]TypePair, len(p.types))
	idx := 0
	for _, pair := range p.types {
		pairs[idx] = pair
		idx = idx + 1
	}

	return pairs
}

// generate a schema for the currently known types, this
// will change over time as the types in the store change
func (p *provider) Schema() (graphql.Schema, error) {
	types := p.Types()
	pairs := p.TypePairs()
	fields := graphql.Fields{}

	for _, pair := range pairs {
		typ := pair.GraphQL
		styp := pair.Store

		name := typ.Name()

		single := fieldCase(name)
		plural := plural(single)

		fields[single] = &graphql.Field{
			Type: typ,
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type: graphql.ID,
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				s, ok := p.Context.Value(StoreContextKey).(store.Store)
				if !ok {
					log.Println("Couldn't get a store instance from the context")
				}

				return s.GetByID(p.Context, store.ID(p.Args["id"].(string)))
			},
		}

		fields[plural] = &graphql.Field{
			Type: NewPaginatedList(typ),
			Args: PageArgs,
			Resolve: ResolvePage(func(listOptions store.ListOptions, p graphql.ResolveParams) (interface{}, error) {
				s, ok := p.Context.Value(StoreContextKey).(store.Store)
				if !ok {
					log.Println("Couldn't get a store instance from the context")
				}

				things, err := s.ListByType(p.Context, styp, listOptions)

				return (interface{})(things), err
			}),
		}
	}

	rootQuery := graphql.NewObject(graphql.ObjectConfig{
		Name:   "Query",
		Fields: fields,
	})

	return graphql.NewSchema(graphql.SchemaConfig{
		Query: rootQuery,
		Types: types,
	})
}




func pageAllTypes(ctx context.Context, fn func(context.Context, store.ListOptions)([]*store.Type, error)) ([]*store.Type, error) {
	numResults := uint(10)
	offset := uint(0)
	currentOpts := store.ListOptions{
		NumberOfResults: numResults,
		Offset: offset,
	}

	allTypes := []*store.Type{}

	for {
		types, err := fn(ctx, currentOpts)
		if err != nil {
			return nil, err
		}

		allTypes = append(allTypes, types...)

		if len(types) == int(numResults) {
			currentOpts.Offset = currentOpts.Offset + numResults
		} else {
			break
		}
	}

	return allTypes, nil
}
