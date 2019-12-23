package types

import (
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	system := NewSystem()

	thing := struct {
		Any
		Properties struct {
			Wibble string `json:"wibble"`
		} `json:"properties"`
	}{}

	system.RegisterType(thing, "/test", "")

	parsedThing, err := system.Parse(`
{
  "metadata": {
    "type": "/test"
  },
  "properties": {
    "wibble": "bibble"
  }
}
`)

	if err != nil {
		t.Fatalf("failed to parsed: %v", err)
	}

	// parse returns a pointer, so we need to reflect the difference in types
	if reflect.TypeOf(parsedThing) != reflect.TypeOf(&thing) {
		t.Fatalf("types didn't match: %T != %T", parsedThing, thing)
	}
}

type Person struct{ Any }
type Computer struct{ Any }
type Laptop struct{ Any }

func TestIsA(t *testing.T) {
	system := NewSystem()

	system.RegisterType(Person{}, "/person", "")
	system.RegisterType(Computer{}, "/computer", "")
	system.RegisterType(Laptop{}, "/laptop", "/computer")

	if !system.IsA(&Computer{Any{Metadata: Metadata{Type: "/computer"}}}, "/computer") {
		t.Errorf("computer should be a type of computer")
	}
	if !system.IsA(&Laptop{Any{Metadata: Metadata{Type: "/laptop"}}}, "/computer") {
		t.Errorf("laptop should be a type of computer")
	}

	if !system.IsA(&Computer{}, "/computer") {
		t.Errorf("computer should be a type of computer")
	}
	if !system.IsA(&Laptop{}, "/computer") {
		t.Errorf("laptop should be a type of computer")
	}
}

func TestInheritsFrom(t *testing.T) {
	system := NewSystem()

	system.RegisterType(Person{}, "/person", "")
	system.RegisterType(Computer{}, "/computer", "")
	system.RegisterType(Laptop{}, "/laptop", "/computer")

	if !system.InheritsFrom("/computer", "/computer") {
		t.Errorf("computer should be a type of computer")
	}
	if system.InheritsFrom("/person", "/computer") {
		t.Errorf("person should not be a type of computer")
	}
	if !system.InheritsFrom("/laptop", "/computer") {
		t.Errorf("laptop should be a type of computer")
	}
}
