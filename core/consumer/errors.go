package consumer

import (
	"fmt"
	"strings"
)

// ConsumerError wraps errors from consumer processing with context.
type ConsumerError struct {
	Consumer string
	Purpose  ConsumerPurpose
	Path     []string
	Cause    error
}

func (e ConsumerError) Error() string {
	path := "root"
	if len(e.Path) > 0 {
		path = strings.Join(e.Path, ".")
	}
	return fmt.Sprintf("consumer %s (%s) failed at %s: %v",
		e.Consumer, e.Purpose, path, e.Cause)
}

func (e ConsumerError) Unwrap() error {
	return e.Cause
}

// NewConsumerError creates a new ConsumerError.
func NewConsumerError(consumer string, purpose ConsumerPurpose, path []string, cause error) *ConsumerError {
	return &ConsumerError{
		Consumer: consumer,
		Purpose:  purpose,
		Path:     append([]string(nil), path...), // copy slice
		Cause:    cause,
	}
}
