package main

import (
	"flag"
	"fmt"
	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/chip"
	"github.com/trafero/tstack/client/examples/keypad/matrix"
	"github.com/trafero/tstack/client/mqtt"
	"github.com/trafero/tstack/client/settings"
	"log"
	"time"
)

var signalPin = 48 // 100 is  PD4 (UART2-CTS) on CHIP Pro
var colPins = []int{132, 133, 134, 135}
var rowPins = []int{136, 137, 138, 139}

func init() {
	flag.IntVar(&signalPin, "sigpin", signalPin, "Signal pin")
	flag.Parse()
}

func main() {

	var err error

	// Read settings from config file
	s, err := settings.Read()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Device name %s", s.Username)
	log.Printf("Device type is %s", s.DeviceType)
	log.Printf("Using broker %s", s.Broker)

	m, err := mqtt.New(s)
	if err != nil {
		panic(err)
	}

	if err = embd.InitGPIO(); err != nil {
		panic(err)
	}
	defer embd.CloseGPIO()

	signal(signalPin)

	keypad, err := matrix.New(rowPins, colPins)
	if err != nil {
		panic(err)
	}

	for {
		key, err := keypad.PressedKey()
		if err != nil {
			panic(err)
		}
		if key != matrix.KNone {
			fmt.Printf("Key Pressed = %v\n", key)
			go signal(signalPin)
			go sendKey(m, s.Username, key.String())
		}

		time.Sleep(500 * time.Millisecond)
	}
}

func sendKey(m *mqtt.MQTT, username string, value string) {
	err := m.PublishMessage(username+"/key", value)
	if err != nil {
		panic(err)
	}
}

// signal turns on the given pin
func signal(pin int) {

	signalPin, err := embd.NewDigitalPin(pin)
	if err = signalPin.SetDirection(embd.Out); err != nil {
		panic(err)
	}
	if err = signalPin.Write(embd.Low); err != nil {
		panic(err)
	}
	if err := signalPin.Write(embd.High); err != nil {
		panic(err)
	}
	time.Sleep(300 * time.Millisecond)
	if err := signalPin.Write(embd.Low); err != nil {
		panic(err)
	}
}
