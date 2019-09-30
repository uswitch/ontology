package graphql

import (
	"testing"

	"github.com/uswitch/ontology/pkg/store/inmem"
)

func TestProviderSyncOnce(t *testing.T) {
	s := inmem.NewInMemoryStore()
	p, err := NewProvider(s)
	if err != nil {
		t.Fatalf("Couldn't create provider: %v", err)
	}

	schema, _, err := p.Schema()
	if err != nil {
		t.Fatalf("Couldn't generate schema: %v", err)
	}

	types := schema.TypeMap()

	expectedTypes := []string{
		"Thing",
		"Metadata",
	}

	for _, expectedType := range expectedTypes {
		foundType := false
		for typeName, _ := range types {
			if typeName == expectedType {
				foundType = true
				break
			}
		}

		if !foundType {
			t.Errorf("Couldn't find expected type '%s' in types: %v", expectedType, types)
		}
	}
}
