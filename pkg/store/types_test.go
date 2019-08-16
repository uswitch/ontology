package store

import (
	"testing"
)

func TestThingEqual(t *testing.T) {
	cases := []struct{
		T1 Thing
		T2 Thing
		ShouldBeEqual bool
	}{
		{
			T1: Thing{Metadata:Metadata{ID: ID("/wibble"), Type: ID("/bibble")}},
			T2: Thing{Metadata:Metadata{ID: ID("/wibble"), Type: ID("/bibble")}},
			ShouldBeEqual: true,
		},
		{
			T1: Thing{
				Metadata:Metadata{ID: ID("/wibble"), Type: ID("/bibble")},
				Properties:Properties{"thing": false},
			},
			T2: Thing{
				Metadata:Metadata{ID: ID("/wibble"), Type: ID("/bibble")},
				Properties:Properties{"thing": false},
			},
			ShouldBeEqual: true,
		},
		{
			T1: Thing{
				Metadata:Metadata{ID: ID("/wibble"), Type: ID("/bibble")},
				Properties:Properties{"thing": false},
			},
			T2: Thing{
				Metadata:Metadata{ID: ID("/wibble"), Type: ID("/bibble")},
				Properties:Properties{"thing": true},
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
