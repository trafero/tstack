package broker

import (
	"strings"
)

func matches(matcher string, topic string) bool {
	topiclevels := strings.Split(topic, "/")
	matcherlevels := strings.Split(matcher, "/")

	for i := 0; i < len(topiclevels); i++ {
		// Rights levels are not deep enough
		if len(matcherlevels) < i {
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
	// Matched all of topic
	return true
}
