package main

import (
	"flag"
	"github.com/trafero/tstack/client/settings"
	"io/ioutil"
	"log"
	"os"
)

const (
	settingsFile      = "/etc/trafero/settings.yml"
	settingsDirectory = "/etc/trafero/"
	cacertfile        = "/etc/trafero/ca.crt"
)

var regservice, regkey string
var verifytls bool

func init() {
	flag.StringVar(&regservice, "regservice", "", "Registration service (e.g. http://localhost:8000/register.json)")
	flag.StringVar(&regkey, "regkey", "", "Registration key")
	flag.BoolVar(&verifytls, "verifytls", true, "Verify MQTT server TLS certificate name")
	flag.Parse()
}

func main() {

	var err error
	var s *settings.Settings

	if regservice == "" || regkey == "" {
		flag.Usage()
		log.Fatal(err)
	}

	// Ensure configuration directory is present
	if _, err = os.Stat(settingsDirectory); os.IsNotExist(err) {
		if err = os.MkdirAll(settingsDirectory, os.ModePerm); err != nil {
			log.Fatal(err)
		}
	}

	if _, err := os.Stat(settingsFile); !os.IsNotExist(err) {
		log.Fatal("ERROR: Settings file \"" + settingsFile + "\" already exists")
	}

	if s, err = registerDevice(); err != nil {
		log.Fatal(err)
	}

	log.Printf("Device name %s", s.Username)
	log.Printf("Using broker %s", s.Broker)

}

func registerDevice() (s *settings.Settings, err error) {

	s = &settings.Settings{
		VerifyTls:  verifytls,
		CaCertFile: cacertfile,
	}

	reply, err := Register(regservice, regkey)
	if err != nil {
		return s, err
	}
	if err := ioutil.WriteFile(s.CaCertFile, []byte(reply.Ca), 0644); err != nil {
		return s, err
	}

	s.Username = reply.Name
	s.Password = reply.Password
	s.Broker = reply.Broker

	// Write settings to file
	err = settings.Write(s)
	return s, err
}
