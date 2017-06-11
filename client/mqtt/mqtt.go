package mqtt

import (
	paho "github.com/eclipse/paho.mqtt.golang"
	"github.com/trafero/tstack/client/settings"
	"github.com/trafero/tstack/tls"
	"log"
	"time"
)

const (
	qos = 0
)

type Message struct {
	Topic   string
	Payload string
}

type MQTT struct {
	client  paho.Client
	handler func(Message)
	topics  []string // Topics subscribed to
}

func New(s *settings.Settings) (m *MQTT, err error) {

	m = &MQTT{}

	// Create paho MQTT Client
	tlsconfig, err := tls.TLSClientConfig(s.CaCertFile)
	if err != nil {
		return nil, err
	}
	tlsconfig.InsecureSkipVerify = !s.VerifyTls
	opts := paho.NewClientOptions()
	opts.SetClientID(s.Username)
	opts.SetTLSConfig(tlsconfig)
	opts.AddBroker(s.Broker)
	opts.SetDefaultPublishHandler(m.controlMessageHandler)
	opts.SetUsername(s.Username)
	opts.SetPassword(s.Password)
	opts.SetKeepAlive(30 * time.Second)
	opts.SetConnectionLostHandler(m.connectionLostHandler)
	c := paho.NewClient(opts)
	m.client = c

	m.connect()

	return m, nil
}

// NewInsecure sets up a new session without TLS
func NewInsecure(s *settings.Settings) (m *MQTT, err error) {

	m = &MQTT{}

	opts := paho.NewClientOptions()
	opts.SetClientID(s.Username)
	opts.AddBroker(s.Broker)
	opts.SetDefaultPublishHandler(m.controlMessageHandler)
	opts.SetUsername(s.Username)
	opts.SetPassword(s.Password)
	opts.SetKeepAlive(30 * time.Second)
	opts.SetConnectionLostHandler(m.connectionLostHandler)
	c := paho.NewClient(opts)
	m.client = c

	m.connect()

	return m, nil
}

func (m *MQTT) controlMessageHandler(client paho.Client, msg paho.Message) {
	// log.Printf("Received topic: %s message: %s", msg.Topic(), msg.Payload())
	if m.handler != nil {
		message := Message{
			Topic:   string(msg.Topic()),
			Payload: string(msg.Payload()),
		}
		m.handler(message)
	}
}

func (m *MQTT) SetHandler(f func(Message)) {
	m.handler = f
}

// sendMessage sends an MQTT message.
func (m *MQTT) PublishMessage(topic string, payload string) error {
	// log.Printf("Sending topic %s and payload %s", topic, payload)
	token := m.client.Publish(topic, qos, false, payload)
	if token.Error() != nil {
		return token.Error()
	}
	token.Wait()
	return token.Error()
}

func (m *MQTT) Subscribe(topic string) {
	var token paho.Token
	m.topics = append(m.topics, topic)
	token = m.client.Subscribe(topic, byte(0), nil)
	token.Wait()
	if token.Error() != nil {
		log.Printf("WARNING: %s\n", token.Error())
	}
}

func (m *MQTT) connectionLostHandler(c paho.Client, err error) {
	log.Printf("WARNING. Connection lost: %s\n", err)

	m.connect()

	// Re-subscribe to all the topics
	for _, topic := range m.topics {
		m.Subscribe(topic)
	}
}

func (m *MQTT) connect() {

	log.Println("Attempting to connect")

	for {
		token := m.client.Connect()
		if token.Error() != nil {
			log.Printf("Connection failure: %s", token.Error())
		}
		token.Wait()
		if m.client.IsConnected() {
			log.Println("Connected")
			break
		}
		time.Sleep(5 * time.Second)
	}

}
