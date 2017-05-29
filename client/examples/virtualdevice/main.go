package main

import (
	"github.com/trafero/tstack/client/mqtt"
	"github.com/trafero/tstack/client/settings"
	"log"
	"math"
	"strconv"
	"time"
)

const (
	ports = 2
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

	for i := 0; i < ports; i++ {
		go analog(i)
	}

	// Wait forever
	select {}
}

func analog(port int) {
	var reading float64

	reading = math.Sin(math.Pi / float64(ports) * float64(port))
	for {
		mq.PublishMessage(s.Username+"/analog/"+strconv.Itoa(port), strconv.FormatFloat(reading, 'f', 4, 64))
		reading = reading + math.Sin(math.Pi/100)
		time.Sleep(5 * time.Second)
	}
}
