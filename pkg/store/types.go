package store

import (
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
func (t1 *Thing) Equal(t2 *Thing) bool {
	return reflect.DeepEqual(t1, t2)
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

type Store interface {
	Add(...*Thing) error
	AddAll([]*Thing) error

	Len() (int, error)

	IsA(*Thing, *Type) (bool, error)

	GetByID(ID) (*Thing, error)
	GetEntityByID(ID) (*Entity, error)
	GetRelationByID(ID) (*Relation, error)
	GetTypeByID(ID) (*Type, error)

	List(ListOptions) ([]*Thing, error)
	ListByType(*Type, ListOptions) ([]*Thing, error)

	ListEntities(ListOptions) ([]*Entity, error)
	ListRelations(ListOptions) ([]*Relation, error)
	ListTypes(ListOptions) ([]*Type, error)

	ListRelationsForEntity(*Entity, ListOptions) ([]*Relation, error)
}
