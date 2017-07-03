package serve

import (
	"strings"
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
	if matches("TopicA/B", "TopicA") {
		t.Error("Expected to match whole topic")
	}

	// Starting with $
	// The Server MUST NOT match Topic Filters starting with a wildcard
	// character (# or +) with Topic Names beginning with a $ character
	// [MQTT-4.7.2-1].
	if matches("#", "$one/two/three") {
		t.Error("Wildcard should not match on $ start")
	}
	if matches("#", "$/two/three") {
		t.Error("Wildcard should not match on $ start")
	}
	if matches("+/two/three", "$one/two/three") {
		t.Error("Wildcard should not match on $ start")
	}

	if !matches("one/#", "one/$two/three") {
		t.Error("Wildcard should match on $ if not at the start")
	}
	if !matches("one/+/three", "one/$two/three") {
		t.Error("Wildcard should match on $ if not at the start")
	}

}

// Calling the same underlying code as TestValidate
func TestAllTopics(t *testing.T) {

	ans := allTopics("this")
	if len(ans) != 3 {
		t.Error("Expected 3 answers, got " + strings.Join(ans, ","))
	}
	ans = allTopics("one/two")
	if len(ans) != 7 {
		t.Error("Expected 7 answers, got " + strings.Join(ans, ","))
	}
}
