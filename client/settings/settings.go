package settings

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
)

const (
	settingsFile = "/etc/trafero/settings.yml"
)

// Settings stored in config file
type Settings struct {
	Username    string
	Password    string
	Broker      string
	CaCertFile  string
	DeviceType  string
	VerifyTls   bool
}

func Read() (s *Settings, err error) {
	log.Printf("Reading settings file %s", settingsFile)

	s = &Settings{}
	if _, err = os.Stat(settingsFile); err != nil {
		return s, err
	}
	data, err := ioutil.ReadFile(settingsFile)
	if err != nil {
		return s, err
	}
	err = yaml.Unmarshal([]byte(data), s)
	return s, err
}

func Write(s *Settings) (err error) {
	log.Printf("Writing settings file %s", settingsFile)

	d, err := yaml.Marshal(&s)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(settingsFile, d, 0644)
	return err
}
