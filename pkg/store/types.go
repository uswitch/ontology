package store

import (
	"github.com/uswitch/ontology/pkg/types"
)

type Store interface {
	Add(types.Instance) error
	Get(types.ID) (types.Instance, error)
	ListByType(types.ID) ([]types.Instance, error)
}
