package broker

import (
	"strings"
)

func matches(matcher string, topic string) bool {
	matcherlevels := strings.Split(matcher, "/")
	topiclevels := strings.Split(topic, "/")
	if len(matcherlevels) > len(topiclevels) {
		// Case of matcher has more levels than topic
		return false
	}
	for i := 0; i < len(topiclevels); i++ {
		// The Server MUST NOT match Topic Filters starting with a wildcard
		// character (# or +) with Topic Names beginning with a $ character
		// [MQTT-4.7.2-1].
		if i == 0 && strings.HasPrefix(topiclevels[0], "$") {
			if matcherlevels[0] == "#" || matcherlevels[0] == "+" {
				return false
			}
		}
		// Rights levels are not deep enough
		if len(matcherlevels) <= i {
			return false
		}
		// Wildcard here on in, so match everything
		if matcherlevels[i] == "#" {
			return true
		}
		// Topics do not match, and not a topic level wildcard
		if matcherlevels[i] != "+" && matcherlevels[i] != topiclevels[i] {
			return false
		}
	}
	return true
}
