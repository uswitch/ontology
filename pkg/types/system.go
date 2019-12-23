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

type meta struct {
	Metadata `json:"metadata"`
}

var (
	types    = map[string]reflect.Type{}
	typeIDs  = map[reflect.Type]string{}
	parent   = map[string]string{}
	children = map[string][]string{}
)

func RegisterType(instance interface{}, id string, parentID string) {
	typ := reflect.TypeOf(instance)
	types[id] = typ
	typeIDs[typ] = id

	if parentID != "" {
		parent[id] = parentID
		if _, ok := children[parentID]; !ok {
			children[parentID] = []string{}
		}
		children[parentID] = append(children[parentID], id)
	}
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

func init() { RegisterType(Any{}, "/any", "") }

func Parse(raw string) (Instance, error) {
	var any meta

	if err := json.Unmarshal([]byte(raw), &any); err != nil {
		return nil, err
	}

	typeID := any.Metadata.Type.String()

	if typ, ok := types[typeID]; !ok {
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

func IsA(inst Instance, id ID) bool {
	idString := id.String()

	for typeID := inst.Type().String(); typeID != ""; typeID = parent[typeID] {
		if typeID == idString {
			return true
		}
	}

	return false
}
