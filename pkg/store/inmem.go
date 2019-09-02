package store

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
)

type inmemStore struct {
	things map[ID]*Thing
	broadcast *broadcast

	rw sync.RWMutex
}

func NewInMemoryStore() Store {
	store := &inmemStore{
		things: map[ID]*Thing{},
		broadcast: newBroadcast(),
	}

	ctx := context.TODO()

	store.Add(ctx, TypeType.Thing())
	store.Add(ctx, EntityType.Thing())
	store.Add(ctx, RelationType.Thing())

	return store
}

func (s *inmemStore) Len(_ context.Context) (int, error) {
	s.rw.Lock()
	defer s.rw.Unlock()

	return len(s.things), nil
}

func (s *inmemStore) Add(ctx context.Context, things ...Thingable) error {
	return s.AddAll(ctx, things)
}

func (s *inmemStore) AddAll(ctx context.Context, things []Thingable) error {
	s.rw.Lock()
	for _, thingable := range things {
		thing := thingable.Thing()
		s.things[thing.ID] = thing
	}
	s.rw.Unlock()

	for _, thingable := range things {
		thing := thingable.Thing()

		// don't broadcast if we can't get types
		// we don't want to enforce validation, but we can't broadcast
		// if we don't know the types
		if ! thing.Equal(TypeType, EntityType, RelationType) {
			if types, err := s.Types(ctx, thingable); err == nil {
				typeIDs := make([]ID, len(types))
				for idx, typ := range types {
					typeIDs[idx] = typ.Metadata.ID
				}

				s.broadcast.Send(ctx, thingable, typeIDs...)
			}
		}
	}

	return nil
}

func (s *inmemStore) Types(ctx context.Context, thingable Thingable) ([]*Type, error) {
	thing := thingable.Thing()

	types := []*Type{}
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
			thingTypeID = ID(parentString)
		}
	}

	return types, nil
}

func (s *inmemStore) TypeHierarchy(ctx context.Context, typ *Type) ([]*Type, error) {
	types := []*Type{}
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
			thingTypeID = ID(parentString)
		}
	}

	return types, nil
}

func (s *inmemStore) IsA(ctx context.Context, thingable Thingable, t *Type) (bool, error) {
	if t == TypeType {
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

func (s *inmemStore) Validate(ctx context.Context, t Thingable, opts ValidateOptions) ([]ValidationError, error) {
	return validate(ctx, s, t, opts)
}

func (s *inmemStore) GetByID(ctx context.Context, id ID) (*Thing, error) {
	s.rw.RLock()
	thing, ok := s.things[id]
	s.rw.RUnlock()

	if !ok {
		return nil, ErrNotFound
	} else {
		return thing, nil
	}
}

func (s *inmemStore) GetEntityByID(ctx context.Context, id ID) (*Entity, error) {
	s.rw.RLock()
	thing, ok := s.things[id]
	s.rw.RUnlock()

	if !ok {
		return nil, ErrNotFound
	} else if ok, err := s.IsA(ctx, thing, EntityType); !ok {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	} else {
		return (*Entity)(thing), nil
	}
}

func (s *inmemStore) GetRelationByID(ctx context.Context, id ID) (*Relation, error) {
	s.rw.RLock()
	thing, ok := s.things[id]
	s.rw.RUnlock()

	if !ok {
		return nil, ErrNotFound
	} else if ok, err := s.IsA(ctx, thing, RelationType); !ok {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	} else {
		return (*Relation)(thing), nil
	}
}

func (s *inmemStore) GetTypeByID(ctx context.Context, id ID) (*Type, error) {
	s.rw.RLock()
	thing, ok := s.things[id]
	s.rw.RUnlock()

	if !ok {
		return nil, ErrNotFound
	} else if ok, err := s.IsA(ctx, thing, TypeType); !ok {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	} else {
		return (*Type)(thing), nil
	}
}

func (s *inmemStore) List(ctx context.Context, options ListOptions) ([]*Thing, error) {
	return s.ListByType(ctx, nil, options)
}

func (s *inmemStore) ListEntities(ctx context.Context, options ListOptions) ([]*Entity, error) {
	things, err := s.ListByType(ctx, EntityType, options)
	entities := make([]*Entity, len(things))

	if err != nil {
		return entities, nil
	}

	for idx, thing := range things {
		entities[idx] = (*Entity)(thing)
	}

	return entities, nil
}

func (s *inmemStore) ListRelations(ctx context.Context, options ListOptions) ([]*Relation, error) {
	things, err := s.ListByType(ctx, RelationType, options)
	relations := make([]*Relation, len(things))

	if err != nil {
		return relations, nil
	}

	for idx, thing := range things {
		relations[idx] = (*Relation)(thing)
	}

	return relations, nil
}

func (s *inmemStore) ListTypes(ctx context.Context, options ListOptions) ([]*Type, error) {
	things, err := s.ListByType(ctx, TypeType, options)
	types := make([]*Type, len(things))

	if err != nil {
		return types, nil
	}

	for idx, thing := range things {
		types[idx] = (*Type)(thing)
	}

	return types, nil
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func (s *inmemStore) listAllByType(ctx context.Context, typ *Type) ([]*Thing, error) {
	s.rw.RLock()
	defer s.rw.RUnlock()

	things := []*Thing{}

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

func (s *inmemStore) constrainList(things []*Thing, options ListOptions) ([]*Thing, error) {
	if options.SortField != SortByID {
		return []*Thing{}, ErrUnimplemented
	}

	var sortFunc func(int, int) bool

	switch options.SortOrder {
	case SortAscending:
		sortFunc = func(i, j int) bool {
			return strings.Compare(string(things[i].Metadata.ID), string(things[j].Metadata.ID)) < 0
		}
	case SortDescending:
		sortFunc = func(i, j int) bool {
			return strings.Compare(string(things[i].Metadata.ID), string(things[j].Metadata.ID)) > 0
		}
	default:
		return []*Thing{}, ErrUnimplemented
	}

	sort.Slice(things, sortFunc)

	// do the sizing and offsetting

	if int(options.Offset) > len(things) {
		return []*Thing{}, nil
	}

	if options.NumberOfResults == 0 {
		options.NumberOfResults = DefaultNumberOfResults
	}

	size := min(len(things) - int(options.Offset), int(options.NumberOfResults))

	out := make([]*Thing, size)
	for idx, _ := range out {
		out[idx] = things[int(options.Offset) + idx]
	}

	return out, nil
}

func (s *inmemStore) ListByType(ctx context.Context, typ *Type, options ListOptions) ([]*Thing, error) {
	things, err := s.listAllByType(ctx, typ)
	if err != nil {
		return nil, err
	}

	return s.constrainList(things, options)
}

func (s *inmemStore) ListRelationsForEntity(ctx context.Context, entity *Entity, options ListOptions) ([]*Relation, error) {
	allRelations, err := s.listAllByType(ctx, RelationType)
	if err != nil {
		return nil, err
	}

	involvedRelations := []*Thing{}

	for _, relationThing := range allRelations {
		relation := (*Relation)(relationThing)
		if relation.Involves(entity) {
			involvedRelations = append(involvedRelations, relation.Thing())
		}
	}

	constrainedThings, err := s.constrainList(involvedRelations, options)
	if err != nil {
		return nil, err
	}

	relations := make([]*Relation, len(constrainedThings))
	for idx, thing := range constrainedThings {
		relations[idx] = (*Relation)(thing)
	}

	return relations, nil
}

func (s *inmemStore) WatchByType(ctx context.Context, typ *Type) (chan *Thing, error) {
	return s.broadcast.Register(ctx, typ.Metadata.ID)
}
