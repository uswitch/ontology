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

	store.Add(TypeType.Thing())
	store.Add(EntityType.Thing())
	store.Add(RelationType.Thing())

	return store
}

func (s *inmemStore) Len() (int, error) {
	s.rw.Lock()
	defer s.rw.Unlock()

	return len(s.things), nil
}

func (s *inmemStore) Add(things ...Thingable) error {
	return s.AddAll(things)
}

func (s *inmemStore) AddAll(things []Thingable) error {
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
			if types, err := s.typeHierarchy(thingable); err == nil {
				s.broadcast.Send(context.TODO(), thingable, types...)
			}
		}
	}

	return nil
}

func (s *inmemStore) typeHierarchy(thingable Thingable) ([]ID, error) {
	thing := thingable.Thing()

	types := []ID{}
	thingTypeID := thing.Metadata.Type

	for {
		thingType, err := s.GetTypeByID(thingTypeID)
		if err != nil {
			return nil, err
		}

		types = append(types, thingTypeID)

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

func (s *inmemStore) IsA(thingable Thingable, t *Type) (bool, error) {
	if t == TypeType {
		return thingable.Thing().Metadata.Type == t.Metadata.ID, nil
	}

	types, err := s.typeHierarchy(thingable)
	if err != nil {
		return false, err
	}

	for _, typeID := range types {
		if typeID == t.Metadata.ID {
			return true, nil
		}
	}

	return false, nil
}

func (s *inmemStore) Validate(t Thingable, opts ValidateOptions) ([]ValidationError, error) {
	return validate(s, t, opts)
}

func (s *inmemStore) GetByID(id ID) (*Thing, error) {
	s.rw.RLock()
	thing, ok := s.things[id]
	s.rw.RUnlock()

	if !ok {
		return nil, ErrNotFound
	} else {
		return thing, nil
	}
}

func (s *inmemStore) GetEntityByID(id ID) (*Entity, error) {
	s.rw.RLock()
	thing, ok := s.things[id]
	s.rw.RUnlock()

	if !ok {
		return nil, ErrNotFound
	} else if ok, err := s.IsA(thing, EntityType); !ok {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	} else {
		return (*Entity)(thing), nil
	}
}

func (s *inmemStore) GetRelationByID(id ID) (*Relation, error) {
	s.rw.RLock()
	thing, ok := s.things[id]
	s.rw.RUnlock()

	if !ok {
		return nil, ErrNotFound
	} else if ok, err := s.IsA(thing, RelationType); !ok {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	} else {
		return (*Relation)(thing), nil
	}
}

func (s *inmemStore) GetTypeByID(id ID) (*Type, error) {
	s.rw.RLock()
	thing, ok := s.things[id]
	s.rw.RUnlock()

	if !ok {
		return nil, ErrNotFound
	} else if ok, err := s.IsA(thing, TypeType); !ok {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	} else {
		return (*Type)(thing), nil
	}
}

func (s *inmemStore) List(options ListOptions) ([]*Thing, error) {
	return s.ListByType(nil, options)
}

func (s *inmemStore) ListEntities(options ListOptions) ([]*Entity, error) {
	things, err := s.ListByType(EntityType, options)
	entities := make([]*Entity, len(things))

	if err != nil {
		return entities, nil
	}

	for idx, thing := range things {
		entities[idx] = (*Entity)(thing)
	}

	return entities, nil
}

func (s *inmemStore) ListRelations(options ListOptions) ([]*Relation, error) {
	things, err := s.ListByType(RelationType, options)
	relations := make([]*Relation, len(things))

	if err != nil {
		return relations, nil
	}

	for idx, thing := range things {
		relations[idx] = (*Relation)(thing)
	}

	return relations, nil
}

func (s *inmemStore) ListTypes(options ListOptions) ([]*Type, error) {
	things, err := s.ListByType(TypeType, options)
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

func (s *inmemStore) listAllByType(typ *Type) ([]*Thing, error) {
	s.rw.RLock()
	defer s.rw.RUnlock()

	things := []*Thing{}

	for _, thing := range s.things {
		if typ != nil {
			if ok, err := s.IsA(thing, typ); !ok {
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

func (s *inmemStore) ListByType(typ *Type, options ListOptions) ([]*Thing, error) {
	things, err := s.listAllByType(typ)
	if err != nil {
		return nil, err
	}

	return s.constrainList(things, options)
}

func (s *inmemStore) ListRelationsForEntity(entity *Entity, options ListOptions) ([]*Relation, error) {
	allRelations, err := s.listAllByType(RelationType)
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
