package main

import (
	"flag"
	"github.com/marpaia/graphite-golang"
	"github.com/trafero/tstack/client/mqtt"
	"github.com/trafero/tstack/client/settings"
	"github.com/trafero/tstack/tstackutil"
	"log"
)

var username, password, graphitehost, mqtturl, topic string
var graphiteport int
var Graphite *graphite.Graphite

const (
	clientid = "consumer"
)

func init() {
	flag.StringVar(&username, "username", "", "Username for MQTT broker")
	flag.StringVar(&password, "password", "", "Password for MQTT broker")
	flag.StringVar(&graphitehost, "graphitehost", "localhost", "Graphite hostname")
	flag.IntVar(&graphiteport, "graphiteport", 2003, "Graphite port")
	flag.StringVar(&mqtturl, "mqtturl", "tcp://localhost:1883", "URL for MQTT broker")
	flag.StringVar(&topic, "topic", "#", "Topic to subscribe to")
	flag.Parse()
}

func main() {

	var err error

	log.Printf("Using broker %s", mqtturl)

	// Graphite
	tstackutil.WaitForTcp(graphitehost, graphiteport)
	Graphite, err = graphite.NewGraphite(graphitehost, graphiteport)
	checkErr(err)

	s := &settings.Settings{
		Username: username,
		Password: password,
		Broker:   mqtturl,
	}

	m, err := mqtt.NewInsecure(s)
	checkErr(err)

	m.SetHandler(controlMessageHandler)
	m.Subscribe(topic)

	// Wait forever
	select {}
}

func controlMessageHandler(msg mqtt.Message) {
	// log.Printf("Received topic: %s message: %s", msg.Topic, msg.Payload)
	Graphite.SimpleSend(msg.Topic, msg.Payload)
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
