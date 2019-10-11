package storetest

import (
	"context"
	"sync"
	"testing"

	"github.com/uswitch/ontology/pkg/store"
)

func Conformance(t *testing.T, newStore func() store.Store) {
	tests := map[string]func(*testing.T, store.Store){
		"Len": TestLen,
		"IsA": TestIsA,
		/*"AddAndGet":      TestAddAndGet,
		"GetCorrectType": TestGetCorrectType,
		"GetNotFound":    TestGetNotFound,

		"List":                           TestList,
		"ListByType":                     TestListByType,
		"ListRelationsForEntity":         TestListRelationsForEntity,
		"ListRelationsForEntityWithType": TestListRelationsForEntityWithType,
		"ListRelationsForEntityBadType":  TestListRelationsForEntityBadType,

		"WatchByID":   TestWatchByID,
		"WatchByType": TestWatchByType,

		"TypeProperties": TestTypeProperties,

		"Validate": TestValidate,*/
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			s := newStore()
			test(t, s)
		})
	}
}

func thingWithType(thingID string, typeID string, properties store.Properties) *store.Thing {
	thing := &store.Thing{
		Metadata: store.Metadata{
			ID:   store.ID(thingID),
			Type: store.ID(typeID),
		},
	}

	if properties != nil {
		thing.Properties = properties
	}

	return thing
}
func entity(ID string) *store.Thing              { return thingWithType(ID, "/entity", nil) }
func entityWithType(ID, typ string) *store.Thing { return thingWithType(ID, typ, nil) }
func relation(ID string) *store.Thing            { return thingWithType(ID, "/relation", nil) }
func ntype(ID string) *store.Thing               { return thingWithType(ID, "/type", nil) }
func typ(ID, parent string, spec map[string]interface{}) *store.Type {
	props := store.Properties{}

	if parent != "" {
		props["parent"] = parent
	}
	props["spec"] = spec

	return (*store.Type)(thingWithType(ID, "/type", props))
}
func relationBetween(ID, a, b string) *store.Thing {
	return relationBetweenWithType(ID, "/relation", a, b)
}
func relationBetweenWithType(ID, typ, a, b string) *store.Thing {
	return thingWithType(
		ID, typ,
		store.Properties{
			"a": a,
			"b": b,
		},
	)
}

func TestLen(t *testing.T, s store.Store) {
	ctx := context.Background()

	if num, err := s.Len(ctx); err != nil {
		t.Error(err)
	} else if num != 5 {
		t.Errorf("Store should have 5 base types, has %d", num)
	}

	if err := s.Add(ctx, entity("/wibble")); err != nil {
		t.Fatal(err)
	}

	if num, err := s.Len(ctx); err != nil {
		t.Error(err)
	} else if num != 6 {
		t.Errorf("Store should have 6 entries, has %d", num)
	}
}

func TestIsA(t *testing.T, s store.Store) {
	ctx := context.Background()

	if ok, err := s.IsA(ctx, store.EntityType.Thing(), store.TypeType); !ok {
		t.Error("EntityType should be a TypeType")
	} else if err != nil {
		t.Error(err)
	}

	if ok, err := s.IsA(ctx, store.RelationType.Thing(), store.TypeType); !ok {
		t.Error("RelationType should be a TypeType")
	} else if err != nil {
		t.Error(err)
	}

	if ok, err := s.IsA(ctx, store.TypeType.Thing(), store.TypeType); !ok {
		t.Error("TypeType should be a TypeType")
	} else if err != nil {
		t.Error(err)
	}

	if ok, err := s.IsA(ctx, entity("/wibble/ent").Thing(), store.EntityType); !ok {
		t.Error("An entity should be type EntityType")
	} else if err != nil {
		t.Error(err)
	}
	if ok, err := s.IsA(ctx, entity("/wibble/ent").Thing(), store.RelationType); ok {
		t.Error("An entity should not be type RelationType")
	} else if err != nil {
		t.Error(err)
	}

	if ok, err := s.IsA(ctx, relation("/wibble/rel").Thing(), store.RelationType); !ok {
		t.Error("A relation should be type RelationType")
	} else if err != nil {
		t.Error(err)
	}
	if ok, err := s.IsA(ctx, relation("/wibble/rel").Thing(), store.EntityType); ok {
		t.Error("A relation should not be type EntityType")
	} else if err != nil {
		t.Error(err)
	}
}

func TestAddAndGet(t *testing.T, s store.Store) {
	ctx := context.Background()

	if err := s.Add(ctx, entity("/wibble/bibble")); err != nil {
		t.Fatalf("Couldn't add to store: %v", err)
	}

	if thing, err := s.GetByID(ctx, store.ID("/wibble/bibble")); err != nil {
		t.Error(err)
	} else if thing.Metadata.ID != "/wibble/bibble" {
		t.Errorf("thing had wrong ID: %s", thing.Metadata.ID)
	}

	if entity, err := s.GetEntityByID(ctx, store.ID("/wibble/bibble")); err != nil {
		t.Error(err)
	} else if entity.Metadata.ID != "/wibble/bibble" {
		t.Errorf("entity had wrong ID: %s", entity.Metadata.ID)
	}

	if _, err := s.GetRelationByID(ctx, store.ID("/wibble/bibble")); err == nil {
		t.Errorf("should not have been able to retrieve a relation, it's an entity")
	} else if err != store.ErrNotFound {
		t.Error(err)
	}

	if _, err := s.GetTypeByID(ctx, store.ID("/wibble/bibble")); err == nil {
		t.Errorf("should not have been able to retrieve a type, it's an entity")
	} else if err != store.ErrNotFound {
		t.Error(err)
	}
}

func TestGetCorrectType(t *testing.T, s store.Store) {
	ctx := context.Background()

	if err := s.Add(ctx,
		entity("/wibble/bibble/1"),
		relation("/wibble/bibble/2"),
		ntype("/wibble/bibble/3"),
		thingWithType("/wibble/bibble/4", "/type/", nil),
	); err != nil {
		t.Fatalf("Couldn't add to store: %v", err)
	}

	// /wibble/bibble/1 ENTITY

	if _, err := s.GetEntityByID(ctx, store.ID("/wibble/bibble/1")); err != nil {
		t.Errorf("should have been able to retrieve an entity, it's an entity")
	}

	if _, err := s.GetRelationByID(ctx, store.ID("/wibble/bibble/1")); err == nil {
		t.Errorf("should not have been able to retrieve a relation, it's an entity")
	} else if err != store.ErrNotFound {
		t.Error(err)
	}

	if _, err := s.GetTypeByID(ctx, store.ID("/wibble/bibble/1")); err == nil {
		t.Errorf("should not have been able to retrieve a type, it's an entity")
	} else if err != store.ErrNotFound {
		t.Error(err)
	}

	// /wibble/bibble/2 RELATION

	if _, err := s.GetEntityByID(ctx, store.ID("/wibble/bibble/2")); err == nil {
		t.Errorf("should not have been able to retrieve an entity, it's an relation")
	} else if err != store.ErrNotFound {
		t.Error(err)
	}

	if _, err := s.GetRelationByID(ctx, store.ID("/wibble/bibble/2")); err != nil {
		t.Errorf("should have been able to retrieve a relation, it's an relation")
	}

	if _, err := s.GetTypeByID(ctx, store.ID("/wibble/bibble/2")); err == nil {
		t.Errorf("should not have been able to retrieve a type, it's an relation")
	} else if err != store.ErrNotFound {
		t.Error(err)
	}

	// /wibble/bibble/3 TYPE

	if _, err := s.GetEntityByID(ctx, store.ID("/wibble/bibble/3")); err == nil {
		t.Errorf("should not have been able to retrieve an entity, it's a type")
	} else if err != store.ErrNotFound {
		t.Error(err)
	}

	if _, err := s.GetRelationByID(ctx, store.ID("/wibble/bibble/3")); err == nil {
		t.Errorf("should not have been able to retrieve a relation, it's a type")
	} else if err != store.ErrNotFound {
		t.Error(err)
	}

	if _, err := s.GetTypeByID(ctx, store.ID("/wibble/bibble/3")); err != nil {
		t.Errorf("should have been able to retrieve a type")
	}
	// /wibble/bibble/3 NOT TYPE TYPE

	if _, err := s.GetEntityByID(ctx, store.ID("/wibble/bibble/4")); err == nil {
		t.Errorf("should not have been able to retrieve an entity, it's not a type")
	} else if err != store.ErrNotFound {
		t.Error(err)
	}

	if _, err := s.GetRelationByID(ctx, store.ID("/wibble/bibble/4")); err == nil {
		t.Errorf("should not have been able to retrieve a relation, it's not a type")
	} else if err != store.ErrNotFound {
		t.Error(err)
	}

	if _, err := s.GetTypeByID(ctx, store.ID("/wibble/bibble/4")); err == nil {
		t.Errorf("should not have been able to retrieve a type, it's not a type")
	} else if err != store.ErrNotFound {
		t.Error(err)
	}
}

func TestGetNotFound(t *testing.T, s store.Store) {
	ctx := context.Background()

	if _, err := s.GetByID(ctx, store.ID("/wibble/bibble")); err == nil {
		t.Errorf("should not have been able to retrieve a thing")
	} else if err != store.ErrNotFound {
		t.Error(err)
	}

	if _, err := s.GetEntityByID(ctx, store.ID("/wibble/bibble")); err == nil {
		t.Errorf("should not have been able to retrieve an entity")
	} else if err != store.ErrNotFound {
		t.Error(err)
	}

	if _, err := s.GetRelationByID(ctx, store.ID("/wibble/bibble")); err == nil {
		t.Errorf("should not have been able to retrieve a relation")
	} else if err != store.ErrNotFound {
		t.Error(err)
	}

	if _, err := s.GetTypeByID(ctx, store.ID("/wibble/bibble")); err == nil {
		t.Errorf("should not have been able to retrieve a type")
	} else if err != store.ErrNotFound {
		t.Error(err)
	}
}

func assertList(t *testing.T, ctx context.Context, listFunc func(context.Context, store.ListOptions) ([]*store.Thing, error), options store.ListOptions, expectedSize int, expectedIDs []string) {
	things, err := listFunc(ctx, options)
	if err != nil {
		t.Fatal(err)
	}

	if len(things) != expectedSize {
		t.Fatalf("expected %d things, got %d\n%+v", expectedSize, len(things), things)
	}

	for idx, thing := range things {
		expectedID := store.ID(expectedIDs[idx])
		actualID := thing.Metadata.ID

		if actualID != expectedID {
			t.Errorf("things[%d]: '%v' doesn't equal expected value '%v'", idx, actualID, expectedID)
		}
	}
}

func compareLists(t *testing.T, ctx context.Context, listFunc1 func(context.Context, store.ListOptions) ([]*store.Thing, error), listFunc2 func(context.Context, store.ListOptions) ([]*store.Thing, error), options store.ListOptions) {
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

func listRelationsForEntity(s store.Store, typ *store.Type, ent *store.Entity) func(context.Context, store.ListOptions) ([]*store.Thing, error) {
	return convertRelationsToThings(func(ctx context.Context, opts store.ListOptions) ([]*store.Relation, error) {
		return s.ListRelationsForEntity(ctx, typ, ent, opts)
	})
}

func convertEntitiesToThings(listFunc func(context.Context, store.ListOptions) ([]*store.Entity, error)) func(context.Context, store.ListOptions) ([]*store.Thing, error) {
	return func(ctx context.Context, opts store.ListOptions) ([]*store.Thing, error) {
		vs, err := listFunc(ctx, opts)
		if err != nil {
			return []*store.Thing{}, err
		}

		things := make([]*store.Thing, len(vs))

		for idx, v := range vs {
			thing := store.Thing(*v)
			things[idx] = &thing
		}

		return things, nil
	}
}

func convertRelationsToThings(listFunc func(context.Context, store.ListOptions) ([]*store.Relation, error)) func(context.Context, store.ListOptions) ([]*store.Thing, error) {
	return func(ctx context.Context, opts store.ListOptions) ([]*store.Thing, error) {
		vs, err := listFunc(ctx, opts)
		if err != nil {
			return []*store.Thing{}, err
		}

		things := make([]*store.Thing, len(vs))

		for idx, v := range vs {
			thing := store.Thing(*v)
			things[idx] = &thing
		}

		return things, nil
	}
}

func convertTypesToThings(listFunc func(context.Context, store.ListOptions) ([]*store.Type, error)) func(context.Context, store.ListOptions) ([]*store.Thing, error) {
	return func(ctx context.Context, opts store.ListOptions) ([]*store.Thing, error) {
		vs, err := listFunc(ctx, opts)
		if err != nil {
			return []*store.Thing{}, err
		}

		things := make([]*store.Thing, len(vs))

		for idx, v := range vs {
			thing := store.Thing(*v)
			things[idx] = &thing
		}

		return things, nil
	}
}

func TestListRelationsForEntity(t *testing.T, s store.Store) {
	ctx := context.Background()

	ent1 := entity("/ent/1")

	if err := s.Add(
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

	assertList(t, ctx, listRelationsForEntity(s, nil, (*store.Entity)(ent1)), store.ListOptions{}, 2, []string{
		"/rel/1",
		"/rel/2",
	})

}

func TestListRelationsForEntityWithType(t *testing.T, s store.Store) {
	ctx := context.Background()

	ent1 := entity("/ent/1")
	relType := typ("/bibble", store.RelationType.Metadata.ID.String(), nil)

	if err := s.Add(
		ctx,
		ent1,
		relType,
		entity("/ent/2"),
		entity("/ent/3"),
		relationBetween("/rel/1", "/ent/1", "/ent/2"),
		relationBetweenWithType("/rel/2", relType.Metadata.ID.String(), "/ent/3", "/ent/1"),
		relationBetween("/rel/3", "/ent/2", "/ent/3"),
	); err != nil {
		t.Fatalf("Couldn't add to store: %v", err)
	}

	assertList(t, ctx, listRelationsForEntity(s, relType, (*store.Entity)(ent1)), store.ListOptions{}, 1, []string{
		"/rel/2",
	})

}

func TestListRelationsForEntityBadType(t *testing.T, s store.Store) {
	ctx := context.Background()

	ent1 := entity("/ent/1")

	_, err := s.ListRelationsForEntity(ctx, store.EntityType, (*store.Entity)(ent1), store.ListOptions{})
	if err == nil {
		t.Errorf("Should have raised an error as entity is not a relation type")
	}
}

func TestList(t *testing.T, s store.Store) {
	ctx := context.Background()

	if err := s.Add(ctx, entity("/wibble")); err != nil {
		t.Fatal(err)
	}

	// default sort order is ascending
	assertList(t, ctx, s.List, store.ListOptions{}, 4, []string{
		"/entity",
		"/relation",
		"/type",
		"/wibble",
	})

	// test descending correctly works
	assertList(t, ctx, s.List, store.ListOptions{SortOrder: store.SortDescending}, 4, []string{
		"/wibble",
		"/type",
		"/relation",
		"/entity",
	})

	// ask for too many results
	assertList(t, ctx, s.List, store.ListOptions{NumberOfResults: 1000}, 4, []string{
		"/entity",
		"/relation",
		"/type",
		"/wibble",
	})

	// ask for less results than are in the store
	assertList(t, ctx, s.List, store.ListOptions{NumberOfResults: 2}, 2, []string{
		"/entity",
		"/relation",
	})

	// offset
	assertList(t, ctx, s.List, store.ListOptions{NumberOfResults: 1, Offset: 1}, 1, []string{
		"/relation",
	})

	// offset the overlaps the end
	assertList(t, ctx, s.List, store.ListOptions{NumberOfResults: 3, Offset: 3}, 1, []string{
		"/wibble",
	})

	// ask for offset bigger than the number of things
	assertList(t, ctx, s.List, store.ListOptions{Offset: 10}, 0, []string{})
}

// the rest of the List* functions are implemented using list by type
// ListByType should include child types

func listByType(s store.Store, typ *store.Type) func(context.Context, store.ListOptions) ([]*store.Thing, error) {
	return func(ctx context.Context, opts store.ListOptions) ([]*store.Thing, error) {
		return s.ListByType(ctx, typ, opts)
	}
}

func TestListByType(t *testing.T, s store.Store) {
	ctx := context.Background()

	if err := s.Add(ctx,
		entity("/wibble/bibble/1"),
		entity("/wibble/bibble/5"),
		entity("/wibble/bibble/4"),
		relation("/wibble/bibble/3"),
		relation("/wibble/bibble/6"),
	); err != nil {
		t.Fatalf("Couldn't add to store: %v", err)
	}

	// ENTITIES

	listEntities := listByType(s, store.EntityType)

	// default sort order is ascending
	assertList(t, ctx, listEntities, store.ListOptions{}, 3, []string{
		"/wibble/bibble/1",
		"/wibble/bibble/4",
		"/wibble/bibble/5",
	})

	// test descending correctly works
	assertList(t, ctx, listEntities, store.ListOptions{SortOrder: store.SortDescending}, 3, []string{
		"/wibble/bibble/5",
		"/wibble/bibble/4",
		"/wibble/bibble/1",
	})

	// ask for too many results
	assertList(t, ctx, listEntities, store.ListOptions{NumberOfResults: 1000}, 3, []string{
		"/wibble/bibble/1",
		"/wibble/bibble/4",
		"/wibble/bibble/5",
	})

	// ask for less results than are in the store
	assertList(t, ctx, listEntities, store.ListOptions{NumberOfResults: 2}, 2, []string{
		"/wibble/bibble/1",
		"/wibble/bibble/4",
	})

	// ask for offset bigger than the number of things
	assertList(t, ctx, listEntities, store.ListOptions{Offset: 10}, 0, []string{})

	compareLists(t, ctx, listEntities, convertEntitiesToThings(s.ListEntities), store.ListOptions{})
	compareLists(t, ctx, listEntities, convertEntitiesToThings(s.ListEntities), store.ListOptions{SortOrder: store.SortDescending})
	compareLists(t, ctx, listEntities, convertEntitiesToThings(s.ListEntities), store.ListOptions{NumberOfResults: 1000})
	compareLists(t, ctx, listEntities, convertEntitiesToThings(s.ListEntities), store.ListOptions{NumberOfResults: 1})
	compareLists(t, ctx, listEntities, convertEntitiesToThings(s.ListEntities), store.ListOptions{Offset: 10})

	// RELATIONS

	listRelations := listByType(s, store.RelationType)

	// default sort order is ascending
	assertList(t, ctx, listRelations, store.ListOptions{}, 2, []string{
		"/wibble/bibble/3",
		"/wibble/bibble/6",
	})

	// test descending correctly works
	assertList(t, ctx, listRelations, store.ListOptions{SortOrder: store.SortDescending}, 2, []string{
		"/wibble/bibble/6",
		"/wibble/bibble/3",
	})

	// ask for too many results
	assertList(t, ctx, listRelations, store.ListOptions{NumberOfResults: 1000}, 2, []string{
		"/wibble/bibble/3",
		"/wibble/bibble/6",
	})

	// ask for less results than are in the store
	assertList(t, ctx, listRelations, store.ListOptions{NumberOfResults: 1}, 1, []string{
		"/wibble/bibble/3",
	})

	// ask for offset bigger than the number of things
	assertList(t, ctx, listRelations, store.ListOptions{Offset: 10}, 0, []string{})

	compareLists(t, ctx, listRelations, convertRelationsToThings(s.ListRelations), store.ListOptions{})
	compareLists(t, ctx, listRelations, convertRelationsToThings(s.ListRelations), store.ListOptions{SortOrder: store.SortDescending})
	compareLists(t, ctx, listRelations, convertRelationsToThings(s.ListRelations), store.ListOptions{NumberOfResults: 1000})
	compareLists(t, ctx, listRelations, convertRelationsToThings(s.ListRelations), store.ListOptions{NumberOfResults: 1})
	compareLists(t, ctx, listRelations, convertRelationsToThings(s.ListRelations), store.ListOptions{Offset: 10})

	// TYPES

	listTypes := listByType(s, store.TypeType)

	// default sort order is ascending
	assertList(t, ctx, listTypes, store.ListOptions{}, 3, []string{
		"/entity",
		"/relation",
		"/type",
	})

	// test descending correctly works
	assertList(t, ctx, listTypes, store.ListOptions{SortOrder: store.SortDescending}, 3, []string{
		"/type",
		"/relation",
		"/entity",
	})

	// ask for too many results
	assertList(t, ctx, listTypes, store.ListOptions{NumberOfResults: 1000}, 3, []string{
		"/entity",
		"/relation",
		"/type",
	})

	// ask for less results than are in the store
	assertList(t, ctx, listTypes, store.ListOptions{NumberOfResults: 1}, 1, []string{
		"/entity",
	})

	// ask for offset bigger than the number of things
	assertList(t, ctx, listTypes, store.ListOptions{Offset: 10}, 0, []string{})

	compareLists(t, ctx, listTypes, convertTypesToThings(s.ListTypes), store.ListOptions{})
	compareLists(t, ctx, listTypes, convertTypesToThings(s.ListTypes), store.ListOptions{SortOrder: store.SortDescending})
	compareLists(t, ctx, listTypes, convertTypesToThings(s.ListTypes), store.ListOptions{NumberOfResults: 1000})
	compareLists(t, ctx, listTypes, convertTypesToThings(s.ListTypes), store.ListOptions{NumberOfResults: 1})
	compareLists(t, ctx, listTypes, convertTypesToThings(s.ListTypes), store.ListOptions{Offset: 10})
}

func appendFromChannel(ch chan *store.Thing, list *[]*store.Thing, wg *sync.WaitGroup) {
	for {
		if thing := <-ch; thing == nil {
			wg.Done()
			return
		} else {
			*list = append(*list, thing)
		}
	}
}

func TestWatchByType(t *testing.T, s store.Store) {
	closedWG := sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())

	closedWG.Add(3)

	typesList1 := []*store.Thing{}
	if ch, err := s.WatchByType(ctx, store.TypeType); err != nil {
		t.Fatalf("Couldn't watch type: %v", err)
	} else {
		go appendFromChannel(ch, &typesList1, &closedWG)
	}

	typesList2 := []*store.Thing{}
	if ch, err := s.WatchByType(ctx, store.TypeType); err != nil {
		t.Fatalf("Couldn't watch type: %v", err)
	} else {
		go appendFromChannel(ch, &typesList2, &closedWG)
	}

	entitiesList1 := []*store.Thing{}
	if ch, err := s.WatchByType(ctx, store.EntityType); err != nil {
		t.Fatalf("Couldn't watch entity: %v", err)
	} else {
		go appendFromChannel(ch, &entitiesList1, &closedWG)
	}

	// do some changes
	if err := s.Add(
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
	if err := s.Add(
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

func TestWatchByID(t *testing.T, s store.Store) {
	closedWG := sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())

	closedWG.Add(2)

	wibbleChanges := []*store.Thing{}
	if ch, err := s.WatchByID(ctx, store.ID("/wibble")); err != nil {
		t.Fatalf("Couldn't watch type: %v", err)
	} else {
		go appendFromChannel(ch, &wibbleChanges, &closedWG)
	}

	nibbleChanges := []*store.Thing{}
	if ch, err := s.WatchByID(ctx, store.ID("/nibble")); err != nil {
		t.Fatalf("Couldn't watch type: %v", err)
	} else {
		go appendFromChannel(ch, &nibbleChanges, &closedWG)
	}

	// do some changes
	if err := s.Add(
		ctx,
		ntype("/wibble"),
		ntype("/bibble"),
	); err != nil {
		t.Fatalf("Couldn't add the bits: %v", err)
	}

	// cancel the context
	cancel()
	closedWG.Wait()

	// check the expected numbers
	if expected := 1; len(wibbleChanges) != expected {
		t.Errorf("Expected there to be %d /wibble changes, but it was %d", expected, len(wibbleChanges))
	}
	if expected := 0; len(nibbleChanges) != expected {
		t.Errorf("Expected there to be %d /nibble changes, but it was %d", expected, len(nibbleChanges))
	}

	// do some changes
	if err := s.Add(
		ctx,
		ntype("/wibble"),
		ntype("/bibble"),
	); err != nil {
		t.Fatalf("Couldn't add the bits: %v", err)
	}

	// check the numbers haven't changed
	if expected := 1; len(wibbleChanges) != expected {
		t.Errorf("Expected there to be %d /wibble changes, but it was %d", expected, len(wibbleChanges))
	}
	if expected := 0; len(nibbleChanges) != expected {
		t.Errorf("Expected there to be %d /nibble changes, but it was %d", expected, len(nibbleChanges))
	}

}

func TestTypeProperties(t *testing.T, s store.Store) {
	ctx := context.Background()

	relType := typ("/relation/wibble", "/relation", map[string]interface{}{
		"a": map[string]interface{}{
			"type":       "string",
			"pointer_to": "/entity/thing",
		},
		"bibble": map[string]interface{}{
			"type": "object",
		},
	})

	if err := s.Add(ctx, relType.Thing()); err != nil {
		t.Fatalf("Couldn't add to store: %v", err)
	}

	props, requiredProps, err := store.TypeProperties(ctx, s, relType)
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

	if pointerTo := props["a"].Validators["pointer_to"].(*store.PointerTo); pointerTo.String() != "/entity/thing" {
		t.Errorf("Expected pointer_to for a to be '/entity/thing', but it was '%s'", pointerTo.String())
	}
	if pointerTo := props["b"].Validators["pointer_to"].(*store.PointerTo); pointerTo.String() != "/entity" {
		t.Errorf("Expected pointer_to for b to be '/entity', but it was '%s'", pointerTo.String())
	}
}

func TestValidate(t *testing.T, s store.Store) {
	ctx := context.Background()

	relType := typ("/relation/wibble", "/relation", map[string]interface{}{
		"a": map[string]interface{}{
			"type":       "string",
			"pointer_to": "/entity/thing",
		},
		"bibble": map[string]interface{}{
			"type": "object",
		},
	})

	entThingType := typ("/entity/thing", "/entity", map[string]interface{}{})

	ent1 := entityWithType("/asdf", "/entity")
	ent2 := entityWithType("/sdfg", "/entity/thing")

	if err := s.Add(ctx, entThingType.Thing(), relType.Thing(), ent1, ent2); err != nil {
		t.Fatalf("Couldn't add to store: %v", err)
	}

	validRel := relationBetweenWithType("/qwer", "/relation/wibble", "/sdfg", "/asdf")
	if valErrs, err := s.Validate(ctx, validRel, store.ValidateOptions{}); err != nil {
		t.Errorf("Failed to validate thing: %v", err)
	} else if len(valErrs) != 0 {
		t.Errorf("Expected 0 validation errors, got %d: %v", len(valErrs), valErrs)
	}

	wrongwayroundRel := relationBetweenWithType("/qwer", "/relation/wibble", "/asdf", "/sdfg")
	if valErrs, err := s.Validate(ctx, wrongwayroundRel, store.ValidateOptions{}); err != nil {
		t.Errorf("Failed to validate thing: %v", err)
	} else if len(valErrs) != 1 {
		t.Errorf("Expected 1 validation error, got %d: %v", len(valErrs), valErrs)
	}
	if valErrs, err := s.Validate(ctx, wrongwayroundRel, store.ValidateOptions{Pointers: store.IgnoreMissingPointers}); err != nil {
		t.Errorf("Failed to validate thing: %v", err)
	} else if len(valErrs) != 1 {
		t.Errorf("Expected 1 validation error, got %d: %v", len(valErrs), valErrs)
	}
	if valErrs, err := s.Validate(ctx, wrongwayroundRel, store.ValidateOptions{Pointers: store.IgnoreAllPointers}); err != nil {
		t.Errorf("Failed to validate thing: %v", err)
	} else if len(valErrs) != 0 {
		t.Errorf("Expected 0 validation errors, got %d: %v", len(valErrs), valErrs)
	}

	invalidRel := thingWithType("/wert", "/relation/wibble", store.Properties{})
	if valErrs, err := s.Validate(ctx, invalidRel, store.ValidateOptions{}); err != nil {
		t.Errorf("Failed to validate thing: %v", err)
	} else if len(valErrs) == 0 {
		t.Errorf("Expected more than 0 validation errors, got %d: %v", len(valErrs), valErrs)
	}

	missingRel := relationBetweenWithType("/qwer", "/relation/wibble", "/unknown-entity", "/asdf")
	if valErrs, err := s.Validate(ctx, missingRel, store.ValidateOptions{}); err != nil {
		t.Errorf("Failed to validate thing: %v", err)
	} else if len(valErrs) != 1 {
		t.Errorf("Expected 1 validation error, got %d: %v", len(valErrs), valErrs)
	}
	if valErrs, err := s.Validate(ctx, missingRel, store.ValidateOptions{Pointers: store.IgnoreMissingPointers}); err != nil {
		t.Errorf("Failed to validate thing: %v", err)
	} else if len(valErrs) != 0 {
		t.Errorf("Expected 0 validation errors, got %d: %v", len(valErrs), valErrs)
	}

}
