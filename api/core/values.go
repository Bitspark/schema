package core

import (
	"fmt"
	"reflect"
)

type Copyable[T any] interface {
	Copy() T
}

type Value[T any] interface {
	Copyable[T]

	Value() T
	String() string

	IsNull() bool
	IsComposite() bool
}

type CompositeEntry[K any, V any] struct {
	// PathFragment is the path fragment of the key.
	//
	// Example: "0", "1" (for arrays), "Field1" (for structures), "key0" (for maps)
	PathFragment string

	// Key is the key of the entry.
	// For arrays, integers. For structures, strings. For maps, usually strings.
	Key K

	// Value is the value of the entry.
	Value Value[V]
}

type CompositeValue[K any, T any] interface {
	Value[T]

	IsEmpty() bool
	Size() int
	Values() []CompositeEntry[K, any]
}

type ArrayValue[T any] interface {
	CompositeValue[int, []T]

	Value() []T
	Length() int
	Get(index int) (T, error)
}

type MapValue[K comparable, V any] interface {
	CompositeValue[K, map[K]V]

	Value() map[K]V
	Values() []CompositeEntry[K, any]
	Get(key K) (V, error)
	Has(key K) bool
}

type StructureValue[S any] interface {
	CompositeValue[string, S]

	Get(key string) (Value[any], error)
	Has(key string) bool
	Keys() []string
	Values() []CompositeEntry[string, any]
	Entries() []map[string]Value[any]
}

type StringValue interface {
	Value[string]
}

type NumberValue interface {
	Value[float64]
}

type IntegerValue interface {
	Value[int]
}

type BooleanValue interface {
	Value[bool]
}

type UnsignedIntegerLike interface {
	uint | uint8 | uint16 | uint32 | uint64
}

type IntegerLike interface {
}

type NumberLike interface {
	int | float64
}

var _ Value[any] = &baseValue[any]{}

type baseValue[T any] struct {
	value     T
	composite bool
}

// Copy implements Value.
func (v *baseValue[T]) Copy() T {
	return v.value
}

// IsComposite implements Value.
func (v *baseValue[T]) IsComposite() bool {
	return v.composite
}

// IsNull implements Value.
func (v *baseValue[T]) IsNull() bool {
	val := reflect.ValueOf(v.value)
	return val.IsNil()
}

// String implements Value.
func (v *baseValue[T]) String() string {
	return fmt.Sprintf("%v", v.value)
}

// Value implements Value.
func (v *baseValue[T]) Value() T {
	return v.value
}
