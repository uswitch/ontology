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

type Statements []Statement

func (ss Statements) String() string {
	lines := make([]string, len(ss))

	for idx, s := range ss {
		lines[idx] = s.String()
	}

	return strings.Join(lines, "\n")
}

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

func (s Statement) Count() Statement {
	return Statement{
		parts: append(s.parts, "count()"),
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

func (s Statement) AddE(label string) Statement {
	return Statement{
		parts: append(s.parts, fmt.Sprintf("addE('%s')", label)),
	}
}

func (s Statement) From(label string) Statement {
	return Statement{
		parts: append(s.parts, fmt.Sprintf("from('%s')", label)),
	}
}

func (s Statement) To(label string) Statement {
	return Statement{
		parts: append(s.parts, fmt.Sprintf("to('%s')", label)),
	}
}

func (s Statement) Union(other Statement) Statement {
	return Statement{
		parts: append(s.parts, fmt.Sprintf("union(%s)", other.String())),
	}
}

func Assign(k string, s Statement) Statement {
	return Statement{
		parts: []string{fmt.Sprintf("%s = %s", k, s.String())},
	}
}

func Var(k string) Statement {
	return Statement{
		parts: []string{k},
	}
}

func Add(a Statement, b Statement) Statement {
	return Statement{
		parts: []string{fmt.Sprintf("%s + %s", a.String(), b.String())},
	}
}
