package relation

import (
	"github.com/uswitch/ontology/pkg/types"
)

var ID = types.ID("/relation")

type Instance interface {
	types.Instance
	A() types.ID
	B() types.ID
}

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

func (r *Relation) A() types.ID { return r.Properties.A }
func (r *Relation) B() types.ID { return r.Properties.B }

func init() { types.RegisterType(Relation{}, ID.String(), "/any") }
