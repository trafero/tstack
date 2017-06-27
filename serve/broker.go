package serve

import (
	"github.com/gomqtt/packet"
	"sync"
)

type Broker struct {
	clients  map[string]*client         // Map by clientid
	retained map[string]*packet.Message // Map by topic
	mutex    *sync.Mutex
}

func NewBroker() *Broker {
	return &Broker{
		mutex:    &sync.Mutex{},
		clients:  make(map[string]*client),
		retained: make(map[string]*packet.Message),
	}
}

func (b *Broker) AddClient(c *client) {

	// Clean session: [MQTT-3.1.2-6]
	if existingClient, exists := b.clients[c.clientid]; exists && c.cleanSession == false {
		// clientid already exists
		c.inboundInTransit = existingClient.inboundInTransit
		c.outboundInTransit = existingClient.outboundInTransit
		c.subscriptions = existingClient.subscriptions
	}
	b.mutex.Lock()
	b.clients[c.clientid] = c
	b.mutex.Unlock()
}

func (b *Broker) deliver(msg *packet.Message) {
	if msg.Retain {
		b.mutex.Lock()
		if len(msg.Payload) == 0 {
			// MQTT-3.3.1-10
			delete(b.retained, msg.Topic)
		} else {
			b.retained[msg.Topic] = msg
		}
		b.mutex.Unlock()
	}
	for _, c := range b.clients {
		for topic, sub := range c.subscriptions {
			if matches(topic, msg.Topic) {
				// Retain to false for all normal subscriptions MQTT-3.3.1-9
				go c.send(msg, sub.QOS, false)
			}
		}
	}
}
