package broker

import (
	"testing"
)

// Calling the same underlying code as TestValidate
func TestMatcher(t *testing.T) {
	// matcher, topic

	if !matches("", "") {
		t.Error("Expected matching topic and rights to be valid")
	}
	if !matches("test", "test") {
		t.Error("Expected matching topic and rights to validate")
	}
	if !matches("one/two/three", "one/two/three") {
		t.Error("Expected matching topic and rights to validate")
	}
	if !matches("one/+/three", "one/two/three") {
		t.Error("Expected matching topic and rights with level wildcard to validate")
	}
	if !matches("one/#", "one/two/three") {
		t.Error("Expected matching topic and rights with partial wildcard to validate")
	}
	if !matches("#", "one/two/three") {
		t.Error("Expected matching topic and rights with wildcard to validate")
	}

	// Invalid matches
	if matches("Test", "Test/Test") {
		t.Error("Should not match")
	}
	if matches("one", "two") {
		t.Error("Expected not matching topic and rights to be invalid")
	}
	if matches("one/bad/three", "one/two/three") {
		t.Error("Expected not matching topic and rights to be invalid")
	}
	if matches("one/+/three", "bad/two/three") {
		t.Error("Expected not matching topic and rights to be invalid")
	}
	if matches("one/#", "bad/two/three") {
		t.Error("Expected not matching topic and rights to be invalid")
	}
	if matches("", "one/two/three") {
		t.Error("Expected not matching topic and rights to be invalid")
	}
}
