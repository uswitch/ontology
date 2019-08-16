package store

import (
	"testing"
)

func thingWithType(thingID string, typeID string) *Thing {
	return &Thing{
		Metadata: Metadata{
			ID:   ID(thingID),
			Type: ID(typeID),
		},
	}
}
func entity(ID string) *Thing   { return thingWithType(ID, "/entity") }
func relation(ID string) *Thing { return thingWithType(ID, "/relation") }
func ntype(ID string) *Thing    { return thingWithType(ID, "/type") }

func TestLen(t *testing.T) {
	store := NewInMemoryStore()

	if num, err := store.Len(); err != nil {
		t.Error(err)
	} else if num != 3 {
		t.Errorf("Store should have 3 base types, has %d", num)
	}

	if err := store.Add(entity("/wibble")); err != nil {
		t.Fatal(err)
	}

	if num, err := store.Len(); err != nil {
		t.Error(err)
	} else if num != 4 {
		t.Errorf("Store should have 4 entries, has %d", num)
	}
}

func TestIsA(t *testing.T) {
	store := NewInMemoryStore()

	if ok, err := store.IsA(EntityType.Thing(), TypeType); !ok {
		t.Error("EntityType should be a TypeType")
	} else if err != nil {
		t.Error(err)
	}

	if ok, err := store.IsA(RelationType.Thing(), TypeType); !ok {
		t.Error("RelationType should be a TypeType")
	} else if err != nil {
		t.Error(err)
	}

	if ok, err := store.IsA(TypeType.Thing(), TypeType); !ok {
		t.Error("TypeType should be a TypeType")
	} else if err != nil {
		t.Error(err)
	}

	if ok, err := store.IsA(entity("/wibble/ent").Thing(), EntityType); !ok {
		t.Error("An entity should be type EntityType")
	} else if err != nil {
		t.Error(err)
	}
	if ok, err := store.IsA(entity("/wibble/ent").Thing(), RelationType); ok {
		t.Error("An entity should not be type RelationType")
	} else if err != nil {
		t.Error(err)
	}

	if ok, err := store.IsA(relation("/wibble/rel").Thing(), RelationType); !ok {
		t.Error("A relation should be type RelationType")
	} else if err != nil {
		t.Error(err)
	}
	if ok, err := store.IsA(relation("/wibble/rel").Thing(), EntityType); ok {
		t.Error("A relation should not be type EntityType")
	} else if err != nil {
		t.Error(err)
	}
}

func TestAddAndGet(t *testing.T) {
	store := NewInMemoryStore()

	if err := store.Add(entity("/wibble/bibble")); err != nil {
		t.Fatalf("Couldn't add to store: %v", err)
	}

	if thing, err := store.GetByID(ID("/wibble/bibble")); err != nil {
		t.Error(err)
	} else if thing.Metadata.ID != "/wibble/bibble" {
		t.Errorf("thing had wrong ID: %s", thing.Metadata.ID)
	}

	if entity, err := store.GetEntityByID(ID("/wibble/bibble")); err != nil {
		t.Error(err)
	} else if entity.Metadata.ID != "/wibble/bibble" {
		t.Errorf("entity had wrong ID: %s", entity.Metadata.ID)
	}

	if _, err := store.GetRelationByID(ID("/wibble/bibble")); err == nil {
		t.Errorf("should not have been able to retrieve a relation, it's an entity")
	} else if err != ErrNotFound {
		t.Error(err)
	}

	if _, err := store.GetTypeByID(ID("/wibble/bibble")); err == nil {
		t.Errorf("should not have been able to retrieve a type, it's an entity")
	} else if err != ErrNotFound {
		t.Error(err)
	}
}

func TestGetCorrectType(t *testing.T) {
	store := NewInMemoryStore()

	if err := store.AddAll([]*Thing{
		entity("/wibble/bibble/1"),
		relation("/wibble/bibble/2"),
		ntype("/wibble/bibble/3"),
		thingWithType("/wibble/bibble/4", "/type/"),
	}); err != nil {
		t.Fatalf("Couldn't add to store: %v", err)
	}

	// /wibble/bibble/1 ENTITY

	if _, err := store.GetEntityByID(ID("/wibble/bibble/1")); err != nil {
		t.Errorf("should have been able to retrieve an entity, it's an entity")
	}

	if _, err := store.GetRelationByID(ID("/wibble/bibble/1")); err == nil {
		t.Errorf("should not have been able to retrieve a relation, it's an entity")
	} else if err != ErrNotFound {
		t.Error(err)
	}

	if _, err := store.GetTypeByID(ID("/wibble/bibble/1")); err == nil {
		t.Errorf("should not have been able to retrieve a type, it's an entity")
	} else if err != ErrNotFound {
		t.Error(err)
	}

	// /wibble/bibble/2 RELATION

	if _, err := store.GetEntityByID(ID("/wibble/bibble/2")); err == nil {
		t.Errorf("should not have been able to retrieve an entity, it's an relation")
	} else if err != ErrNotFound {
		t.Error(err)
	}

	if _, err := store.GetRelationByID(ID("/wibble/bibble/2")); err != nil {
		t.Errorf("should have been able to retrieve a relation, it's an relation")
	}

	if _, err := store.GetTypeByID(ID("/wibble/bibble/2")); err == nil {
		t.Errorf("should not have been able to retrieve a type, it's an relation")
	} else if err != ErrNotFound {
		t.Error(err)
	}

	// /wibble/bibble/3 TYPE

	if _, err := store.GetEntityByID(ID("/wibble/bibble/3")); err == nil {
		t.Errorf("should not have been able to retrieve an entity, it's a type")
	} else if err != ErrNotFound {
		t.Error(err)
	}

	if _, err := store.GetRelationByID(ID("/wibble/bibble/3")); err == nil {
		t.Errorf("should not have been able to retrieve a relation, it's a type")
	} else if err != ErrNotFound {
		t.Error(err)
	}

	if _, err := store.GetTypeByID(ID("/wibble/bibble/3")); err != nil {
		t.Errorf("should have been able to retrieve a type")
	}
	// /wibble/bibble/3 NOT TYPE TYPE

	if _, err := store.GetEntityByID(ID("/wibble/bibble/4")); err == nil {
		t.Errorf("should not have been able to retrieve an entity, it's not a type")
	} else if err != ErrNotFound {
		t.Error(err)
	}

	if _, err := store.GetRelationByID(ID("/wibble/bibble/4")); err == nil {
		t.Errorf("should not have been able to retrieve a relation, it's not a type")
	} else if err != ErrNotFound {
		t.Error(err)
	}

	if _, err := store.GetTypeByID(ID("/wibble/bibble/4")); err == nil {
		t.Errorf("should not have been able to retrieve a type, it's not a type")
	} else if err != ErrNotFound {
		t.Error(err)
	}
}

func TestGetNotFound(t *testing.T) {
	store := NewInMemoryStore()

	if _, err := store.GetByID(ID("/wibble/bibble")); err == nil {
		t.Errorf("should not have been able to retrieve a thing")
	} else if err != ErrNotFound {
		t.Error(err)
	}

	if _, err := store.GetEntityByID(ID("/wibble/bibble")); err == nil {
		t.Errorf("should not have been able to retrieve an entity")
	} else if err != ErrNotFound {
		t.Error(err)
	}

	if _, err := store.GetRelationByID(ID("/wibble/bibble")); err == nil {
		t.Errorf("should not have been able to retrieve a relation")
	} else if err != ErrNotFound {
		t.Error(err)
	}

	if _, err := store.GetTypeByID(ID("/wibble/bibble")); err == nil {
		t.Errorf("should not have been able to retrieve a type")
	} else if err != ErrNotFound {
		t.Error(err)
	}
}

func assertList(t *testing.T, listFunc func(ListOptions) ([]*Thing, error), options ListOptions, expectedSize int, expectedIDs []string) {
	things, err := listFunc(options)
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

func compareLists(t *testing.T, listFunc1 func(ListOptions) ([]*Thing, error), listFunc2 func(ListOptions) ([]*Thing, error), options ListOptions) {
	things1, err := listFunc1(options)
	if err != nil {
		t.Fatal(err)
	}
	things2, err := listFunc2(options)
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

func TestList(t *testing.T) {
	store := NewInMemoryStore()

	if err := store.Add(entity("/wibble")); err != nil {
		t.Fatal(err)
	}

	// default sort order is ascending
	assertList(t, store.List, ListOptions{}, 4, []string{
		"/entity",
		"/relation",
		"/type",
		"/wibble",
	})

	// test descending correctly works
	assertList(t, store.List, ListOptions{SortOrder: SortDescending}, 4, []string{
		"/wibble",
		"/type",
		"/relation",
		"/entity",
	})

	// ask for too many results
	assertList(t, store.List, ListOptions{NumberOfResults: 1000}, 4, []string{
		"/entity",
		"/relation",
		"/type",
		"/wibble",
	})

	// ask for less results than are in the store
	assertList(t, store.List, ListOptions{NumberOfResults: 2}, 2, []string{
		"/entity",
		"/relation",
	})

	// offset
	assertList(t, store.List, ListOptions{NumberOfResults: 1, Offset: 1}, 1, []string{
		"/relation",
	})

	// offset the overlaps the end
	assertList(t, store.List, ListOptions{NumberOfResults: 3, Offset: 3}, 1, []string{
		"/wibble",
	})

	// ask for offset bigger than the number of things
	assertList(t, store.List, ListOptions{Offset: 10}, 0, []string{})
}

// the rest of the List* functions are implemented using list by type
// ListByType should include child types

func listByType(store Store, typ *Type) func(ListOptions) ([]*Thing, error) {
	return func(opts ListOptions) ([]*Thing, error) {
		return store.ListByType(typ, opts)
	}
}

func convertEntitiesToThings(listFunc func(ListOptions) ([]*Entity, error)) func(ListOptions) ([]*Thing, error) {
	return func(opts ListOptions) ([]*Thing, error) {
		vs, err := listFunc(opts)
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

func convertRelationsToThings(listFunc func(ListOptions) ([]*Relation, error)) func(ListOptions) ([]*Thing, error) {
	return func(opts ListOptions) ([]*Thing, error) {
		vs, err := listFunc(opts)
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

func convertTypesToThings(listFunc func(ListOptions) ([]*Type, error)) func(ListOptions) ([]*Thing, error) {
	return func(opts ListOptions) ([]*Thing, error) {
		vs, err := listFunc(opts)
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

	if err := store.AddAll([]*Thing{
		entity("/wibble/bibble/1"),
		entity("/wibble/bibble/5"),
		entity("/wibble/bibble/4"),
		relation("/wibble/bibble/3"),
		relation("/wibble/bibble/6"),
	}); err != nil {
		t.Fatalf("Couldn't add to store: %v", err)
	}

	// ENTITIES

	listEntities := listByType(store, EntityType)

	// default sort order is ascending
	assertList(t, listEntities, ListOptions{}, 3, []string{
		"/wibble/bibble/1",
		"/wibble/bibble/4",
		"/wibble/bibble/5",
	})

	// test descending correctly works
	assertList(t, listEntities, ListOptions{SortOrder: SortDescending}, 3, []string{
		"/wibble/bibble/5",
		"/wibble/bibble/4",
		"/wibble/bibble/1",
	})

	// ask for too many results
	assertList(t, listEntities, ListOptions{NumberOfResults: 1000}, 3, []string{
		"/wibble/bibble/1",
		"/wibble/bibble/4",
		"/wibble/bibble/5",
	})

	// ask for less results than are in the store
	assertList(t, listEntities, ListOptions{NumberOfResults: 2}, 2, []string{
		"/wibble/bibble/1",
		"/wibble/bibble/4",
	})

	// ask for offset bigger than the number of things
	assertList(t, listEntities, ListOptions{Offset: 10}, 0, []string{})

	compareLists(t, listEntities, convertEntitiesToThings(store.ListEntities), ListOptions{})
	compareLists(t, listEntities, convertEntitiesToThings(store.ListEntities), ListOptions{SortOrder: SortDescending})
	compareLists(t, listEntities, convertEntitiesToThings(store.ListEntities), ListOptions{NumberOfResults: 1000})
	compareLists(t, listEntities, convertEntitiesToThings(store.ListEntities), ListOptions{NumberOfResults: 1})
	compareLists(t, listEntities, convertEntitiesToThings(store.ListEntities), ListOptions{Offset: 10})

	// RELATIONS

	listRelations := listByType(store, RelationType)

	// default sort order is ascending
	assertList(t, listRelations, ListOptions{}, 2, []string{
		"/wibble/bibble/3",
		"/wibble/bibble/6",
	})

	// test descending correctly works
	assertList(t, listRelations, ListOptions{SortOrder: SortDescending}, 2, []string{
		"/wibble/bibble/6",
		"/wibble/bibble/3",
	})

	// ask for too many results
	assertList(t, listRelations, ListOptions{NumberOfResults: 1000}, 2, []string{
		"/wibble/bibble/3",
		"/wibble/bibble/6",
	})

	// ask for less results than are in the store
	assertList(t, listRelations, ListOptions{NumberOfResults: 1}, 1, []string{
		"/wibble/bibble/3",
	})

	// ask for offset bigger than the number of things
	assertList(t, listRelations, ListOptions{Offset: 10}, 0, []string{})

	compareLists(t, listRelations, convertRelationsToThings(store.ListRelations), ListOptions{})
	compareLists(t, listRelations, convertRelationsToThings(store.ListRelations), ListOptions{SortOrder: SortDescending})
	compareLists(t, listRelations, convertRelationsToThings(store.ListRelations), ListOptions{NumberOfResults: 1000})
	compareLists(t, listRelations, convertRelationsToThings(store.ListRelations), ListOptions{NumberOfResults: 1})
	compareLists(t, listRelations, convertRelationsToThings(store.ListRelations), ListOptions{Offset: 10})

	// TYPES

	listTypes := listByType(store, TypeType)

	// default sort order is ascending
	assertList(t, listTypes, ListOptions{}, 3, []string{
		"/entity",
		"/relation",
		"/type",
	})

	// test descending correctly works
	assertList(t, listTypes, ListOptions{SortOrder: SortDescending}, 3, []string{
		"/type",
		"/relation",
		"/entity",
	})

	// ask for too many results
	assertList(t, listTypes, ListOptions{NumberOfResults: 1000}, 3, []string{
		"/entity",
		"/relation",
		"/type",
	})

	// ask for less results than are in the store
	assertList(t, listTypes, ListOptions{NumberOfResults: 1}, 1, []string{
		"/entity",
	})

	// ask for offset bigger than the number of things
	assertList(t, listTypes, ListOptions{Offset: 10}, 0, []string{})

	compareLists(t, listTypes, convertTypesToThings(store.ListTypes), ListOptions{})
	compareLists(t, listTypes, convertTypesToThings(store.ListTypes), ListOptions{SortOrder: SortDescending})
	compareLists(t, listTypes, convertTypesToThings(store.ListTypes), ListOptions{NumberOfResults: 1000})
	compareLists(t, listTypes, convertTypesToThings(store.ListTypes), ListOptions{NumberOfResults: 1})
	compareLists(t, listTypes, convertTypesToThings(store.ListTypes), ListOptions{Offset: 10})
}
