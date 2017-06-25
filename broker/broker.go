package broker

import (
	"github.com/gomqtt/packet"
	"log"
	"sync"
)

type broker struct {
	clients []*client
	mutex   *sync.Mutex
}

func NewBroker() *broker {
	return &broker{
		mutex: &sync.Mutex{},
	}
}

func (b *broker) AddClient(c *client) {
	b.mutex.Lock()
	b.clients = append(b.clients, c)
	b.mutex.Unlock()
}

func (b *broker) receive(m *packet.Message) {
	log.Printf("Received a message %s", m)
	for _, c := range b.clients {
		for _, s := range c.subscriptions {
			if matches(s.Topic, m.Topic) {
				log.Printf("Delivering message to client %s", c.clientid)
				go c.Send(m, s.QOS)
			}
		}
	}
}
