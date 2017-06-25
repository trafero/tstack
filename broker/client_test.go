package broker

import (
	"testing"
)

func TestValidate(t *testing.T) {
	// Valid matches
	c := client{rights: ""}
	if !c.authorized("") {
		t.Error("Expected matching topic and rights to be valid")
	}
	c = client{rights: "test"}
	if !c.authorized("test") {
		t.Error("Expected matching topic and rights to validate")
	}
	c = client{rights: "one/two/three"}
	if !c.authorized("one/two/three") {
		t.Error("Expected matching topic and rights to validate")
	}
	c = client{rights: "one/+/three"}
	if !c.authorized("one/two/three") {
		t.Error("Expected matching topic and rights with level wildcard to validate")
	}
	c = client{rights: "one/#"}
	if !c.authorized("one/two/three") {
		t.Error("Expected matching topic and rights with partial wildcard to validate")
	}
	c = client{rights: "#"}
	if !c.authorized("one/two/three") {
		t.Error("Expected matching topic and rights with wildcard to validate")
	}

	// Invalid matches
	c = client{rights: "two"}
	if c.authorized("one") {
		t.Error("Expected not matching topic and rights to be invalid")
	}
	c = client{rights: "one/two/three"}
	if c.authorized("one/bad/three") {
		t.Error("Expected not matching topic and rights to be invalid")
	}
	c = client{rights: "one/+/three"}
	if c.authorized("bad/two/three") {
		t.Error("Expected not matching topic and rights to be invalid")
	}
	c = client{rights: "one/#"}
	if c.authorized("bad/two/three") {
		t.Error("Expected not matching topic and rights to be invalid")
	}
	c = client{rights: ""}
	if c.authorized("one/two/three") {
		t.Error("Expected not matching topic and rights to be invalid")
	}
}
