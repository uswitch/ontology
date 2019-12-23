package entity

import (
	"github.com/uswitch/ontology/pkg/types"
)

var ID = types.ID("/entity")

type Entity struct {
	types.Any
}

func init() { types.RegisterType(Entity{}, ID.String(), "") }
