package store

import (
	"context"
	"testing"
)

func TestTypeProperties(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()

	relType := typ("/relation/wibble", "/relation", map[string]interface{}{
		"a": map[string]interface{}{
			"type": "string",
			"pointer_to": "/entity/thing",
		},
		"bibble": map[string]interface{}{
			"type": "object",
		},
	})

	if err := store.Add(ctx, relType.Thing()); err != nil {
		t.Fatalf("Couldn't add to store: %v", err)
	}

	props, requiredProps, err := typeProperties(ctx, store, relType)
	if err != nil {
		t.Fatalf("Couldn't get the types properties: %v", err)
	}

	requiredPropsArr := ([]string)(requiredProps)

	expectedRequiredProps := map[string]bool{"a": true, "b": true}
	if len(requiredPropsArr) != len(expectedRequiredProps) {
		t.Errorf("Expected there to be %d required props, there were %d", len(expectedRequiredProps), len(requiredPropsArr))
	}

	for _, requiredProp := range requiredPropsArr {
		if _, ok := expectedRequiredProps[requiredProp]; !ok {
			t.Errorf("Expected '%v' to be required", requiredProp)
		}
	}

	expectedKeys := []string{"a", "b", "bibble"}
	for _, expectedKey := range expectedKeys {
		if _, ok := props[expectedKey]; !ok {
			t.Errorf("Didn't find key '%s' in %v", expectedKey, props)
		}
	}

	if pointerTo := props["a"].Validators["pointer_to"].(*PointerTo); pointerTo.String() != "/entity/thing" {
		t.Errorf("Expected pointer_to for a to be '/entity/thing', but it was '%s'", pointerTo.String())
	}
	if pointerTo := props["b"].Validators["pointer_to"].(*PointerTo); pointerTo.String() != "/entity" {
		t.Errorf("Expected pointer_to for b to be '/entity', but it was '%s'", pointerTo.String())
	}
}

func TestValidate(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()

	relType := typ("/relation/wibble", "/relation", map[string]interface{}{
		"a": map[string]interface{}{
			"type": "string",
			"pointer_to": "/entity/thing",
		},
		"bibble": map[string]interface{}{
			"type": "object",
		},
	})

	entThingType := typ("/entity/thing", "/entity", map[string]interface{}{})

	ent1 := entityWithType("/asdf", "/entity")
	ent2 := entityWithType("/sdfg", "/entity/thing")

	if err := store.Add(ctx, entThingType.Thing(), relType.Thing(), ent1, ent2); err != nil {
		t.Fatalf("Couldn't add to store: %v", err)
	}

	validRel := relationBetweenWithType("/qwer", "/relation/wibble", "/sdfg", "/asdf")
	if valErrs, err := store.Validate(ctx, validRel, ValidateOptions{}); err != nil {
		t.Errorf("Failed to validate thing: %v", err)
	} else if len(valErrs) != 0 {
		t.Errorf("Expected 0 validation errors, got %d: %v", len(valErrs), valErrs)
	}

	wrongwayroundRel := relationBetweenWithType("/qwer", "/relation/wibble", "/asdf", "/sdfg")
	if valErrs, err := store.Validate(ctx, wrongwayroundRel, ValidateOptions{}); err != nil {
		t.Errorf("Failed to validate thing: %v", err)
	} else if len(valErrs) != 1 {
		t.Errorf("Expected 1 validation error, got %d: %v", len(valErrs), valErrs)
	}
	if valErrs, err := store.Validate(ctx, wrongwayroundRel, ValidateOptions{Pointers: IgnoreMissingPointers}); err != nil {
		t.Errorf("Failed to validate thing: %v", err)
	} else if len(valErrs) != 1 {
		t.Errorf("Expected 1 validation error, got %d: %v", len(valErrs), valErrs)
	}
	if valErrs, err := store.Validate(ctx, wrongwayroundRel, ValidateOptions{Pointers: IgnoreAllPointers}); err != nil {
		t.Errorf("Failed to validate thing: %v", err)
	} else if len(valErrs) != 0 {
		t.Errorf("Expected 0 validation errors, got %d: %v", len(valErrs), valErrs)
	}

	invalidRel := thingWithType("/wert", "/relation/wibble", Properties{})
	if valErrs, err := store.Validate(ctx, invalidRel, ValidateOptions{}); err != nil {
		t.Errorf("Failed to validate thing: %v", err)
	} else if len(valErrs) == 0 {
		t.Errorf("Expected more than 0 validation errors, got %d: %v", len(valErrs), valErrs)
	}

	missingRel := relationBetweenWithType("/qwer", "/relation/wibble", "/unknown-entity", "/asdf")
	if valErrs, err := store.Validate(ctx, missingRel, ValidateOptions{}); err != nil {
		t.Errorf("Failed to validate thing: %v", err)
	} else if len(valErrs) != 1 {
		t.Errorf("Expected 1 validation error, got %d: %v", len(valErrs), valErrs)
	}
	if valErrs, err := store.Validate(ctx, missingRel, ValidateOptions{Pointers: IgnoreMissingPointers}); err != nil {
		t.Errorf("Failed to validate thing: %v", err)
	} else if len(valErrs) != 0 {
		t.Errorf("Expected 0 validation errors, got %d: %v", len(valErrs), valErrs)
	}

}
