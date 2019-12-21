package relation

import (
	"github.com/uswitch/ontology/pkg/types"
)

type RelationProperties struct {
	A types.ID `json:"a" ontology:"pointer,/entity"`
	B types.ID `json:"b" ontology:"pointer,/entity"`
}

type Relation struct {
	types.Any
	Properties struct {
		RelationProperties
	} `json:"properties"`
}

func init() { types.RegisterType(Relation{}, "/relation", "/any") }
