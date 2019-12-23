package storetest

import (
	"context"
	"testing"

	"github.com/uswitch/ontology/pkg/store"
	"github.com/uswitch/ontology/pkg/types"
	"github.com/uswitch/ontology/pkg/types/entity"
	"github.com/uswitch/ontology/pkg/types/relation"
)

func Conformance(t *testing.T, newStore func() store.Store) {
	tests := map[string]func(*testing.T, store.Store){
		"AddAndGet": TestAddAndGet,
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			s := newStore()
			test(t, s)
		})
	}
}

func TestAddAndGet(t *testing.T, s store.Store) {
	entityID := types.ID("/wibble")
	relationID := types.ID("/wibble+bibble")

	if err := s.Add(
		context.Background(),
		&entity.Entity{
			types.Any{Metadata: types.Metadata{ID: entityID, Type: entity.ID}},
		},
		&entity.Entity{
			types.Any{Metadata: types.Metadata{ID: "/bibble", Type: entity.ID}},
		},
		&relation.Relation{
			types.Any{Metadata: types.Metadata{ID: relationID, Type: relation.ID}},
			relation.Properties{A: entityID, B: "/bibble"},
		},
	); err != nil {
		t.Fatalf("failed to add entity: %v", err)
	}

	if inst, err := s.Get(context.Background(), entityID); err != nil {
		t.Fatalf("failed to get entity: %v", err)
	} else if inst == nil {
		t.Fatalf("no entity returned")
	} else if inst.ID() != entityID {
		t.Fatalf("entity ids don't match: %s != %s", inst.ID(), entityID)
	}

	if inst, err := s.Get(context.Background(), relationID); err != nil {
		t.Fatalf("failed to get relation: %v", err)
	} else if inst == nil {
		t.Fatalf("no relation returned")
	} else if inst.ID() != relationID {
		t.Fatalf("relation ids don't match: %s != %s", inst.ID(), entityID)
	}
}
