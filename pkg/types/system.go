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
	parents  = map[string]string{}
	children = map[string][]string{}
)

func RegisterType(instance interface{}, id string, parent string) {
	typ := reflect.TypeOf(instance)
	types[id] = typ
	typeIDs[typ] = id

	if parent != "" {
		parents[id] = parent
		if _, ok := children[parent]; !ok {
			children[parent] = []string{}
		}
		children[parent] = append(children[parent], id)
	}
}

type Instance interface {
	ID() ID
	Type() ID
}

type Any struct {
	Metadata   `json:"metadata"`
	Properties struct{} `json:"properties"`
}

func (i *Any) ID() ID {
	return i.Metadata.ID
}

func (i *Any) Type() ID {
	return i.Metadata.Type
}

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
