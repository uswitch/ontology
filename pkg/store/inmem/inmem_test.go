package inmem

import (
	"testing"

	"github.com/uswitch/ontology/pkg/store/storetest"
)

func TestConformance(t *testing.T) {
	storetest.Conformance(t, NewInMemoryStore)
}
