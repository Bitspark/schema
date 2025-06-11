package tests

import (
	"defs.dev/schema/examples"
	"testing"
)

func TestExamples(t *testing.T) {
	// This test just runs the examples to make sure they don't panic
	// and work as expected
	examples.RunAllExamples()
}
