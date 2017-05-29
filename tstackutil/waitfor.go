package tstackutil

import (
	"errors"
	"log"
	"net"
	"net/url"
	"os"
	"strconv"
	"time"
)

var defaultports = map[string]int{
	"http":  80,
	"https": 443,
	"ssh":   22,
}

// WaitForFile waits for a file to exist before continuing
// This is especially useful for situations where the start order of various
// services cannot be easily determined (such as docker-compose)
func WaitForFile(filename string) {
	log.Printf("Waiting for file (%s) to exist", filename)
	for {
		_, err := os.Stat(filename)
		if err == nil {
			log.Printf("%s exists", filename)
			break
		}
		log.Println(err)
		time.Sleep(1000 * time.Millisecond)
	}
}

// WaitForTcp waits for a TCP port to be available before coninuing
func WaitForTcp(hostname string, port int) {
	log.Printf("Waiting for service (%s:%d) to be available.", hostname, port)
	for {
		conn, err := net.Dial("tcp", hostname+":"+strconv.Itoa(port))
		if err == nil {
			conn.Close()
			log.Printf("%s:%d is ready", hostname, port)
			break
		}
		log.Println(err)
		time.Sleep(1000 * time.Millisecond)
	}
}

// WaitForUrl waits for a given TCP service to be listening.
func WaitForUrl(urlstring string) {

	var port int

	u, err := url.Parse(urlstring)
	if err != nil {
		log.Fatal(err)
	}

	// Requires go >= 1.8
	host := u.Hostname()
	portstring := u.Port()

	// If port is not supplied find a default
	if portstring == "" {
		port = defaultports[u.Scheme]
	} else {
		port, err = strconv.Atoi(portstring)
		if err != nil {
			log.Fatal(err)
		}
	}
	if port == 0 {
		log.Fatal(errors.New("Port not known for URL " + urlstring))
	}

	WaitForTcp(host, port)
}
