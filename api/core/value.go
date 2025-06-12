package core

type Copyable[T any] interface {
	Copy() T
}

type Value[T any] interface {
	Copyable[T]

	Value() T
	String() string

	IsNull() bool
	IsComplex() bool
}

type ComplexValue[T any] interface {
	Value[T]

	IsEmpty() bool
	Size() int
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
