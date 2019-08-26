package store

import (
	"context"
	"sync"
	"testing"
)

func thingWithType(thingID string, typeID string, properties Properties) *Thing {
	thing := &Thing{
		Metadata: Metadata{
			ID:   ID(thingID),
			Type: ID(typeID),
		},
	}

	if properties != nil {
		thing.Properties = properties
	}

	return thing
}
func entity(ID string) *Thing   { return thingWithType(ID, "/entity", nil) }
func entityWithType(ID, typ string) *Thing   { return thingWithType(ID, typ, nil) }
func relation(ID string) *Thing { return thingWithType(ID, "/relation", nil) }
func ntype(ID string) *Thing    { return thingWithType(ID, "/type", nil) }
func typ(ID, parent string, spec map[string]interface{}) *Type {
	props := Properties{}

	if parent != "" {
		props["parent"] = parent
	}
	props["spec"] = spec

	return (*Type)(thingWithType(ID, "/type", props))
}
func relationBetween(ID, a, b string) *Thing {
	return relationBetweenWithType(ID, "/relation", a, b)
}
func relationBetweenWithType(ID, typ, a, b string) *Thing {
	return thingWithType(
		ID, typ,
		Properties{
			"a": a,
			"b": b,
		},
	)
}

func TestLen(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()

	if num, err := store.Len(ctx); err != nil {
		t.Error(err)
	} else if num != 3 {
		t.Errorf("Store should have 3 base types, has %d", num)
	}

	if err := store.Add(ctx, entity("/wibble")); err != nil {
		t.Fatal(err)
	}

	if num, err := store.Len(ctx); err != nil {
		t.Error(err)
	} else if num != 4 {
		t.Errorf("Store should have 4 entries, has %d", num)
	}
}

func TestIsA(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()

	if ok, err := store.IsA(ctx, EntityType.Thing(), TypeType); !ok {
		t.Error("EntityType should be a TypeType")
	} else if err != nil {
		t.Error(err)
	}

	if ok, err := store.IsA(ctx, RelationType.Thing(), TypeType); !ok {
		t.Error("RelationType should be a TypeType")
	} else if err != nil {
		t.Error(err)
	}

	if ok, err := store.IsA(ctx, TypeType.Thing(), TypeType); !ok {
		t.Error("TypeType should be a TypeType")
	} else if err != nil {
		t.Error(err)
	}

	if ok, err := store.IsA(ctx, entity("/wibble/ent").Thing(), EntityType); !ok {
		t.Error("An entity should be type EntityType")
	} else if err != nil {
		t.Error(err)
	}
	if ok, err := store.IsA(ctx, entity("/wibble/ent").Thing(), RelationType); ok {
		t.Error("An entity should not be type RelationType")
	} else if err != nil {
		t.Error(err)
	}

	if ok, err := store.IsA(ctx, relation("/wibble/rel").Thing(), RelationType); !ok {
		t.Error("A relation should be type RelationType")
	} else if err != nil {
		t.Error(err)
	}
	if ok, err := store.IsA(ctx, relation("/wibble/rel").Thing(), EntityType); ok {
		t.Error("A relation should not be type EntityType")
	} else if err != nil {
		t.Error(err)
	}
}

func TestAddAndGet(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()

	if err := store.Add(ctx, entity("/wibble/bibble")); err != nil {
		t.Fatalf("Couldn't add to store: %v", err)
	}

	if thing, err := store.GetByID(ctx, ID("/wibble/bibble")); err != nil {
		t.Error(err)
	} else if thing.Metadata.ID != "/wibble/bibble" {
		t.Errorf("thing had wrong ID: %s", thing.Metadata.ID)
	}

	if entity, err := store.GetEntityByID(ctx, ID("/wibble/bibble")); err != nil {
		t.Error(err)
	} else if entity.Metadata.ID != "/wibble/bibble" {
		t.Errorf("entity had wrong ID: %s", entity.Metadata.ID)
	}

	if _, err := store.GetRelationByID(ctx, ID("/wibble/bibble")); err == nil {
		t.Errorf("should not have been able to retrieve a relation, it's an entity")
	} else if err != ErrNotFound {
		t.Error(err)
	}

	if _, err := store.GetTypeByID(ctx, ID("/wibble/bibble")); err == nil {
		t.Errorf("should not have been able to retrieve a type, it's an entity")
	} else if err != ErrNotFound {
		t.Error(err)
	}
}

func TestGetCorrectType(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()

	if err := store.Add(ctx,
		entity("/wibble/bibble/1"),
		relation("/wibble/bibble/2"),
		ntype("/wibble/bibble/3"),
		thingWithType("/wibble/bibble/4", "/type/", nil),
	); err != nil {
		t.Fatalf("Couldn't add to store: %v", err)
	}

	// /wibble/bibble/1 ENTITY

	if _, err := store.GetEntityByID(ctx, ID("/wibble/bibble/1")); err != nil {
		t.Errorf("should have been able to retrieve an entity, it's an entity")
	}

	if _, err := store.GetRelationByID(ctx, ID("/wibble/bibble/1")); err == nil {
		t.Errorf("should not have been able to retrieve a relation, it's an entity")
	} else if err != ErrNotFound {
		t.Error(err)
	}

	if _, err := store.GetTypeByID(ctx, ID("/wibble/bibble/1")); err == nil {
		t.Errorf("should not have been able to retrieve a type, it's an entity")
	} else if err != ErrNotFound {
		t.Error(err)
	}

	// /wibble/bibble/2 RELATION

	if _, err := store.GetEntityByID(ctx, ID("/wibble/bibble/2")); err == nil {
		t.Errorf("should not have been able to retrieve an entity, it's an relation")
	} else if err != ErrNotFound {
		t.Error(err)
	}

	if _, err := store.GetRelationByID(ctx, ID("/wibble/bibble/2")); err != nil {
		t.Errorf("should have been able to retrieve a relation, it's an relation")
	}

	if _, err := store.GetTypeByID(ctx, ID("/wibble/bibble/2")); err == nil {
		t.Errorf("should not have been able to retrieve a type, it's an relation")
	} else if err != ErrNotFound {
		t.Error(err)
	}

	// /wibble/bibble/3 TYPE

	if _, err := store.GetEntityByID(ctx, ID("/wibble/bibble/3")); err == nil {
		t.Errorf("should not have been able to retrieve an entity, it's a type")
	} else if err != ErrNotFound {
		t.Error(err)
	}

	if _, err := store.GetRelationByID(ctx, ID("/wibble/bibble/3")); err == nil {
		t.Errorf("should not have been able to retrieve a relation, it's a type")
	} else if err != ErrNotFound {
		t.Error(err)
	}

	if _, err := store.GetTypeByID(ctx, ID("/wibble/bibble/3")); err != nil {
		t.Errorf("should have been able to retrieve a type")
	}
	// /wibble/bibble/3 NOT TYPE TYPE

	if _, err := store.GetEntityByID(ctx, ID("/wibble/bibble/4")); err == nil {
		t.Errorf("should not have been able to retrieve an entity, it's not a type")
	} else if err != ErrNotFound {
		t.Error(err)
	}

	if _, err := store.GetRelationByID(ctx, ID("/wibble/bibble/4")); err == nil {
		t.Errorf("should not have been able to retrieve a relation, it's not a type")
	} else if err != ErrNotFound {
		t.Error(err)
	}

	if _, err := store.GetTypeByID(ctx, ID("/wibble/bibble/4")); err == nil {
		t.Errorf("should not have been able to retrieve a type, it's not a type")
	} else if err != ErrNotFound {
		t.Error(err)
	}
}

func TestGetNotFound(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()

	if _, err := store.GetByID(ctx, ID("/wibble/bibble")); err == nil {
		t.Errorf("should not have been able to retrieve a thing")
	} else if err != ErrNotFound {
		t.Error(err)
	}

	if _, err := store.GetEntityByID(ctx, ID("/wibble/bibble")); err == nil {
		t.Errorf("should not have been able to retrieve an entity")
	} else if err != ErrNotFound {
		t.Error(err)
	}

	if _, err := store.GetRelationByID(ctx, ID("/wibble/bibble")); err == nil {
		t.Errorf("should not have been able to retrieve a relation")
	} else if err != ErrNotFound {
		t.Error(err)
	}

	if _, err := store.GetTypeByID(ctx, ID("/wibble/bibble")); err == nil {
		t.Errorf("should not have been able to retrieve a type")
	} else if err != ErrNotFound {
		t.Error(err)
	}
}

func assertList(t *testing.T, ctx context.Context, listFunc func(context.Context,ListOptions) ([]*Thing, error), options ListOptions, expectedSize int, expectedIDs []string) {
	things, err := listFunc(ctx, options)
	if err != nil {
		t.Fatal(err)
	}

	if len(things) != expectedSize {
		t.Fatalf("expected %d things, got %d\n%+v", expectedSize, len(things), things)
	}

	for idx, thing := range things {
		expectedID := ID(expectedIDs[idx])
		actualID := thing.Metadata.ID

		if actualID != expectedID {
			t.Errorf("things[%d]: '%v' doesn't equal expected value '%v'", idx, actualID, expectedID)
		}
	}
}

func compareLists(t *testing.T, ctx context.Context, listFunc1 func(context.Context,ListOptions) ([]*Thing, error), listFunc2 func(context.Context, ListOptions) ([]*Thing, error), options ListOptions) {
	things1, err := listFunc1(ctx, options)
	if err != nil {
		t.Fatal(err)
	}
	things2, err := listFunc2(ctx, options)
	if err != nil {
		t.Fatal(err)
	}

	if len(things1) != len(things2) {
		t.Fatalf("expected two lists to match. %d != %d", len(things1), len(things2))
	}

	for idx, thing := range things1 {
		ID1 := thing.Metadata.ID
		ID2 := things2[idx].Metadata.ID

		if ID1 != ID2 {
			t.Errorf("things1[%d] %v != things[%d] %v", idx, ID1, idx, ID2)
		}
	}
}

func listRelationsForEntity(store Store, ent *Entity) func(context.Context,ListOptions) ([]*Thing, error) {
	return convertRelationsToThings(func(ctx context.Context,opts ListOptions) ([]*Relation, error) {
		return store.ListRelationsForEntity(ctx, ent, opts)
	})
}

func TestListRelationsForEntity(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()

	ent1 := entity("/ent/1")

	if err := store.Add(
		ctx,
		ent1,
		entity("/ent/2"),
		entity("/ent/3"),
		relationBetween("/rel/1", "/ent/1", "/ent/2"),
		relationBetween("/rel/2", "/ent/3", "/ent/1"),
		relationBetween("/rel/3", "/ent/2", "/ent/3"),
	); err != nil {
		t.Fatalf("Couldn't add to store: %v", err)
	}

	assertList(t, ctx, listRelationsForEntity(store, (*Entity)(ent1)), ListOptions{}, 2, []string{
		"/rel/1",
		"/rel/2",
	})

}

func TestList(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()

	if err := store.Add(ctx, entity("/wibble")); err != nil {
		t.Fatal(err)
	}

	// default sort order is ascending
	assertList(t, ctx, store.List, ListOptions{}, 4, []string{
		"/entity",
		"/relation",
		"/type",
		"/wibble",
	})

	// test descending correctly works
	assertList(t, ctx, store.List, ListOptions{SortOrder: SortDescending}, 4, []string{
		"/wibble",
		"/type",
		"/relation",
		"/entity",
	})

	// ask for too many results
	assertList(t, ctx, store.List, ListOptions{NumberOfResults: 1000}, 4, []string{
		"/entity",
		"/relation",
		"/type",
		"/wibble",
	})

	// ask for less results than are in the store
	assertList(t, ctx, store.List, ListOptions{NumberOfResults: 2}, 2, []string{
		"/entity",
		"/relation",
	})

	// offset
	assertList(t, ctx, store.List, ListOptions{NumberOfResults: 1, Offset: 1}, 1, []string{
		"/relation",
	})

	// offset the overlaps the end
	assertList(t, ctx, store.List, ListOptions{NumberOfResults: 3, Offset: 3}, 1, []string{
		"/wibble",
	})

	// ask for offset bigger than the number of things
	assertList(t, ctx, store.List, ListOptions{Offset: 10}, 0, []string{})
}

// the rest of the List* functions are implemented using list by type
// ListByType should include child types

func listByType(store Store, typ *Type) func(context.Context,ListOptions) ([]*Thing, error) {
	return func(ctx context.Context, opts ListOptions) ([]*Thing, error) {
		return store.ListByType(ctx, typ, opts)
	}
}

func convertEntitiesToThings(listFunc func(context.Context,ListOptions) ([]*Entity, error)) func(context.Context,ListOptions) ([]*Thing, error) {
	return func(ctx context.Context, opts ListOptions) ([]*Thing, error) {
		vs, err := listFunc(ctx, opts)
		if err != nil {
			return []*Thing{}, err
		}

		things := make([]*Thing, len(vs))

		for idx, v := range vs {
			thing := Thing(*v)
			things[idx] = &thing
		}

		return things, nil
	}
}

func convertRelationsToThings(listFunc func(context.Context,ListOptions) ([]*Relation, error)) func(context.Context,ListOptions) ([]*Thing, error) {
	return func(ctx context.Context, opts ListOptions) ([]*Thing, error) {
		vs, err := listFunc(ctx, opts)
		if err != nil {
			return []*Thing{}, err
		}

		things := make([]*Thing, len(vs))

		for idx, v := range vs {
			thing := Thing(*v)
			things[idx] = &thing
		}

		return things, nil
	}
}

func convertTypesToThings(listFunc func(context.Context,ListOptions) ([]*Type, error)) func(context.Context,ListOptions) ([]*Thing, error) {
	return func(ctx context.Context, opts ListOptions) ([]*Thing, error) {
		vs, err := listFunc(ctx, opts)
		if err != nil {
			return []*Thing{}, err
		}

		things := make([]*Thing, len(vs))

		for idx, v := range vs {
			thing := Thing(*v)
			things[idx] = &thing
		}

		return things, nil
	}
}

func TestListByType(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()

	if err := store.Add(ctx,
		entity("/wibble/bibble/1"),
		entity("/wibble/bibble/5"),
		entity("/wibble/bibble/4"),
		relation("/wibble/bibble/3"),
		relation("/wibble/bibble/6"),
	); err != nil {
		t.Fatalf("Couldn't add to store: %v", err)
	}

	// ENTITIES

	listEntities := listByType(store, EntityType)

	// default sort order is ascending
	assertList(t, ctx, listEntities, ListOptions{}, 3, []string{
		"/wibble/bibble/1",
		"/wibble/bibble/4",
		"/wibble/bibble/5",
	})

	// test descending correctly works
	assertList(t, ctx, listEntities, ListOptions{SortOrder: SortDescending}, 3, []string{
		"/wibble/bibble/5",
		"/wibble/bibble/4",
		"/wibble/bibble/1",
	})

	// ask for too many results
	assertList(t, ctx, listEntities, ListOptions{NumberOfResults: 1000}, 3, []string{
		"/wibble/bibble/1",
		"/wibble/bibble/4",
		"/wibble/bibble/5",
	})

	// ask for less results than are in the store
	assertList(t, ctx, listEntities, ListOptions{NumberOfResults: 2}, 2, []string{
		"/wibble/bibble/1",
		"/wibble/bibble/4",
	})

	// ask for offset bigger than the number of things
	assertList(t, ctx, listEntities, ListOptions{Offset: 10}, 0, []string{})

	compareLists(t, ctx, listEntities, convertEntitiesToThings(store.ListEntities), ListOptions{})
	compareLists(t, ctx, listEntities, convertEntitiesToThings(store.ListEntities), ListOptions{SortOrder: SortDescending})
	compareLists(t, ctx, listEntities, convertEntitiesToThings(store.ListEntities), ListOptions{NumberOfResults: 1000})
	compareLists(t, ctx, listEntities, convertEntitiesToThings(store.ListEntities), ListOptions{NumberOfResults: 1})
	compareLists(t, ctx, listEntities, convertEntitiesToThings(store.ListEntities), ListOptions{Offset: 10})

	// RELATIONS

	listRelations := listByType(store, RelationType)

	// default sort order is ascending
	assertList(t, ctx, listRelations, ListOptions{}, 2, []string{
		"/wibble/bibble/3",
		"/wibble/bibble/6",
	})

	// test descending correctly works
	assertList(t, ctx, listRelations, ListOptions{SortOrder: SortDescending}, 2, []string{
		"/wibble/bibble/6",
		"/wibble/bibble/3",
	})

	// ask for too many results
	assertList(t, ctx, listRelations, ListOptions{NumberOfResults: 1000}, 2, []string{
		"/wibble/bibble/3",
		"/wibble/bibble/6",
	})

	// ask for less results than are in the store
	assertList(t, ctx, listRelations, ListOptions{NumberOfResults: 1}, 1, []string{
		"/wibble/bibble/3",
	})

	// ask for offset bigger than the number of things
	assertList(t, ctx, listRelations, ListOptions{Offset: 10}, 0, []string{})

	compareLists(t, ctx, listRelations, convertRelationsToThings(store.ListRelations), ListOptions{})
	compareLists(t, ctx, listRelations, convertRelationsToThings(store.ListRelations), ListOptions{SortOrder: SortDescending})
	compareLists(t, ctx, listRelations, convertRelationsToThings(store.ListRelations), ListOptions{NumberOfResults: 1000})
	compareLists(t, ctx, listRelations, convertRelationsToThings(store.ListRelations), ListOptions{NumberOfResults: 1})
	compareLists(t, ctx, listRelations, convertRelationsToThings(store.ListRelations), ListOptions{Offset: 10})

	// TYPES

	listTypes := listByType(store, TypeType)

	// default sort order is ascending
	assertList(t, ctx, listTypes, ListOptions{}, 3, []string{
		"/entity",
		"/relation",
		"/type",
	})

	// test descending correctly works
	assertList(t, ctx, listTypes, ListOptions{SortOrder: SortDescending}, 3, []string{
		"/type",
		"/relation",
		"/entity",
	})

	// ask for too many results
	assertList(t, ctx, listTypes, ListOptions{NumberOfResults: 1000}, 3, []string{
		"/entity",
		"/relation",
		"/type",
	})

	// ask for less results than are in the store
	assertList(t, ctx, listTypes, ListOptions{NumberOfResults: 1}, 1, []string{
		"/entity",
	})

	// ask for offset bigger than the number of things
	assertList(t, ctx, listTypes, ListOptions{Offset: 10}, 0, []string{})

	compareLists(t, ctx, listTypes, convertTypesToThings(store.ListTypes), ListOptions{})
	compareLists(t, ctx, listTypes, convertTypesToThings(store.ListTypes), ListOptions{SortOrder: SortDescending})
	compareLists(t, ctx, listTypes, convertTypesToThings(store.ListTypes), ListOptions{NumberOfResults: 1000})
	compareLists(t, ctx, listTypes, convertTypesToThings(store.ListTypes), ListOptions{NumberOfResults: 1})
	compareLists(t, ctx, listTypes, convertTypesToThings(store.ListTypes), ListOptions{Offset: 10})
}

func TestWatchByType(t *testing.T) {
	store := NewInMemoryStore()

	closedWG := sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())

	closedWG.Add(3)

	typesList1 := []*Thing{}
	if ch, err := store.WatchByType(ctx, TypeType); err != nil {
		t.Fatalf("Couldn't watch type: %v", err)
	} else {
		go appendFromChannel(ch, &typesList1, &closedWG)
	}

	typesList2 := []*Thing{}
	if ch, err := store.WatchByType(ctx, TypeType); err != nil {
		t.Fatalf("Couldn't watch type: %v", err)
	} else {
		go appendFromChannel(ch, &typesList2, &closedWG)
	}

	entitiesList1 := []*Thing{}
	if ch, err := store.WatchByType(ctx, EntityType); err != nil {
		t.Fatalf("Couldn't watch entity: %v", err)
	} else {
		go appendFromChannel(ch, &entitiesList1, &closedWG)
	}

	// do some changes
	if err := store.Add(
		ctx,
		ntype("/wibble"),
		ntype("/bibble"),
		entity("/nibble"),
	); err != nil {
		t.Fatalf("Couldn't add the bits: %v", err)
	}

	// cancel the context
	cancel()
	closedWG.Wait()

	// check the expected numbers
	if expected := 2; len(typesList1) != expected {
		t.Errorf("Expected there to be %d types, but it was %d", expected, len(typesList1))
	}
	if expected := 2; len(typesList2) != expected {
		t.Errorf("Expected there to be %d types, but it was %d", expected, len(typesList2))
	}
	if expected := 1; len(entitiesList1) != expected {
		t.Errorf("Expected there to be %d types, but it was %d", expected, len(entitiesList1))
	}

	// do some changes
	if err := store.Add(
		ctx,
		ntype("/wibble/1"),
		ntype("/bibble/1"),
		entity("/nibble/1"),
	); err != nil {
		t.Fatalf("Couldn't add the bits: %v", err)
	}

	// check the numbers haven't changed
	if expected := 2; len(typesList1) != expected {
		t.Errorf("Expected there to be %d types, but it was %d", expected, len(typesList1))
	}
	if expected := 2; len(typesList2) != expected {
		t.Errorf("Expected there to be %d types, but it was %d", expected, len(typesList2))
	}
	if expected := 1; len(entitiesList1) != expected {
		t.Errorf("Expected there to be %d types, but it was %d", expected, len(entitiesList1))
	}

}
