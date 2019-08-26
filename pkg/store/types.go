package store

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"
)

var (
	ErrUnimplemented     = errors.New("Unimplemented")
	ErrNotFound          = errors.New("Thing not found")
	ErrEntityNotInvolved = errors.New("Entity not involved in relationship")
)

type ID string

func (id ID) String() string { return string(id) }

type Metadata struct {
	ID        ID
	Type      ID
	Name      string
	UpdatedAt time.Time
}

type Properties map[string]interface{}

type Thing struct {
	Metadata
	Properties
}

type Thingable interface {
	Thing() *Thing
}

func (t *Thing) Thing() *Thing { return t }
func (t *Thing) String() string {
	return fmt.Sprintf("%v[%v]%v", t.Metadata.ID, t.Metadata.Type, t.Properties)
}
func (t1 *Thing) Equal(ts ...Thingable) bool {
	for _, t := range ts {
		if ! reflect.DeepEqual(t1, t.Thing()) {
			return false
		}
	}
	return true
}

type Entity Thing

func (t *Entity) Thing() *Thing { return (*Thing)(t) }

type Relation Thing

func (t *Relation) Thing() *Thing { return (*Thing)(t) }
func (r *Relation) Involves(entity *Entity) bool {
	a, aOk := r.Properties["a"].(string)
	b, bOk := r.Properties["b"].(string)

	return (aOk && ID(a) == entity.Metadata.ID) || (bOk && ID(b) == entity.Metadata.ID)
}
func (r *Relation) OtherID(entity *Entity) (ID, error) {
	a := ID(r.Properties["a"].(string))
	b := ID(r.Properties["b"].(string))

	if a == entity.Metadata.ID {
		return b, nil
	} else if b == entity.Metadata.ID {
		return a, nil
	} else {
		return ID(""), ErrEntityNotInvolved
	}
}

type Type Thing

func (t *Type) Thing() *Thing { return (*Thing)(t) }

var (
	TypeType = &Type{
		Metadata: Metadata{
			ID:   ID("/type"),
			Type: ID("/type"),
		},
		Properties: Properties{
			"spec": map[string]interface{}{
				"parent": map[string]string{
					"type": "string",
					"pointer_to": "/type",
				},
				"template": map[string]string{
					"type": "string",
				},
				"spec": map[string]string{
					"type": "object",
				},
				"required": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"type": "string",
					},
				},
			},
		},
	}

	EntityType = &Type{
		Metadata: Metadata{
			ID:   ID("/entity"),
			Type: ID("/type"),
		},
	}

	RelationType = &Type{
		Metadata: Metadata{
			ID:   ID("/relation"),
			Type: ID("/type"),
		},
		Properties: Properties{
			"spec": map[string]interface{}{
				"a": map[string]string{
					"type": "string",
					"pointer_to": "/entity",
				},
				"b": map[string]string{
					"type": "string",
					"pointer_to": "/entity",
				},
			},
			"required": []string{"a", "b"},
		},
	}
)

type SortOrder uint

const (
	SortAscending = SortOrder(iota)
	SortDescending
)

type SortField uint

const (
	SortByID = SortField(iota)
)

const DefaultNumberOfResults = uint(10)

type ListOptions struct {
	SortOrder
	SortField

	Offset          uint
	NumberOfResults uint
}


type PointerOptions int
const (
	ResolvePointers = PointerOptions(iota)
	IgnoreMissingPointers
	IgnoreAllPointers
)

type ValidateOptions struct {
	Pointers PointerOptions
}

type ValidationError string

func (e ValidationError) Error() string {
	return string(e)
}

type Store interface {
	Add(context.Context, ...Thingable) error
	AddAll(context.Context, []Thingable) error

	Len(context.Context) (int, error)

	Types(context.Context, Thingable) ([]*Type, error)
	IsA(context.Context, Thingable, *Type) (bool, error)
	Validate(context.Context, Thingable, ValidateOptions) ([]ValidationError, error)

	GetByID(context.Context, ID) (*Thing, error)
	GetEntityByID(context.Context, ID) (*Entity, error)
	GetRelationByID(context.Context, ID) (*Relation, error)
	GetTypeByID(context.Context, ID) (*Type, error)

	List(context.Context, ListOptions) ([]*Thing, error)
	ListByType(context.Context, *Type, ListOptions) ([]*Thing, error)

	ListEntities(context.Context, ListOptions) ([]*Entity, error)
	ListRelations(context.Context, ListOptions) ([]*Relation, error)
	ListTypes(context.Context, ListOptions) ([]*Type, error)

	ListRelationsForEntity(context.Context, *Entity, ListOptions) ([]*Relation, error)

	WatchByType(context.Context, *Type) (chan *Thing, error)
}
