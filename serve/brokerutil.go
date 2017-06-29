package serve

import (
	"math"
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

/*
 * allTopics returns a list of all possible matches for the given topic,
 * including the possible wildcard matches
 */
func allTopics(topic string) []string {
	topics := strings.Split(topic, "/")
	numTopics := len(topics)
	all := make([]string, 0)
	wtop := addIntopicWilds(topics)
	for i := 0; i < len(wtop); i++ {
		all = append(all, strings.Join(wtop[i], "/"))
	}
	for i := 1; i < numTopics; i++ {
		wt := addIntopicWilds(topics[0 : numTopics-i])
		for j := 0; j < len(wt); j++ {
			all = append(all, strings.Join(wt[j], "/")+"/#")
		}
	}
	all = append(all, "#")
	return all
}

/*
 * Topic level wildcard
 *
 * All the combiniations of using "+" by considering "+" as a binary
 * placement wherever 1 is in a sequence
 */
func addIntopicWilds(topics []string) (ret [][]string) {

	numTopics := len(topics)
	numIterations := int(math.Pow(2, float64(numTopics)))

	// Start with a blank canvas
	ret = make([][]string, numIterations)
	for i := 0; i < numIterations; i++ {
		ret[i] = append([]string{}, topics...)
	}
	// Now add "+" in the right places
	for i := 0; i < numIterations; i++ {
		bits := i // Use binary representation of i
		for j := 0; j < numTopics; j++ {
			if bits&1 == 1 {
				ret[i][j] = "+"
			}
			bits = bits >> 1
		}
	}
	return ret
}
