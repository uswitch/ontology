package entity

import (
	"github.com/uswitch/ontology/pkg/types"
)

type Entity struct {
	types.Any
}

func init() { types.RegisterType(Entity{}, "/entity", "/any") }
