package api

import (
	"context"
	"time"

	"defs.dev/schema/api/core"
)

// ServiceStatus represents the current status of a service entity.
type ServiceStatus struct {
	State     ServiceState   `json:"state"`
	StartedAt *time.Time     `json:"startedAt,omitempty"`
	StoppedAt *time.Time     `json:"stoppedAt,omitempty"`
	LastError *string        `json:"lastError,omitempty"`
	Healthy   bool           `json:"healthy"`
	Message   string         `json:"message,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

// ServiceState represents the lifecycle state of a service.
type ServiceState string

const (
	ServiceStateUnknown  ServiceState = "unknown"
	ServiceStateStopped  ServiceState = "stopped"
	ServiceStateStarting ServiceState = "starting"
	ServiceStateRunning  ServiceState = "running"
	ServiceStateStopping ServiceState = "stopping"
	ServiceStateError    ServiceState = "error"
)

// Service defines the interface for service implementations as executable entities.
// This creates symmetry with the Function interface - services are entities that can be called and managed.
type Service interface {
	// Entity execution - call methods on the service
	CallMethod(ctx context.Context, methodName string, params FunctionData) (FunctionData, error)

	// Entity introspection
	Schema() core.ServiceSchema
	Name() string

	// Entity lifecycle management
	Start(ctx context.Context) error
	Stop(ctx context.Context) error

	// Entity state and health
	Status(ctx context.Context) (ServiceStatus, error)
	IsRunning() bool

	// Method introspection
	HasMethod(methodName string) bool
	MethodNames() []string
}
