package v1

import (
	"github.com/uswitch/ontology/pkg/types"
	"github.com/uswitch/ontology/pkg/types/relation"
)

type IsTheSameAs struct{ relation.Relation }

func init() { types.RegisterType(IsTheSameAs{}, "/relation/v1/is_the_same_as", relation.ID) }

type IsPartOf struct{ relation.Relation }

func init() { types.RegisterType(IsPartOf{}, "/relation/v1/is_part_of", relation.ID) }

type IsClassifiedAs struct{ relation.Relation }

func init() { types.RegisterType(IsClassifiedAs{}, "/relation/v1/is_classified_as", relation.ID) }

type IsRunningOn struct{ relation.Relation }

func init() { types.RegisterType(IsRunningOn{}, "/relation/v1/is_running_on", relation.ID) }

type Supervises struct{ relation.Relation }

func init() { types.RegisterType(Supervises{}, "/relation/v1/supervises", relation.ID) }

type WasBuiltBy struct {
	relation.Relation
	Properties struct {
		relation.Properties
		Ref *string `json:"ref"`
		At  string  `json:"at"`
	} `json:"properties"`
}

func init() { types.RegisterType(WasBuiltBy{}, "/relation/v1/was_built_by", relation.ID) }

type WasDeployedBy struct {
	relation.Relation
	Properties struct {
		relation.Properties
		Ref *string `json:"ref"`
		At  string  `json:"at"`
		URL string  `json:"url"`
	} `json:"properties"`
}

func init() { types.RegisterType(WasDeployedBy{}, "/relation/v1/was_deployed_by", relation.ID) }
