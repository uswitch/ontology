package inmem

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/uswitch/ontology/pkg/store"
)

type inmemStore struct {
	things        map[store.ID]*store.Thing
	typeBroadcast *store.Broadcast
	idBroadcast   *store.Broadcast

	rw sync.RWMutex
}

func NewInMemoryStore() store.Store {
	s := &inmemStore{
		things:        map[store.ID]*store.Thing{},
		typeBroadcast: store.NewBroadcast(),
		idBroadcast:   store.NewBroadcast(),
	}

	ctx := context.TODO()

	s.Add(
		ctx,
		store.TypeType.Thing(),
		store.EntityType.Thing(),
		store.RelationType.Thing(),
		store.TypeOfType.Thing(),
		store.SubtypeOfType.Thing(),
	)

	return s
}

func (s *inmemStore) Len(_ context.Context) (int, error) {
	s.rw.Lock()
	defer s.rw.Unlock()

	return len(s.things), nil
}

func (s *inmemStore) Add(ctx context.Context, things ...store.Thingable) error {
	return s.AddAll(ctx, things)
}

func (s *inmemStore) AddAll(ctx context.Context, things []store.Thingable) error {
	s.rw.Lock()
	for _, thingable := range things {
		thing := thingable.Thing()
		s.things[thing.ID()] = thing
	}
	s.rw.Unlock()

	for _, thingable := range things {
		thing := thingable.Thing()

		s.idBroadcast.Send(ctx, thingable, thing.ID())

		// don't broadcast if we can't get types
		// we don't want to enforce validation, but we can't broadcast
		// if we don't know the types
		if !thing.Equal(store.TypeType, store.EntityType, store.RelationType) {
			if types, err := s.Types(ctx, thingable); err == nil {
				typeIDs := make([]store.ID, len(types))
				for idx, typ := range types {
					typeIDs[idx] = typ.Metadata.ID
				}

				s.typeBroadcast.Send(ctx, thingable, typeIDs...)
			}
		}
	}

	return nil
}

func (s *inmemStore) Types(ctx context.Context, thingable store.Thingable) ([]*store.Type, error) {
	thing := thingable.Thing()

	types := []*store.Type{}
	thingTypeID := thing.Metadata.Type

	for {
		thingType, err := s.GetTypeByID(ctx, thingTypeID)
		if err != nil {
			return nil, err
		}

		types = append(types, thingType)

		if parent, ok := thingType.Properties["parent"]; !ok {
			break
		} else if parentString, ok := parent.(string); !ok {
			return nil, fmt.Errorf("%v should be a string", parent)
		} else {
			thingTypeID = store.ID(parentString)
		}
	}

	return types, nil
}

func (s *inmemStore) TypeHierarchy(ctx context.Context, typ *store.Type) ([]*store.Type, error) {
	types := []*store.Type{}
	thingTypeID := typ.Metadata.ID

	for {
		thingType, err := s.GetTypeByID(ctx, thingTypeID)
		if err != nil {
			return nil, err
		}

		types = append(types, thingType)

		if parent, ok := thingType.Properties["parent"]; !ok {
			break
		} else if parentString, ok := parent.(string); !ok {
			return nil, fmt.Errorf("%v should be a string", parent)
		} else {
			thingTypeID = store.ID(parentString)
		}
	}

	return types, nil
}

func (s *inmemStore) Inherits(ctx context.Context, typ *store.Type, parent *store.Type) (bool, error) {
	typeHierarchy, err := s.TypeHierarchy(ctx, typ)
	if err != nil {
		return false, err
	}

	isInheritted := false

	for _, t := range typeHierarchy {
		if t.Thing().Equal(parent) {
			isInheritted = true
			break
		}
	}

	return isInheritted, nil
}

func (s *inmemStore) IsA(ctx context.Context, thingable store.Thingable, t *store.Type) (bool, error) {
	if t == store.TypeType {
		return thingable.Thing().Metadata.Type == t.Metadata.ID, nil
	}

	types, err := s.Types(ctx, thingable)
	if err != nil {
		return false, err
	}

	for _, typ := range types {
		if typ.Metadata.ID == t.Metadata.ID {
			return true, nil
		}
	}

	return false, nil
}

func (s *inmemStore) Validate(ctx context.Context, t store.Thingable, opts store.ValidateOptions) ([]store.ValidationError, error) {
	return store.Validate(ctx, s, t, opts)
}

func (s *inmemStore) GetByID(ctx context.Context, idable store.IDable) (*store.Thing, error) {
	id := idable.ID()

	s.rw.RLock()
	thing, ok := s.things[id]
	s.rw.RUnlock()

	if !ok {
		return nil, store.ErrNotFound
	} else {
		return thing, nil
	}
}

func (s *inmemStore) GetEntityByID(ctx context.Context, idable store.IDable) (*store.Entity, error) {
	id := idable.ID()

	s.rw.RLock()
	thing, ok := s.things[id]
	s.rw.RUnlock()

	if !ok {
		return nil, store.ErrNotFound
	} else if ok, err := s.IsA(ctx, thing, store.EntityType); !ok {
		return nil, store.ErrNotFound
	} else if err != nil {
		return nil, err
	} else {
		return (*store.Entity)(thing), nil
	}
}

func (s *inmemStore) GetRelationByID(ctx context.Context, idable store.IDable) (*store.Relation, error) {
	id := idable.ID()

	s.rw.RLock()
	thing, ok := s.things[id]
	s.rw.RUnlock()

	if !ok {
		return nil, store.ErrNotFound
	} else if ok, err := s.IsA(ctx, thing, store.RelationType); !ok {
		return nil, store.ErrNotFound
	} else if err != nil {
		return nil, err
	} else {
		return (*store.Relation)(thing), nil
	}
}

func (s *inmemStore) GetTypeByID(ctx context.Context, idable store.IDable) (*store.Type, error) {
	id := idable.ID()

	s.rw.RLock()
	thing, ok := s.things[id]
	s.rw.RUnlock()

	if !ok {
		return nil, store.ErrNotFound
	} else if ok, err := s.IsA(ctx, thing, store.TypeType); !ok {
		return nil, store.ErrNotFound
	} else if err != nil {
		return nil, err
	} else {
		return (*store.Type)(thing), nil
	}
}

func (s *inmemStore) List(ctx context.Context, options store.ListOptions) ([]*store.Thing, error) {
	return s.ListByType(ctx, nil, options)
}

func (s *inmemStore) ListEntities(ctx context.Context, options store.ListOptions) ([]*store.Entity, error) {
	things, err := s.ListByType(ctx, store.EntityType, options)
	entities := make([]*store.Entity, len(things))

	if err != nil {
		return entities, nil
	}

	for idx, thing := range things {
		entities[idx] = (*store.Entity)(thing)
	}

	return entities, nil
}

func (s *inmemStore) ListRelations(ctx context.Context, options store.ListOptions) ([]*store.Relation, error) {
	things, err := s.ListByType(ctx, store.RelationType, options)
	relations := make([]*store.Relation, len(things))

	if err != nil {
		return relations, nil
	}

	for idx, thing := range things {
		relations[idx] = (*store.Relation)(thing)
	}

	return relations, nil
}

func (s *inmemStore) ListTypes(ctx context.Context, options store.ListOptions) ([]*store.Type, error) {
	things, err := s.ListByType(ctx, store.TypeType, options)
	types := make([]*store.Type, len(things))

	if err != nil {
		return types, nil
	}

	for idx, thing := range things {
		types[idx] = (*store.Type)(thing)
	}

	return types, nil
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func (s *inmemStore) listAllByType(ctx context.Context, typ *store.Type) ([]*store.Thing, error) {
	s.rw.RLock()
	defer s.rw.RUnlock()

	things := []*store.Thing{}

	for _, thing := range s.things {
		if typ != nil {
			if ok, err := s.IsA(ctx, thing, typ); !ok {
				continue
			} else if err != nil {
				return things, err
			}
		}

		things = append(things, thing)
	}

	return things, nil
}

func (s *inmemStore) constrainList(things []*store.Thing, options store.ListOptions) ([]*store.Thing, error) {
	if options.SortField != store.SortByID {
		return []*store.Thing{}, store.ErrUnimplemented
	}

	var sortFunc func(int, int) bool

	switch options.SortOrder {
	case store.SortAscending:
		sortFunc = func(i, j int) bool {
			return strings.Compare(string(things[i].Metadata.ID), string(things[j].Metadata.ID)) < 0
		}
	case store.SortDescending:
		sortFunc = func(i, j int) bool {
			return strings.Compare(string(things[i].Metadata.ID), string(things[j].Metadata.ID)) > 0
		}
	default:
		return []*store.Thing{}, store.ErrUnimplemented
	}

	sort.Slice(things, sortFunc)

	// do the sizing and offsetting

	if int(options.Offset) > len(things) {
		return []*store.Thing{}, nil
	}

	if options.NumberOfResults == 0 {
		options.NumberOfResults = store.DefaultNumberOfResults
	}

	size := min(len(things)-int(options.Offset), int(options.NumberOfResults))

	out := make([]*store.Thing, size)
	for idx, _ := range out {
		out[idx] = things[int(options.Offset)+idx]
	}

	return out, nil
}

func (s *inmemStore) ListByType(ctx context.Context, typ *store.Type, options store.ListOptions) ([]*store.Thing, error) {
	things, err := s.listAllByType(ctx, typ)
	if err != nil {
		return nil, err
	}

	return s.constrainList(things, options)
}

func (s *inmemStore) ListRelationsForEntity(ctx context.Context, relConstraint *store.Type, entity *store.Entity, options store.ListOptions) ([]*store.Relation, error) {

	relType := store.RelationType

	if relConstraint != nil {
		if isRelation, err := s.Inherits(ctx, relConstraint, store.RelationType); err != nil {
			return nil, err
		} else if !isRelation {
			return nil, fmt.Errorf("%v is not a relation", relConstraint)
		}

		relType = relConstraint
	}

	allRelations, err := s.listAllByType(ctx, relType)
	if err != nil {
		return nil, err
	}

	involvedRelations := []*store.Thing{}

	for _, relationThing := range allRelations {
		relation := (*store.Relation)(relationThing)
		if relation.Involves(entity) {
			involvedRelations = append(involvedRelations, relation.Thing())
		}
	}

	constrainedThings, err := s.constrainList(involvedRelations, options)
	if err != nil {
		return nil, err
	}

	relations := make([]*store.Relation, len(constrainedThings))
	for idx, thing := range constrainedThings {
		relations[idx] = (*store.Relation)(thing)
	}

	return relations, nil
}

func (s *inmemStore) WatchByType(ctx context.Context, idable store.IDable) (chan *store.Thing, error) {
	return s.typeBroadcast.Register(ctx, idable.ID())
}

func (s *inmemStore) WatchByID(ctx context.Context, idable store.IDable) (chan *store.Thing, error) {
	return s.idBroadcast.Register(ctx, idable.ID())
}
