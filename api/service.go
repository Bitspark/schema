package api

// Service represents a service with multiple methods.
type Service interface {
	Name() string
	Description() string
	Schema() ServiceSchema
	Methods() []string
	GetMethod(name string) (Function, bool)
}
