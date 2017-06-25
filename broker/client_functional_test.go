package broker

import (
	authall "github.com/trafero/tstack/auth/all"
	"github.com/trafero/tstack/client/mqtt"
	"github.com/trafero/tstack/client/settings"
	"net"
	"testing"
)

const thost = "127.0.0.1:9883"

func TestPublish(t *testing.T) {

	go listen(t)

	// Client connection
	s := &settings.Settings{
		Username: "username",
		Password: "password",
		Broker:   "tcp://" + thost,
	}
	m, err := mqtt.NewInsecure(s)
	if err != nil {
		t.Error("Error initiating mqtt client", err)
	}

	t.Log("Publishing MQTT message")
	err = m.PublishMessage("topic", "payload")
	if err != nil {
		t.Error("Error publising message", err)
	}

}

/*
 *  listen sets up a broker
 */
func listen(t *testing.T) {
	a, _ := authall.New()
	b := NewBroker()

	t.Log("Listening for MQTT connections")
	l, err := net.Listen("tcp", thost)
	if err != nil {
		t.Error("Error settinng up listener: %s", err)
	}
	defer l.Close()
	for {
		c, err := l.Accept()
		if err != nil {
			t.Error("Error accepting connection: %s", err)
		}

		client := NewClient(a, b, c)
		go client.HandleConnection()
	}
}
