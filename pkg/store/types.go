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

type IDable interface {
	ID() ID
}

func (id ID) String() string { return string(id) }
func (id ID) ID() ID         { return id }

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
func (t *Thing) ID() ID        { return t.Metadata.ID }
func (t *Thing) String() string {
	return fmt.Sprintf("%v[%v]%v", t.Metadata.ID, t.Metadata.Type, t.Properties)
}
func (t1 *Thing) Equal(ts ...Thingable) bool {
	for _, t := range ts {
		if !reflect.DeepEqual(t1, t.Thing()) {
			return false
		}
	}
	return true
}

type Entity Thing

func (t *Entity) ID() ID        { return t.Metadata.ID }
func (t *Entity) Thing() *Thing { return (*Thing)(t) }

type Relation Thing

func (t *Relation) ID() ID        { return t.Metadata.ID }
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

func (t *Type) ID() ID        { return t.Metadata.ID }
func (t *Type) Thing() *Thing { return (*Thing)(t) }

var (
	TypeType = &Type{
		Metadata: Metadata{
			ID:   ID("/type"),
			Type: ID("/type"),
			Name: "Type",
		},
		Properties: Properties{
			"spec": map[string]interface{}{
				"parent": map[string]interface{}{
					"type":       "string",
					"pointer_to": "/type",
				},
				"template": map[string]interface{}{
					"type": "string",
				},
				"spec": map[string]interface{}{
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
			Name: "Entity",
		},
	}

	RelationType = &Type{
		Metadata: Metadata{
			ID:   ID("/relation"),
			Type: ID("/type"),
			Name: "Relation",
		},
		Properties: Properties{
			"spec": map[string]interface{}{
				"a": map[string]interface{}{
					"type":       "string",
					"pointer_to": "/entity",
				},
				"b": map[string]interface{}{
					"type":       "string",
					"pointer_to": "/entity",
				},
			},
			"required": []string{"a", "b"},
		},
	}

	TypeOfType = &Type{
		Metadata: Metadata{
			ID:   ID("/relation/type_of"),
			Type: ID("/type"),
			Name: "TypeOF",
		},
		Properties: Properties{
			"parent": "/relation",
		},
	}

	SubtypeOfType = &Type{
		Metadata: Metadata{
			ID:   ID("/relation/subtype_of"),
			Type: ID("/type"),
			Name: "SubtypeOF",
		},
		Properties: Properties{
			"parent": "/relation",
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
	TypeHierarchy(context.Context, *Type) ([]*Type, error)
	Inherits(context.Context, *Type, *Type) (bool, error)
	IsA(context.Context, Thingable, *Type) (bool, error)
	Validate(context.Context, Thingable, ValidateOptions) ([]ValidationError, error)

	GetByID(context.Context, IDable) (*Thing, error)
	GetEntityByID(context.Context, IDable) (*Entity, error)
	GetRelationByID(context.Context, IDable) (*Relation, error)
	GetTypeByID(context.Context, IDable) (*Type, error)

	List(context.Context, ListOptions) ([]*Thing, error)
	ListByType(context.Context, *Type, ListOptions) ([]*Thing, error)

	ListEntities(context.Context, ListOptions) ([]*Entity, error)
	ListRelations(context.Context, ListOptions) ([]*Relation, error)
	ListTypes(context.Context, ListOptions) ([]*Type, error)

	ListRelationsForEntity(context.Context, *Type, *Entity, ListOptions) ([]*Relation, error)

	WatchByID(context.Context, IDable) (chan *Thing, error)
	WatchByType(context.Context, IDable) (chan *Thing, error)
}
