package broker

import (
	"github.com/gomqtt/packet"
	"log"
	"sync"
)

type broker struct {
	clients map[string]*client
	mutex   *sync.Mutex
}

func NewBroker() *broker {
	return &broker{
		mutex:   &sync.Mutex{},
		clients: make(map[string]*client),
	}
}

func (b *broker) AddClient(c *client) {
	
	if existingClient,exists := b.clients[c.clientid]; exists && c.cleanSession == false{
		log.Println("Old client wants another shot")
		// clientid already exists
		c.inboundInTransit = existingClient.inboundInTransit
		c.outboundInTransit = existingClient.outboundInTransit
	}
	b.mutex.Lock()
	b.clients[c.clientid] = c
	b.mutex.Unlock()
}

func (b *broker) receive(msg *packet.Message) {
	log.Printf("Received a message %s", msg)
	for _, c := range b.clients {
		for _, sub := range c.subscriptions {
			if matches(sub.Topic, msg.Topic) {
				log.Printf("Delivering message to client %s", c.clientid)
				go c.send(msg, sub.QOS)
			}
		}
	}
}
