package serve

import (
	"github.com/gomqtt/packet"
	"sync"
)

type Broker struct {
	sync.RWMutex
	clients     map[string]*client         // Map by clientid
	retained    map[string]*packet.Message // Map by topic
	deliverChan chan *packet.Message       // Place to send message for delierfy
}

func NewBroker() *Broker {
	b := &Broker{
		clients:     make(map[string]*client),
		retained:    make(map[string]*packet.Message),
		deliverChan: make(chan *packet.Message, 10),
	}
	go b.deliveryRound()
	return b
}

func (b *Broker) AddClient(c *client) {

	b.RLock()
	// Clean session: [MQTT-3.1.2-6]
	if existingClient, exists := b.clients[c.clientid]; exists && c.cleanSession == false {
		// clientid already exists
		c.inboundInTransit = existingClient.inboundInTransit
		c.outboundInTransit = existingClient.outboundInTransit
		c.subscriptions = existingClient.subscriptions
	}
	b.RUnlock()
	b.Lock()
	b.clients[c.clientid] = c
	b.Unlock()
}

func (b *Broker) RemoveClient(c *client) {
	b.Lock()
	delete(b.clients, c.clientid)
	b.Unlock()
}

func (b *Broker) deliveryRound() {
	for {
		msg := <-b.deliverChan
		if msg.Retain {
			b.Lock()
			if len(msg.Payload) == 0 {
				// MQTT-3.3.1-10
				delete(b.retained, msg.Topic)
			} else {
				b.retained[msg.Topic] = msg
			}
			b.Unlock()
		}

		allMatchers := allTopics(msg.Topic) // All possible matches for msg.topic

		b.RLock()
		for _, c := range b.clients {
			go checkDeliverToClient(c, allMatchers, msg)
		}
		b.RUnlock()
	}
}

/*
 * Deliver given message to the given client if one of the topic matchers
 * matches one of the clients subscriptions
 */
func checkDeliverToClient(c *client, allMatchers []string, msg *packet.Message) {
	for _, matcher := range allMatchers {
		c.mutex.Lock()
		if sub, ok := c.subscriptions[matcher]; ok {
			// Re-package the message with the correct QOS (matching the subscription) and retain
			// Retain to false for all normal subscriptions (MQTT-3.3.1-9)
			m := repackage(msg, sub.QOS, false)
			c.deliveryChannel <- m
		}
		c.mutex.Unlock()
	}
}
