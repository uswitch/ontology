package gremlin

import (
	"fmt"
	"strings"
)

/*type Statement interface {
	Visit() string
}

type StringType string

func (t StringType) Visit() string {
	// TODO: this needs to be protected from injection!
	return string(t)
}

type FuncType struct {
	Name string
	Args []Statement
}

type VarType struct {
	Name string
}

type AssignmentStatement struct {
}

type CallStatement struct {
	Thing Statement
	Call  Statement
}*/

type Statement struct {
	parts []string
}

func Graph() Statement {
	return Statement{
		parts: []string{"graph.traversal()"},
	}
}

func (s Statement) String() string {
	return strings.Join(s.parts, ".")
}

func (s Statement) V() Statement {
	return Statement{
		parts: append(s.parts, "V()"),
	}
}

func (s Statement) E() Statement {
	return Statement{
		parts: append(s.parts, "E()"),
	}
}

func (s Statement) Drop() Statement {
	return Statement{
		parts: append(s.parts, "drop()"),
	}
}

func (s Statement) Iterate() Statement {
	return Statement{
		parts: append(s.parts, "iterate()"),
	}
}

func (s Statement) Has(k, v string) Statement {
	return Statement{
		parts: append(s.parts, fmt.Sprintf("has('%s', '%s')", k, v)),
	}
}
func (s Statement) Values(k string) Statement {
	return Statement{
		parts: append(s.parts, fmt.Sprintf("values('%s')", k)),
	}
}

func (s Statement) AddV(label string) Statement {
	return Statement{
		parts: append(s.parts, fmt.Sprintf("addV('%s')", label)),
	}
}
