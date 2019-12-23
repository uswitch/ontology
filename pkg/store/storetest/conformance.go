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
		"List":      TestList,
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

func TestList(t *testing.T, s store.Store) {
	if err := s.Add(
		context.Background(),
		&entity.Entity{
			types.Any{Metadata: types.Metadata{ID: "/wibble", Type: entity.ID}},
		},
		&entity.Entity{
			types.Any{Metadata: types.Metadata{ID: "/bibble", Type: entity.ID}},
		},
		&relation.Relation{
			types.Any{Metadata: types.Metadata{ID: "/wibble+bibble", Type: relation.ID}},
			relation.Properties{A: "/wibble", B: "/bibble"},
		},
	); err != nil {
		t.Fatalf("failed to add entity: %v", err)
	}

	if list, err := s.ListByType(context.Background(), entity.ID); err != nil {
		t.Errorf("failed to list entities: %v", err)
	} else if list == nil {
		t.Errorf("no entity list returned")
	} else if expected := 2; len(list) != expected {
		t.Errorf("entity list is the wrong size: %d != %d", len(list), expected)
	}

	if list, err := s.ListByType(context.Background(), relation.ID); err != nil {
		t.Errorf("failed to list relations: %v", err)
	} else if list == nil {
		t.Errorf("no relation list returned")
	} else if expected := 1; len(list) != expected {
		t.Errorf("relation list is the wrong size: %d != %d", len(list), expected)
	}
}
