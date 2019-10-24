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

func Empty() Statement {
	return Statement{
		parts: []string{"__"},
	}
}

func BothE(label string) Statement {
	return Statement{
		parts: []string{fmt.Sprintf("bothE('%s')", label)},
	}
}
func InE(label string) Statement {
	return Statement{
		parts: []string{fmt.Sprintf("inE('%s')", label)},
	}
}
func OutE(label string) Statement {
	return Statement{
		parts: []string{fmt.Sprintf("outE('%s')", label)},
	}
}
func Select(vals ...string) Statement {
	args := make([]string, len(vals))

	for idx, val := range vals {
		args[idx] = fmt.Sprintf("'%s'", val)
	}

	return Statement{
		parts: []string{fmt.Sprintf("select(%s)", strings.Join(args, ", "))},
	}
}
func (s Statement) Select(vals ...string) Statement {
	args := make([]string, len(vals))

	for idx, val := range vals {
		args[idx] = fmt.Sprintf("'%s'", val)
	}

	return Statement{
		parts: append(s.parts, fmt.Sprintf("select(%s)", strings.Join(args, ", "))),
	}
}

func Repeat(other Statement) Statement {
	return Statement{
		parts: []string{fmt.Sprintf("repeat(%s)", other.String())},
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

func (s Statement) V1(inner string) Statement {
	return Statement{
		parts: append(s.parts, fmt.Sprintf("V('%s')", inner)),
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

func (s Statement) Next() Statement {
	return Statement{
		parts: append(s.parts, "next()"),
	}
}

func (s Statement) Count() Statement {
	return Statement{
		parts: append(s.parts, "count()"),
	}
}

func (s Statement) OutE(id string) Statement {
	return Statement{
		parts: append(s.parts, fmt.Sprintf("outE('%s')", id)),
	}
}

func (s Statement) InV() Statement {
	return Statement{
		parts: append(s.parts, "inV()"),
	}
}

func (s Statement) OtherV() Statement {
	return Statement{
		parts: append(s.parts, "otherV()"),
	}
}

func (s Statement) OutV() Statement {
	return Statement{
		parts: append(s.parts, "outV()"),
	}
}

func (s Statement) SimplePath() Statement {
	return Statement{
		parts: append(s.parts, "simplePath()"),
	}
}

func (s Statement) PropertyMap() Statement {
	return Statement{
		parts: append(s.parts, "propertyMap()"),
	}
}

func (s Statement) Emit() Statement {
	return Statement{
		parts: append(s.parts, "emit()"),
	}
}

func (s Statement) Has(k, v string) Statement {
	return Statement{
		parts: append(s.parts, fmt.Sprintf("has('%s', '%s')", k, v)),
	}
}
func (s Statement) Property(k, v string) Statement {
	return Statement{
		parts: append(s.parts, fmt.Sprintf("property('%s', '%s')", k, v)),
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

func (s Statement) BothE(label string) Statement {
	return Statement{
		parts: append(s.parts, fmt.Sprintf("bothE('%s')", label)),
	}
}

func (s Statement) As(label string) Statement {
	return Statement{
		parts: append(s.parts, fmt.Sprintf("as('%s')", label)),
	}
}

func (s Statement) HasLabel(label string) Statement {
	return Statement{
		parts: append(s.parts, fmt.Sprintf("hasLabel('%s')", label)),
	}
}

func (s Statement) From(label string) Statement {
	return Statement{
		parts: append(s.parts, fmt.Sprintf("from('%s')", label)),
	}
}

func (s Statement) Times(num int) Statement {
	return Statement{
		parts: append(s.parts, fmt.Sprintf("times(%d)", num)),
	}
}

func (s Statement) Is(num int) Statement {
	return Statement{
		parts: append(s.parts, fmt.Sprintf("is(%d)", num)),
	}
}

func (s Statement) To(label Statement) Statement {
	return Statement{
		parts: append(s.parts, fmt.Sprintf("to(%s)", label.String())),
	}
}

func (s Statement) Until(label Statement) Statement {
	return Statement{
		parts: append(s.parts, fmt.Sprintf("until(%s)", label.String())),
	}
}

func (s Statement) Repeat(other Statement) Statement {
	return Statement{
		parts: append(s.parts, fmt.Sprintf("repeat(%s)", other.String())),
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

func (s Statement) Union(ss ...Statement) Statement {
	strs := make([]string, len(ss))

	for idx, s := range ss {
		strs[idx] = s.String()
	}

	return Statement{
		parts: append(s.parts, fmt.Sprintf("union(%s)", strings.Join(strs, ", "))),
	}
}
