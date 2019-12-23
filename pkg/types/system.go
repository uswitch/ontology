package types

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"
)

type ID string

func (id ID) String() string { return string(id) }

type Metadata struct {
	ID        ID        `json:"id"`
	Type      ID        `json:"type"`
	Name      string    `json:"name"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Instance interface {
	ID() ID
	Type() ID
	Name() string
	UpdatedAt() time.Time
}

type Any struct {
	Metadata   `json:"metadata"`
	Properties struct{} `json:"properties"`
}

func (i *Any) ID() ID               { return i.Metadata.ID }
func (i *Any) Type() ID             { return i.Metadata.Type }
func (i *Any) Name() string         { return i.Metadata.Name }
func (i *Any) UpdatedAt() time.Time { return i.Metadata.UpdatedAt }

type System struct {
	types    map[string]reflect.Type
	typeIDs  map[reflect.Type]string
	parent   map[string]string
	children map[string][]string
}

func NewSystem() *System {
	return &System{
		types:    map[string]reflect.Type{},
		typeIDs:  map[reflect.Type]string{},
		parent:   map[string]string{},
		children: map[string][]string{},
	}
}

func (s *System) RegisterType(instance interface{}, id string, parentID string) {
	typ := reflect.TypeOf(instance)
	s.types[id] = typ
	s.typeIDs[typ] = id

	if parentID != "" {
		s.parent[id] = parentID
		if _, ok := s.children[parentID]; !ok {
			s.children[parentID] = []string{}
		}
		s.children[parentID] = append(s.children[parentID], id)
	}
}

func (s *System) Parse(raw string) (Instance, error) {
	var any struct {
		Metadata `json:"metadata"`
	}

	if err := json.Unmarshal([]byte(raw), &any); err != nil {
		return nil, err
	}

	typeID := any.Metadata.Type.String()

	if typ, ok := s.types[typeID]; !ok {
		return nil, fmt.Errorf("unknown type: %s", typeID)
	} else {
		val := reflect.New(typ).Interface()

		if err := json.Unmarshal([]byte(raw), val); err != nil {
			return nil, err
		}

		inst, ok := val.(Instance)

		if !ok {
			return nil, fmt.Errorf("type was not an instance: %T", val)
		}

		return inst, nil
	}
}

func (s *System) InheritsFrom(id ID, super ID) bool {
	superString := super.String()

	for typeID := id.String(); typeID != ""; typeID = s.parent[typeID] {
		if typeID == superString {
			return true
		}
	}

	return false
}

func (s *System) IsA(inst Instance, id ID) bool {
	firstTypeID := inst.Type().String()

	if firstTypeID == "" {
		instType := reflect.TypeOf(inst)
		if instType.Kind() == reflect.Ptr {
			instType = instType.Elem()
		}
		firstTypeID = s.typeIDs[instType]
	}

	return s.InheritsFrom(ID(firstTypeID), id)
}

var system = NewSystem()

func RegisterType(instance interface{}, id string, parentID string) {
	system.RegisterType(instance, id, parentID)
}

func Parse(raw string) (Instance, error) {
	return system.Parse(raw)
}

func InheritsFrom(id ID, super ID) bool {
	return system.InheritsFrom(id, super)
}

func IsA(inst Instance, id ID) bool {
	return system.IsA(inst, id)
}
