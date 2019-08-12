package store

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrUnimplemented = errors.New("Unimplemented")
	ErrNotFound      = errors.New("Thing not found")
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

func (t *Thing) Thing() *Thing { return t }
func (t *Thing) String() string {
	return fmt.Sprintf("%v[%v]%v", t.Metadata.ID, t.Metadata.Type, t.Properties)
}

type Entity Thing

func (t *Entity) Thing() *Thing { return (*Thing)(t) }

type Relation Thing

func (t *Relation) Thing() *Thing { return (*Thing)(t) }

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

type ListOptions struct {
	SortOrder
	SortField

	Offset          uint
	NumberOfResults uint
}

type Store interface {
	Add(*Thing) error
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
}
