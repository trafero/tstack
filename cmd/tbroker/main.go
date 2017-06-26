package main

import (
	authall "github.com/trafero/tstack/auth/all"
	"github.com/trafero/tstack/broker"
	"log"
	"net"
)

/*
 *  listen sets up a broker
 */
func main() {
	a, _ := authall.New()
	b := broker.NewBroker()

	log.Println("Listening for MQTT connections")
	l, err := net.Listen("tcp", "0.0.0.0:1883")
	checkErr(err)

	defer l.Close()
	for {
		c, err := l.Accept()
		checkErr(err)
		client := broker.NewClient(a, b, c)
		go client.HandleConnection()
	}
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
