package storetest

import (
	"github.com/uswitch/ontology/pkg/store"
)


func ThingWithType(thingID string, typeID string, properties store.Properties) *store.Thing {
	thing := &store.Thing{
		Metadata: store.Metadata{
			ID:   store.ID(thingID),
			Type: store.ID(typeID),
		},
	}

	if properties != nil {
		thing.Properties = properties
	}

	return thing
}
func Entity(ID string) *store.Thing   { return ThingWithType(ID, "/entity", nil) }
func EntityWithType(ID, typ string) *store.Thing   { return ThingWithType(ID, typ, nil) }
func Relation(ID string) *store.Thing { return ThingWithType(ID, "/relation", nil) }
func Type(ID, parent string, spec map[string]interface{}) *store.Type {
	props := store.Properties{}

	if parent != "" {
		props["parent"] = parent
	}
	props["spec"] = spec

	return (*store.Type)(ThingWithType(ID, "/type", props))
}
func RelationBetween(ID, a, b string) *store.Thing {
	return RelationBetweenWithType(ID, "/relation", a, b)
}
func RelationBetweenWithType(ID, typ, a, b string) *store.Thing {
	return ThingWithType(
		ID, typ,
		store.Properties{
			"a": a,
			"b": b,
		},
	)
}
