package types

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"
)

// type basics

type ID string

type IDable interface {
	ID() ID
}

func (id ID) String() string { return string(id) }
func (id ID) ID() ID         { return id }

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

func Parse(raw string) (interface{}, error) {
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

		return val, nil
	}
}

type Any struct {
	Metadata   `json:"metadata"`
	Properties struct{} `json:"properties"`
}

func init() { RegisterType(Any{}, "/any", "") }
