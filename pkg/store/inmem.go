package store

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

type inmemStore struct {
	things map[ID]*Thing

	rw sync.RWMutex
}

func NewInMemoryStore() Store {
	store := &inmemStore{
		things: map[ID]*Thing{},
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

func (s *inmemStore) Add(thing *Thing) error {
	s.rw.Lock()
	defer s.rw.Unlock()

	s.things[thing.ID] = thing

	return nil
}

func (s *inmemStore) AddAll(things []*Thing) error {
	s.rw.Lock()
	defer s.rw.Unlock()

	for _, thing := range things {
		s.things[thing.ID] = thing
	}

	return nil
}

func (s *inmemStore) IsA(thing *Thing, t *Type) (bool, error) {
	if t == TypeType {
		return thing.Metadata.Type == t.Metadata.ID, nil
	}

	thingTypeID := thing.Metadata.Type

	for {
		if thingTypeID == t.Metadata.ID {
			return true, nil
		}

		thingType, err := s.GetTypeByID(thingTypeID)
		if err != nil {
			return false, err
		}

		if parent, ok := thingType.Properties["parent"]; !ok {
			break
		} else if parentString, ok := parent.(string); !ok {
			return false, fmt.Errorf("%v should be a string", parent)
		} else {
			thingTypeID = ID(parentString)
		}
	}

	return false, nil
}

func (s *inmemStore) GetByID(id ID) (*Thing, error) {
	s.rw.RLock()
	defer s.rw.RUnlock()

	if thing, ok := s.things[id]; !ok {
		return nil, ErrNotFound
	} else {
		return thing, nil
	}
}

func (s *inmemStore) GetEntityByID(id ID) (*Entity, error) {
	s.rw.RLock()
	defer s.rw.RUnlock()

	if thing, ok := s.things[id]; !ok {
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
	defer s.rw.RUnlock()

	if thing, ok := s.things[id]; !ok {
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
	defer s.rw.RUnlock()

	if thing, ok := s.things[id]; !ok {
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

func (s *inmemStore) ListByType(typ *Type, options ListOptions) ([]*Thing, error) {

	// extract a list of things
	things := []*Thing{}

	{
		s.rw.RLock()
		defer s.rw.RUnlock()

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
	}

	// sort that list of things

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
		options.NumberOfResults = 10
	}

	size := min(len(things), int(options.NumberOfResults))

	out := make([]*Thing, size)
	for idx, _ := range out {
		out[idx] = things[idx]
	}

	return out, nil
}
