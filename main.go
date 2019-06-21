package main

import "fmt"

type AttributeMap map[string]interface{}
type Schema string

type TypeReference struct {
	Name    string
	Version string
}

type TypeRegistry map[TypeReference]Type

func (tr TypeRegistry) Register(ts ...Type) {
	for _, t := range ts {
		tr[t.Reference()] = t
	}
}

type Type struct {
	Name    string
	Version string

	Parent   TypeReference
	children []TypeReference
}

func (t Type) Subtype(st Type) Type {
	st.Parent = t.Reference()
	t.children = append(t.children, st.Reference())
	return st
}

func (t Type) Reference() TypeReference {
	return TypeReference{Name: t.Name, Version: t.Version}
}

type Metadata struct {
	Type TypeReference
	Name string
}

type EntityType Type

type Entity struct {
	Metadata
}

type RelationType struct {
	Type
	AClass *EntityType
	BClass *EntityType
}

type Relation struct {
	Metadata
	A *Entity
	B *Entity
}

var Base = Type{Name: "Base", Version: "v1"}

var Asset = Base.Subtype(EntityType{Name: "Asset", Version: "v1"})
var Repository = Asset.Subtype(Type{Name: "Repository", Version: "v1"})
var Service = Asset.Subtype(Type{Name: "Service", Version: "v1"})
var AwsS3Bucket = Asset.Subtype(Type{Name: "AwsS3Bucket", Version: "v1"})
var KubernetesPod = Asset.Subtype(Type{Name: "KubernetesPod", Version: "v1"})

var Role = Base.Subtype(Type{Name: "Role", Version: "v1"})

var Person = Base.Subtype(Type{Name: "Person", Version: "v1"})
var Team = Base.Subtype(Type{Name: "Team", Version: "v1"})

var DataClassification = Base.Subtype(Type{Name: "DataClassification", Version: "v1"})

func main() {
	registry := TypeRegistry{}

	registry.Register(
		Base, Asset, Repository, Service, AwsS3Bucket,
		KubernetesPod, Role, Person, Team, DataClassification,
	)

	fmt.Println("Hello")
}
