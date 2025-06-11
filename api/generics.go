package api

// Generic schema interfaces for type-safe schema construction

// ListBuilder defines the interface for building List[T] schemas with type safety.
type ListBuilder[T any] interface {
	Builder[ArraySchema]
	MetadataBuilder[ListBuilder[T]]

	MinItems(min int) ListBuilder[T]
	MaxItems(max int) ListBuilder[T]
	UniqueItems() ListBuilder[T]
	Example(example []T) ListBuilder[T]
}

// OptionalBuilder defines the interface for building Optional[T] schemas.
type OptionalBuilder[T any] interface {
	Builder[Schema]
	MetadataBuilder[OptionalBuilder[T]]

	Example(example *T) OptionalBuilder[T]
}

// OptionalSchema defines the interface for optional/nullable schemas.
type OptionalSchema[T any] interface {
	Schema
	Accepter

	// Introspection methods
	ItemSchema() Schema
}

// ResultBuilder defines the interface for building Result[T, E] schemas.
type ResultBuilder[T, E any] interface {
	Builder[Schema]
	MetadataBuilder[ResultBuilder[T, E]]
}

// ResultSchema defines the interface for result schemas (success/failure patterns).
type ResultSchema[T, E any] interface {
	Schema
	Accepter

	// Introspection methods
	SuccessSchema() Schema
	ErrorSchema() Schema
}

// MapBuilder defines the interface for building Map[K, V] schemas.
type MapBuilder[K comparable, V any] interface {
	Builder[Schema]
	MetadataBuilder[MapBuilder[K, V]]

	MinItems(min int) MapBuilder[K, V]
	MaxItems(max int) MapBuilder[K, V]
}

// MapSchema defines the interface for map schemas with key-value constraints.
type MapSchema[K comparable, V any] interface {
	Schema
	Accepter

	// Introspection methods
	KeySchema() Schema
	ValueSchema() Schema
	MinItems() *int
	MaxItems() *int
}

// UnionBuilder2 defines the interface for building Union[T1, T2] schemas.
type UnionBuilder2[T1, T2 any] interface {
	Builder[UnionSchema]
	MetadataBuilder[UnionBuilder2[T1, T2]]
}
