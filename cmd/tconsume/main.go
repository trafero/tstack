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
	"net/url"
)

var username, password, mqtturl, topic, ctype string
var tlscertfile, tlskeyfile, cacertfile string

var influxhost, influxdatabase string
var influxport int

var graphitehost string
var graphiteport int

var verifytls, useconfig bool

var consumer consume.Consume

const (
	clientid = "consumer"
)

func init() {
	flag.StringVar(&username, "username", "", "Username for MQTT broker")
	flag.StringVar(&password, "password", "", "Password for MQTT broker")
	flag.StringVar(&mqtturl, "mqtturl", "tcp://localhost:1883", "URL for MQTT broker")
	flag.StringVar(&topic, "topic", "", "Topic to subscribe to. Defaults to USERNAME/#")

	flag.StringVar(&influxhost, "influxhost", "localhost", "InfluxDB hostname")
	flag.IntVar(&influxport, "influxport", 8086, "InfluxDB port")
	flag.StringVar(&influxdatabase, "influxdatabase", "device", "InfluxDB database")

	flag.StringVar(&graphitehost, "graphitehost", "localhost", "Graphite hostname")
	flag.IntVar(&graphiteport, "graphiteport", 2003, "Graphite port")

	flag.StringVar(&ctype, "ctype", "", "Consumer type. One of influxdb, graphite, stdout")
	flag.StringVar(&cacertfile, "cacrtfile", "/etc/trafero/ca.crt", "CA Cert file")

	flag.BoolVar(&verifytls, "verifytls", true, "Verify MQTT certificate")
	flag.BoolVar(&useconfig, "useconfig", false, "Use tstack configuration file")

	flag.Parse()

}

func main() {

	var err error
	var secure bool // secure connection or not
	var s *settings.Settings

	// Read settings into s
	if useconfig {
		s, err = settings.Read()
		if err != nil {
			flag.Usage()
			log.Fatal(err)
		}
	} else {

		s = &settings.Settings{
			Username: username,
			Password: password,
			Broker:   mqtturl,
			// Only used for TLS
			CaCertFile:  cacertfile,
			VerifyTls:   verifytls,
		}

	}

	// Set topic to receive everything for that user, if it hasn'y been set
	// already
	if topic == "" {
		topic = s.Username + `/#`
	}

	log.Printf("Using broker %s and topic %s.", s.Broker, topic)

	secure, err = isSecureUrl(s.Broker)
	checkErr(err)

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

	checkErr(err)

	log.Printf("Connecting to broker %s", s.Broker)
	var m *mqtt.MQTT

	if secure {
		m, err = mqtt.New(s)
	} else {
		m, err = mqtt.NewInsecure(s)
	}

	checkErr(err)

	m.SetHandler(consumer.ControlMessageHandler)
	m.Subscribe(topic)

	// Go to forever land
	select {}
}

// isSecureUrl determines if the given URL has a secure scheme type
func isSecureUrl(urlstring string) (bool, error) {
	u, err := url.Parse(urlstring)
	if err != nil {
		return false, err
	}

	switch u.Scheme {
	case "tcp":
		return false, nil
	case "ssl":
		return true, nil
	case "http":
		return false, nil
	case "https":
		return true, nil
	default:
		return false, errors.New("Unknown broker URL type.")
	}
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
