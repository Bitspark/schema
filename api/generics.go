package api

import "defs.dev/schema/api/core"

// Generic schema interfaces for type-safe schema construction

// SchemaBuilder defines the interface for type-safe schema building.
type SchemaBuilder[T any] interface {
	core.Builder[core.Schema]
}

// ListBuilder defines the interface for building List[T] schemas with type safety.
type ListBuilder[T any] interface {
	core.Builder[core.ArraySchema]
	core.MetadataBuilder[ListBuilder[T]]

	MinItems(min int) ListBuilder[T]
	MaxItems(max int) ListBuilder[T]
	UniqueItems() ListBuilder[T]
	Example(example []T) ListBuilder[T]
}

// ListSchema defines the interface for list schemas with type constraints.
type ListSchema[T any] interface {
	core.ArraySchema
}

// OptionalBuilder defines the interface for building Optional[T] schemas.
type OptionalBuilder[T any] interface {
	core.Builder[core.Schema]
	core.MetadataBuilder[OptionalBuilder[T]]

	Example(example *T) OptionalBuilder[T]
}

// OptionalSchema defines the interface for optional/nullable schemas.
type OptionalSchema[T any] interface {
	core.Schema
	core.Accepter

	// Introspection methods
	ItemSchema() core.Schema
}

// ResultBuilder defines the interface for building Result[T, E] schemas.
type ResultBuilder[T, E any] interface {
	core.Builder[core.Schema]
	core.MetadataBuilder[ResultBuilder[T, E]]
}

// ResultSchema defines the interface for result schemas (success/failure patterns).
type ResultSchema[T, E any] interface {
	core.Schema
	core.Accepter

	// Introspection methods
	SuccessSchema() core.Schema
	ErrorSchema() core.Schema
}

// MapBuilder defines the interface for building Map[K, V] schemas.
type MapBuilder[K comparable, V any] interface {
	core.Builder[core.Schema]
	core.MetadataBuilder[MapBuilder[K, V]]

	MinItems(min int) MapBuilder[K, V]
	MaxItems(max int) MapBuilder[K, V]
}

// MapSchema defines the interface for map schemas with key-value constraints.
type MapSchema[K comparable, V any] interface {
	core.Schema
	core.Accepter

	// Introspection methods
	KeySchema() core.Schema
	ValueSchema() core.Schema
	MinItems() *int
	MaxItems() *int
}

// UnionBuilder2 defines the interface for building Union[T1, T2] schemas.
type UnionBuilder2[T1, T2 any] interface {
	core.Builder[core.UnionSchema]
	core.MetadataBuilder[UnionBuilder2[T1, T2]]
}
