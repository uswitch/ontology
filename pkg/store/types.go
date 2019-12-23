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
}
