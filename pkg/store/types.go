package store

import (
	"context"
	"errors"

	"github.com/uswitch/ontology/pkg/types"
)

var (
	ErrNotFound = errors.New("not found")
)

type Store interface {
	Add(context.Context, ...types.Instance) error
	Get(context.Context, types.ID) (types.Instance, error)
	ListByType(context.Context, types.ID) ([]types.Instance, error)

	// Lists any instances from a root instance following all relations.
	// There is an assumption that the root instance will be an entity
	ListFromByType(context.Context, types.ID, types.ID) ([]types.Instance, error)
}
