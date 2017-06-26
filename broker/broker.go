package broker

import (
	"github.com/gomqtt/packet"
	"log"
	"sync"
)

type broker struct {
	clients  map[string]*client         // Map by clientid
	retained map[string]*packet.Message // Map by topic
	mutex    *sync.Mutex
}

func NewBroker() *broker {
	return &broker{
		mutex:    &sync.Mutex{},
		clients:  make(map[string]*client),
		retained: make(map[string]*packet.Message),
	}
}

func (b *broker) AddClient(c *client) {

	// Clean session: [MQTT-3.1.2-6]
	if existingClient, exists := b.clients[c.clientid]; exists && c.cleanSession == false {
		log.Println("Old client wants another shot")
		// clientid already exists
		c.inboundInTransit = existingClient.inboundInTransit
		c.outboundInTransit = existingClient.outboundInTransit
		c.subscriptions = existingClient.subscriptions
	}
	b.mutex.Lock()
	b.clients[c.clientid] = c
	b.mutex.Unlock()
}

func (b *broker) deliver(msg *packet.Message) {
	log.Printf("Delivering message %s", msg)
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
				log.Printf("Delivering message %s to client %s", msg.Topic, c.clientid)
				// Retain to false for all normal subscriptions MQTT-3.3.1-9
				go c.send(msg, sub.QOS, false)
			}
		}
	}
}
