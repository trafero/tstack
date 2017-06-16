package main

import (
	"errors"
	"flag"
	"github.com/trafero/tstack/client/mqtt"
	"github.com/trafero/tstack/client/settings"
	"log"
	"net/url"
)

var username, password, mqtturl, topic, payload, cacertfile string
var verifytls, useconfig bool

const (
	clientid = "tpublish"
)

func init() {
	flag.StringVar(&username, "username", "", "Username for MQTT broker")
	flag.StringVar(&password, "password", "", "Password for MQTT broker")
	flag.StringVar(&mqtturl, "mqtturl", "tcp://localhost:1883", "URL for MQTT broker")
	flag.StringVar(&topic, "topic", "", "Topic to publish to")
	flag.StringVar(&payload, "payload", "", "Payload to publish")
	flag.StringVar(&cacertfile, "cacrtfile", "/etc/trafero/ca.crt", "CA Cert file")
	flag.BoolVar(&verifytls, "verifytls", true, "Verify MQTT certificate")
	flag.BoolVar(&useconfig, "useconfig", true, "Use tstack configuration file")

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
			CaCertFile: cacertfile,
			VerifyTls:  verifytls,
		}

	}

	// Set topic to receive everything for that user, if it hasn'y been set
	// already
	if topic == "" {
		flag.Usage()
		log.Fatal("No topic")
	}

	log.Printf("Publishing topic %s.", topic)

	secure, err = isSecureUrl(s.Broker)
	checkErr(err)

	log.Printf("Connecting to broker %s", s.Broker)
	var m *mqtt.MQTT

	if secure {
		m, err = mqtt.New(s)
	} else {
		m, err = mqtt.NewInsecure(s)
	}

	checkErr(err)

	err = m.PublishMessage(topic, payload)
	checkErr(err)
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
