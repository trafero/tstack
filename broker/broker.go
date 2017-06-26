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
				go c.Send(msg, sub.QOS)
			}
		}
	}
}
