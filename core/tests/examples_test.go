package tests

import (
	"testing"

	"defs.dev/schema/core/examples"
)

func TestExamples(t *testing.T) {
	// This test just runs the examples to make sure they don't panic
	// and work as expected
	examples.RunAllExamples()
}
