package v1

import (
	"github.com/uswitch/ontology/pkg/types"
	"github.com/uswitch/ontology/pkg/types/relation"
)

type IsTheSameAs relation.Relation

func init() { types.RegisterType(IsTheSameAs{}, "/relation/v1/is_the_same_as", "/relation") }

type IsPartOf relation.Relation

func init() { types.RegisterType(IsPartOf{}, "/relation/v1/is_part_of", "/relation") }

type WasBuiltBy struct {
	relation.Relation
	Properties struct {
		relation.RelationProperties
		Ref *string `json:"ref"`
		At  string  `json:"at"`
	} `json:"properties"`
}

func init() { types.RegisterType(WasBuiltBy{}, "/relation/v1/was_built_by", "/relation") }
