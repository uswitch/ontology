package graphql

import (
	"context"
	"testing"

	"github.com/uswitch/ontology/pkg/store/inmem"
	"github.com/uswitch/ontology/pkg/store/storetest"
)

func TestDeriveName(t *testing.T) {
	tests := []struct {
		In  string
		Out string
	}{
		{In: "/wibble/bibble", Out: "WibbleBibble"},
	}

	for _, test := range tests {
		actual := nameFromID(test.In)
		if actual != test.Out {
			t.Errorf("expected output to be '%s', but it was '%s'", test.Out, actual)
		}
	}
}

func TestDeriveObjectFromType(t *testing.T) {
	s := inmem.NewInMemoryStore()
	ctx := context.Background()

	entityType := storetest.Type("/entity/wibble", "/entity", map[string]interface{}{
		"data": map[string]interface{}{
			"type": "string",
		},
	})

	s.Add(ctx, entityType)

	graphqlObject, err := objectFromType(ctx, s, entityType)
	if err != nil {
		t.Fatalf("Failed to make object: %v", err)
	}

	if expected := "EntityWibble"; graphqlObject.Name() != expected {
		t.Errorf("graphql object was not named correctly: %s != %s", graphqlObject.Name(), expected)
	}

	expectedInterfaces := []string{
		//	"Thing",  //we need to extract this from NewSchema before we can easily add it
	}
	interfaces := graphqlObject.Interfaces()

	for _, expectedInterface := range expectedInterfaces {
		foundInterface := false
		for _, iface := range interfaces {
			if iface.Name() == expectedInterface {
				foundInterface = true
				break
			}
		}
		if !foundInterface {
			t.Errorf("Didn't find interface %s in %v", expectedInterface, interfaces)
		}
	}

	expectedFields := []string{
		"metadata", "data",
	}
	fields := graphqlObject.Fields()

	for _, expectedField := range expectedFields {
		foundField := false
		for field, _ := range fields {
			if field == expectedField {
				foundField = true
				break
			}
		}
		if !foundField {
			t.Errorf("Didn't find field %s in %v", expectedField, fields)
		}
	}
}
