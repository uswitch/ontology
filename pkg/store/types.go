package store

import (
	"context"

	"github.com/uswitch/ontology/pkg/types"
)

type Store interface {
	Add(context.Context, ...types.Instance) error
	Get(context.Context, types.ID) (types.Instance, error)
	ListByType(context.Context, types.ID) ([]types.Instance, error)
}
