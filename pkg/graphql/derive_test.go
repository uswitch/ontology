package graphql

import (
	"context"
	"testing"

	"github.com/graphql-go/graphql"

	"github.com/uswitch/ontology/pkg/store/inmem"
	"github.com/uswitch/ontology/pkg/store/storetest"
)

func TestDeriveName(t *testing.T) {
	tests := []struct {
		In  string
		Out string
	}{
		{In: "/wibble/bibble", Out: "WibbleBibble"},
		{In: "/wibble/bibble_nibble", Out: "WibbleBibbleNibble"},
	}

	for _, test := range tests {
		actual := nameFromID(test.In)
		if actual != test.Out {
			t.Errorf("expected output to be '%s', but it was '%s'", test.Out, actual)
		}
	}
}

func TestDeriveFieldCase(t *testing.T) {
	tests := []struct {
		In  string
		Out string
	}{
		{In: "WibbleBibble", Out: "wibbleBibble"},
	}

	for _, test := range tests {
		actual := fieldCase(test.In)
		if actual != test.Out {
			t.Errorf("expected output to be '%s', but it was '%s'", test.Out, actual)
		}
	}
}

func TestDerivePlural(t *testing.T) {
	tests := []struct {
		In  string
		Out string
	}{
		{In: "relation", Out: "relations"},
		{In: "entity", Out: "entities"},
	}

	for _, test := range tests {
		actual := plural(test.In)
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
		"something_else": map[string]interface{}{
			"type": "number",
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

	expectedFields := map[string]graphql.Type{
		"metadata":       metadataType,
		"data":           graphql.String,
		"something_else": graphql.Float,
	}
	fields := graphqlObject.Fields()

	for expectedField, expectedType := range expectedFields {
		foundField := false
		for field, def := range fields {
			if field == expectedField {
				foundField = true
				if def.Type != expectedType {
					t.Errorf("types did not match: %v != %v", def.Type, expectedType)
				}
				break
			}

		}
		if !foundField {
			t.Errorf("Didn't find field %s in %v", expectedField, fields)
		}
	}
}
