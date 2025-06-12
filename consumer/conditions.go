package consumer

import (
	"regexp"
	"strings"

	"defs.dev/schema/api/core"
)

// ----------------------------------------------------------------------------
//  SchemaCondition Interface
// ----------------------------------------------------------------------------

type SchemaCondition interface {
	Matches(schema core.Schema) bool
	String() string
}

// ----------------------------------------------------------------------------
//  Logical Conditions
// ----------------------------------------------------------------------------

type AndCondition struct {
	Conditions []SchemaCondition
}

func (c AndCondition) Matches(schema core.Schema) bool {
	for _, cond := range c.Conditions {
		if !cond.Matches(schema) {
			return false
		}
	}
	return true
}

func (c AndCondition) String() string {
	parts := make([]string, len(c.Conditions))
	for i, cond := range c.Conditions {
		parts[i] = cond.String()
	}
	return "AND(" + strings.Join(parts, ",") + ")"
}

type OrCondition struct {
	Conditions []SchemaCondition
}

func (c OrCondition) Matches(schema core.Schema) bool {
	if len(c.Conditions) == 0 {
		return false
	}
	for _, cond := range c.Conditions {
		if cond.Matches(schema) {
			return true
		}
	}
	return false
}

func (c OrCondition) String() string {
	parts := make([]string, len(c.Conditions))
	for i, cond := range c.Conditions {
		parts[i] = cond.String()
	}
	return "OR(" + strings.Join(parts, ",") + ")"
}

type NotCondition struct {
	Condition SchemaCondition
}

func (c NotCondition) Matches(schema core.Schema) bool {
	return !c.Condition.Matches(schema)
}

func (c NotCondition) String() string { return "NOT(" + c.Condition.String() + ")" }

// ----------------------------------------------------------------------------
//  Primitive Conditions
// ----------------------------------------------------------------------------

type TypeCondition struct {
	Type core.SchemaType
}

func (c TypeCondition) Matches(schema core.Schema) bool {
	return schema != nil && schema.Type() == c.Type
}

func (c TypeCondition) String() string { return "Type(" + string(c.Type) + ")" }

type AnyTypeCondition struct {
	Types []core.SchemaType
}

func (c AnyTypeCondition) Matches(schema core.Schema) bool {
	if schema == nil {
		return false
	}
	st := schema.Type()
	for _, t := range c.Types {
		if st == t {
			return true
		}
	}
	return false
}

func (c AnyTypeCondition) String() string { return "AnyType" }

type HasAnnotationCondition struct {
	AnnotationName string
}

func (c HasAnnotationCondition) Matches(schema core.Schema) bool {
	if schema == nil {
		return false
	}
	for _, ann := range schema.Annotations() {
		if ann.Name() == c.AnnotationName {
			return true
		}
	}
	return false
}

func (c HasAnnotationCondition) String() string { return "HasAnn(" + c.AnnotationName + ")" }

type AnnotationCondition struct {
	AnnotationName string
	Value          any
	Operator       string // "equals" (default), "contains", "matches"
}

func (c AnnotationCondition) Matches(schema core.Schema) bool {
	if schema == nil {
		return false
	}
	for _, ann := range schema.Annotations() {
		if ann.Name() != c.AnnotationName {
			continue
		}
		if c.Value == nil {
			return true
		}
		switch c.Operator {
		case "", "equals":
			return ann.Value() == c.Value
		case "contains":
			annStr, ok1 := ann.Value().(string)
			valStr, ok2 := c.Value.(string)
			return ok1 && ok2 && strings.Contains(annStr, valStr)
		case "matches":
			pattern, ok := c.Value.(string)
			if !ok {
				return false
			}
			annStr, ok := ann.Value().(string)
			if !ok {
				return false
			}
			matched, _ := regexp.MatchString(pattern, annStr)
			return matched
		default:
			return false
		}
	}
	return false
}

func (c AnnotationCondition) String() string { return "Ann(" + c.AnnotationName + ")" }

// ----------------------------------------------------------------------------
//  DSL Helper Functions (in same package)
// ----------------------------------------------------------------------------

func And(conds ...SchemaCondition) SchemaCondition { return AndCondition{Conditions: conds} }
func Or(conds ...SchemaCondition) SchemaCondition  { return OrCondition{Conditions: conds} }
func Not(cond SchemaCondition) SchemaCondition     { return NotCondition{Condition: cond} }

func Type(t core.SchemaType) SchemaCondition { return TypeCondition{Type: t} }

func HasAnnotation(name string, value ...any) SchemaCondition {
	if len(value) == 0 {
		return HasAnnotationCondition{AnnotationName: name}
	}
	return AnnotationCondition{AnnotationName: name, Value: value[0]}
}
