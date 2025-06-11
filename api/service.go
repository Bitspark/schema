package api

import "defs.dev/schema/api/core"

// Service defines the interface for service implementations.
type Service interface {
	Schema() core.ServiceSchema
}
