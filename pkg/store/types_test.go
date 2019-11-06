package store

import (
	"testing"
)

func TestThingEqual(t *testing.T) {
	cases := []struct {
		T1            Thing
		T2            Thing
		ShouldBeEqual bool
	}{
		{
			T1:            Thing{Metadata: Metadata{ID: ID("/wibble"), Type: ID("/bibble")}},
			T2:            Thing{Metadata: Metadata{ID: ID("/wibble"), Type: ID("/bibble")}},
			ShouldBeEqual: true,
		},
		{
			T1: Thing{
				Metadata:   Metadata{ID: ID("/wibble"), Type: ID("/bibble")},
				Properties: Properties{"thing": false},
			},
			T2: Thing{
				Metadata:   Metadata{ID: ID("/wibble"), Type: ID("/bibble")},
				Properties: Properties{"thing": false},
			},
			ShouldBeEqual: true,
		},
		{
			T1: Thing{
				Metadata:   Metadata{ID: ID("/wibble"), Type: ID("/bibble")},
				Properties: Properties{"thing": false},
			},
			T2: Thing{
				Metadata:   Metadata{ID: ID("/wibble"), Type: ID("/bibble")},
				Properties: Properties{"thing": true},
			},
			ShouldBeEqual: false,
		},
	}

	for _, c := range cases {
		if c.T1.Equal(&c.T2) != c.ShouldBeEqual {
			t.Errorf("expected %v equal to %v to be %v, it wasn't", c.T1, c.T2, c.ShouldBeEqual)
		}
	}
}

func TestRelationInvolves(t *testing.T) {
	cases := []struct {
		Rel Relation
		Ent Entity

		ShouldBeInvolved bool
	}{
		{
			Rel: Relation{Properties: Properties{"a": "/ent/1", "b": "/ent/2"}},
			Ent: Entity{Metadata: Metadata{ID: ID("/ent/1")}},

			ShouldBeInvolved: true,
		},
	}

	for _, c := range cases {
		if c.Rel.Involves(&c.Ent) != c.ShouldBeInvolved {
			t.Errorf("expected %v to be involved with %v to be %v, it wasn't", c.Ent, c.Rel, c.ShouldBeInvolved)
		}
	}
}

func TestRelationOtherID(t *testing.T) {
	cases := []struct {
		Rel     Relation
		Ent     Entity
		OtherID ID
		Err     error
	}{
		{
			Rel:     Relation{Properties: Properties{"a": "/ent/1", "b": "/ent/2"}},
			Ent:     Entity{Metadata: Metadata{ID: ID("/ent/1")}},
			OtherID: ID("/ent/2"),
			Err:     nil,
		},
		{
			Rel:     Relation{Properties: Properties{"a": "/ent/1", "b": "/ent/2"}},
			Ent:     Entity{Metadata: Metadata{ID: ID("/ent/3")}},
			OtherID: ID(""),
			Err:     ErrEntityNotInvolved,
		},
	}

	for _, c := range cases {
		if otherID, err := c.Rel.OtherID(&c.Ent); err != c.Err {
			t.Errorf("expected err %v, but it was %v", c.Err, err)
		} else if otherID != c.OtherID {
			t.Errorf("expected %v to be the other id of %v + %v, but it was %v", c.OtherID, c.Rel, c.Ent, otherID)
		}
	}
}
