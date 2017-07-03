package serve

import (
	"github.com/gomqtt/packet"
	"math"
	"strings"
)

func matches(matcher string, topic string) bool {

	matchingPatterns := allTopics(topic)
	for _, m := range matchingPatterns {
		if m == matcher {
			return true
		}
	}
	return false
}

/*
 * allTopics returns a list of all possible matches for the given topic,
 * including the possible wildcard matches
 *
 * Compiling a list of possible matches is faster if used several times,
 * for example, against many client authorization rules
 *
 */
func allTopics(topic string) []string {
	topics := strings.Split(topic, "/")
	systemTopic := ""

	// MQTT-4.7.2-1 topics begining with $ should not match on wildcard
	// Remove the first topic if this is the case and then append it to the
	// answers at the end
	if strings.HasPrefix(topics[0], "$") {
		systemTopic = topics[0]
		topics = topics[1 : len(topics)-1]
	}

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

	// MQTT-4.7.2-1 topics begining with $
	if systemTopic != "" {
		for i := 0; i < len(all); i++ {
			all[i] = systemTopic + "/" + all[i]
		}
	}

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

/*
 * re-packages a message with the given QOS and retail flag.
 */
func repackage(msg *packet.Message, qos byte, retain bool) (m *packet.Message) {
	m = &packet.Message{
		Topic:   msg.Topic,
		Payload: msg.Payload,
		QOS:     qos,
		Retain:  retain, // Retain to false for all normal subscriptions (MQTT-3.3.1-9)
	}
	return m
}
