package mqtt

import (
	"testing"
)

// TestValidate tests that the validate function correctly identiys which
// topic and access pairs are valid
func TestValidate(t *testing.T) {

	// Valid matches

	if !validate("", "") {
		t.Error("Expected matching topic and rights to be valid")
	}
	if !validate("test", "test") {
		t.Error("Expected matching topic and rights to validate")
	}
	if !validate("one/two/three", "one/two/three") {
		t.Error("Expected matching topic and rights to validate")
	}
	if !validate("one/two/three", "one/+/three") {
		t.Error("Expected matching topic and rights with level wildcard to validate")
	}
	if !validate("one/two/three", "one/#") {
		t.Error("Expected matching topic and rights with partial wildcard to validate")
	}
	if !validate("one/two/three", "#") {
		t.Error("Expected matching topic and rights with wildcard to validate")
	}

	// Invalid matches

	if validate("one", "two") {
		t.Error("Expected not matching topic and rights to be invalid")
	}
	if validate("one/bad/three", "one/two/three") {
		t.Error("Expected not matching topic and rights to be invalid")
	}
	if validate("bad/two/three", "one/+/three") {
		t.Error("Expected not matching topic and rights to be invalid")
	}
	if validate("bad/two/three", "one/#") {
		t.Error("Expected not matching topic and rights to be invalid")
	}
	if validate("one/two/three", "") {
		t.Error("Expected not matching topic and rights to be invalid")
	}

}
