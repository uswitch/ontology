package gremlin

import (
	"testing"

	"github.com/uswitch/ontology/pkg/store"
	"github.com/uswitch/ontology/pkg/store/storetest"
)

func TestConformance(t *testing.T) {
	s, err := NewLocalServer("ws://localhost:8182")
	if err != nil {
		t.Fatalf("failed to setup local connection: %v", err)
	}

	storetest.Conformance(t, func() store.Store {
		ls := s.(*localStore)

		_, err := ls.client.Execute(
			Graph().E().Drop().Iterate().String(),
			map[string]string{},
			map[string]string{},
		)
		if err != nil {
			t.Fatalf("failed to drop all edges: %v", err)
		}
		_, err = ls.client.Execute(
			Graph().V().Drop().Iterate().String(),
			map[string]string{},
			map[string]string{},
		)
		if err != nil {
			t.Fatalf("failed to drop all vertices: %v", err)
		}

		return s
	})
}
