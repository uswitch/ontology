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

func funcCall(name string, args []Statement) string {
	argStrings := make([]string, len(args))
	for idx, arg := range args {
		argStrings[idx] = arg.String()
	}

	return fmt.Sprintf("%s(%s)", name, strings.Join(argStrings, ", "))
}

func Keyword(raw interface{}) Statement {
	var part string
	switch s := raw.(type) {
	case string:
		part = s
	case []byte:
		part = string(s)
	case fmt.Stringer:
		part = s.String()
	default:
		panic(fmt.Sprintf("keyword unknown type: %T!", raw))
	}

	return Statement{
		parts: []string{part},
	}
}

func String(raw interface{}) Statement {
	var part string
	switch s := raw.(type) {
	case string:
		part = s
	case []byte:
		part = string(s)
	case fmt.Stringer:
		part = s.String()
	default:
		panic(fmt.Sprintf("string unknown type: %T!", raw))
	}

	return Statement{
		parts: []string{fmt.Sprintf("'%s'", part)},
	}
}

func Int(n int) Statement {
	return Statement{
		parts: []string{fmt.Sprintf("%d", n)},
	}
}

var G = Keyword("g")

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
func InE(ss ...Statement) Statement {
	return Statement{
		parts: []string{funcCall("inE", ss)},
	}
}
func (s Statement) InE(label Statement) Statement {
	return Statement{
		parts: append(s.parts, fmt.Sprintf("inE(%s)", label.String())),
	}
}
func OutE(ss ...Statement) Statement {
	return Statement{
		parts: []string{funcCall("outE", ss)},
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

func Within(ss ...Statement) Statement {
	return Statement{
		parts: []string{funcCall("within", ss)},
	}
}

func Without(ss ...Statement) Statement {
	return Statement{
		parts: []string{funcCall("without", ss)},
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

func (s Statement) E(ss ...Statement) Statement {
	return Statement{
		parts: append(s.parts, funcCall("E", ss)),
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

func (s Statement) OutE(ss ...Statement) Statement {
	return Statement{
		parts: append(s.parts, funcCall("outE", ss)),
	}
}

func (s Statement) InV() Statement {
	return Statement{
		parts: append(s.parts, "inV()"),
	}
}

func (s Statement) Label() Statement {
	return Statement{
		parts: append(s.parts, "label()"),
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

func (s Statement) Order() Statement {
	return Statement{
		parts: append(s.parts, "order()"),
	}
}

func (s Statement) By(k, v string) Statement {
	return Statement{
		parts: append(s.parts, fmt.Sprintf("by(%s, %s)", k, v)),
	}
}
func (s Statement) Range(k, v uint) Statement {
	return Statement{
		parts: append(s.parts, fmt.Sprintf("range(%d, %d)", k, v)),
	}
}
func (s Statement) Has(ss ...Statement) Statement {
	return Statement{
		parts: append(s.parts, funcCall("has", ss)),
	}
}
func (s Statement) Property(ss ...Statement) Statement {
	return Statement{
		parts: append(s.parts, funcCall("property", ss)),
	}
}
func (s Statement) Values(ss ...Statement) Statement {
	return Statement{
		parts: append(s.parts, funcCall("values", ss)),
	}
}

func (s Statement) AddV(ss ...Statement) Statement {
	return Statement{
		parts: append(s.parts, funcCall("addV", ss)),
	}
}

func (s Statement) AddE(ss ...Statement) Statement {
	return Statement{
		parts: append(s.parts, funcCall("addE", ss)),
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

func (s Statement) HasLabel(ss ...Statement) Statement {
	return Statement{
		parts: append(s.parts, funcCall("hasLabel", ss)),
	}
}

func (s Statement) HasNot(ss ...Statement) Statement {
	return Statement{
		parts: append(s.parts, funcCall("hasNot", ss)),
	}
}

func (s Statement) From(ss ...Statement) Statement {
	return Statement{
		parts: append(s.parts, funcCall("from", ss)),
	}
}

func (s Statement) Times(ss ...Statement) Statement {
	return Statement{
		parts: append(s.parts, funcCall("times", ss)),
	}
}

func (s Statement) Dedup(ss ...Statement) Statement {
	return Statement{
		parts: append(s.parts, funcCall("dedup", ss)),
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

func (s Statement) ToList(ss ...Statement) Statement {
	return Statement{
		parts: append(s.parts, funcCall("toList", ss)),
	}
}

func (s Statement) Coalesce(ss ...Statement) Statement {
	return Statement{
		parts: append(s.parts, funcCall("coalesce", ss)),
	}
}

func (s Statement) Barrier(ss ...Statement) Statement {
	return Statement{
		parts: append(s.parts, funcCall("barrier", ss)),
	}
}

func (s Statement) TryNext(ss ...Statement) Statement {
	return Statement{
		parts: append(s.parts, funcCall("tryNext", ss)),
	}
}

func (s Statement) Inject(ss ...Statement) Statement {
	return Statement{
		parts: append(s.parts, funcCall("inject", ss)),
	}
}

func (s Statement) Fold(ss ...Statement) Statement {
	return Statement{
		parts: append(s.parts, funcCall("fold", ss)),
	}
}

func (s Statement) Map(ss ...Statement) Statement {
	return Statement{
		parts: append(s.parts, funcCall("map", ss)),
	}
}

func (s Statement) Where(ss ...Statement) Statement {
	return Statement{
		parts: append(s.parts, funcCall("where", ss)),
	}
}

func HasLabel(ss ...Statement) Statement {
	return Statement{
		parts: []string{funcCall("hasLabel", ss)},
	}
}
func Has(ss ...Statement) Statement {
	return Statement{
		parts: []string{funcCall("has", ss)},
	}
}

func E(ss ...Statement) Statement {
	return Statement{
		parts: []string{funcCall("E", ss)},
	}
}

func Identity(ss ...Statement) Statement {
	return Statement{
		parts: []string{funcCall("identity", ss)},
	}
}

func Range(ss ...Statement) Statement {
	return Statement{
		parts: []string{funcCall("range", ss)},
	}
}

func Out(ss ...Statement) Statement {
	return Statement{
		parts: []string{funcCall("out", ss)},
	}
}

func In(ss ...Statement) Statement {
	return Statement{
		parts: []string{funcCall("in", ss)},
	}
}

func Unfold(ss ...Statement) Statement {
	return Statement{
		parts: []string{funcCall("unfold", ss)},
	}
}

func AddV(ss ...Statement) Statement {
	return Statement{
		parts: []string{funcCall("addV", ss)},
	}
}

func AddE(ss ...Statement) Statement {
	return Statement{
		parts: []string{funcCall("addE", ss)},
	}
}

func OutV(ss ...Statement) Statement {
	return Statement{
		parts: []string{funcCall("outV", ss)},
	}
}

func V(ss ...Statement) Statement {
	return Statement{
		parts: []string{funcCall("V", ss)},
	}
}
