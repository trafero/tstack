package main

import (
	"errors"
	"flag"
	"github.com/trafero/tstack/client/mqtt"
	"github.com/trafero/tstack/client/settings"
	"github.com/trafero/tstack/consume"
	"github.com/trafero/tstack/consume/graphite"
	"github.com/trafero/tstack/consume/influxdb"
	"github.com/trafero/tstack/consume/stdout"
	"log"
)

var username, password, mqtturl, topic, ctype string

var influxhost, influxdatabase string
var influxport int

var graphitehost string
var graphiteport int

var consumer consume.Consume

const (
	clientid = "consumer"
)

func init() {
	flag.StringVar(&username, "username", "", "Username for MQTT broker")
	flag.StringVar(&password, "password", "", "Password for MQTT broker")
	flag.StringVar(&mqtturl, "mqtturl", "tcp://localhost:1883", "URL for MQTT broker")
	flag.StringVar(&topic, "topic", "#", "Topic to subscribe to")

	flag.StringVar(&influxhost, "influxhost", "localhost", "InfluxDB hostname")
	flag.IntVar(&influxport, "influxport", 8086, "InfluxDB port")
	flag.StringVar(&influxdatabase, "influxdatabase", "device", "InfluxDB database")

	flag.StringVar(&graphitehost, "graphitehost", "localhost", "Graphite hostname")
	flag.IntVar(&graphiteport, "graphiteport", 2003, "Graphite port")

	flag.StringVar(&ctype, "ctype", "", "Consumer type. One of influxdb, graphite, stdout")

	flag.Parse()
}

func main() {

	var err error

	log.Printf("Using broker %s", mqtturl)

	switch ctype {
	case "influxdb":
		consumer, err = influxdb.New(
			influxhost,
			influxport,
			influxdatabase,
		)

	case "graphite":
		consumer, err = graphite.New(
			graphitehost,
			graphiteport,
		)

	case "stdout":
		consumer, err = stdout.New()
	default:
		err = errors.New("Please use a valid consumer type (ctype option)")
	}

	if err != nil {
		flag.Usage()
		log.Fatal(err)
	}

	// TODO add ability to use secure message handler
	s := &settings.Settings{
		Username: username,
		Password: password,
		Broker:   mqtturl,
	}
	log.Printf("Connecting to broker %s", s.Broker)
	m, err := mqtt.NewInsecure(s)
	checkErr(err)

	m.SetHandler(consumer.ControlMessageHandler)
	m.Subscribe(topic)

	// Go to forever land
	select {}
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
