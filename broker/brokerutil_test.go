package broker

import (
	"testing"
)

// Calling the same underlying code as TestValidate
func TestMatcher(t *testing.T) {
	// matcher, topic
	if matches("Test", "Test/Test") {
		t.Error("Should not match")
	}
}
