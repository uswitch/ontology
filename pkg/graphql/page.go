package graphql

import (
	"encoding/base64"
	"fmt"
	"math/bits"
	"strconv"

	"github.com/graphql-go/graphql"

	"github.com/uswitch/ontology/pkg/store"
)

type PageInfo struct {
	Cursor string
	Limit  int
}

type Page struct {
	PageInfo
	List interface{}
}

var (
	pageInfoType = graphql.NewObject(graphql.ObjectConfig{
		Name:        "PageInfo",
		Description: "Information about a page",
		Fields: graphql.Fields{
			"cursor": &graphql.Field{
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					pageInfo, ok := p.Source.(PageInfo)
					if !ok {
						return nil, fmt.Errorf("Not page info")
					}

					return pageInfo.Cursor, nil
				},
			},
			"limit": &graphql.Field{
				Type: graphql.Int,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					pageInfo, ok := p.Source.(PageInfo)
					if !ok {
						return nil, fmt.Errorf("Not page info")
					}

					return pageInfo.Limit, nil
				},
			},
		},
	})
)

func NewPaginatedList(typ graphql.Type) graphql.Type {
	return graphql.NewObject(graphql.ObjectConfig{
		Name: fmt.Sprintf("%sPage", typ.Name()),
		Fields: graphql.Fields{
			"list": &graphql.Field{
				Type: graphql.NewList(typ),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					page, ok := p.Source.(*Page)
					if !ok {
						return nil, fmt.Errorf("Not a Page: %v", p.Source)
					}

					return page.List, nil
				},
			},
			"page": &graphql.Field{
				Type: pageInfoType,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					page, ok := p.Source.(*Page)
					if !ok {
						return nil, fmt.Errorf("Not a Page: %v", p.Source)
					}

					return page.PageInfo, nil
				},
			},
		},
	})
}

type PageResolveFn func(store.ListOptions, graphql.ResolveParams) (interface{}, error)

func ResolvePage(resolveFn PageResolveFn) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
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

		newOffset := int(offset) + limit
		encodedCursor := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d", newOffset)))

		listOptions := store.ListOptions{
			SortOrder:       store.SortAscending,
			SortField:       store.SortByID,
			Offset:          offset,
			NumberOfResults: uint(limit),
		}

		list, err := resolveFn(listOptions, p)

		return &Page{
			PageInfo: PageInfo{
				Cursor: encodedCursor,
				Limit:  limit,
			},
			List: list,
		}, err
	}
}

func PageArgsWith(args graphql.FieldConfigArgument) graphql.FieldConfigArgument {
	args["limit"] = &graphql.ArgumentConfig{
		Type:         graphql.Int,
		DefaultValue: int(store.DefaultNumberOfResults),
	}
	args["cursor"] = &graphql.ArgumentConfig{
		Type: graphql.String,
	}

	return args
}

var PageArgs = PageArgsWith(graphql.FieldConfigArgument{})
