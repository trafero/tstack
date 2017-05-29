package main

import (
	"github.com/trafero/tstack/client/mqtt"
	"github.com/trafero/tstack/client/settings"
	"log"
)

const (
	ports = 4
)

var mq *mqtt.MQTT
var s *settings.Settings

func main() {

	var err error

	// Read settings from config file
	s, err = settings.Read()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Device name %s", s.Username)
	log.Printf("Device type is %s", s.DeviceType)
	log.Printf("Using broker %s", s.Broker)

	mq, err = mqtt.New(s)
	if err != nil {
		panic(err)
	}

	// Set handler
	mq.SetHandler(handler)

	go subscribe()

	// Wait forever
	select {}
}

func handler(msg mqtt.Message) {
	log.Printf("Got %s, %s", msg.Topic, msg.Payload)
}
func subscribe() {
	topic := s.Username + "/#"
	mq.Subscribe(topic)
}
