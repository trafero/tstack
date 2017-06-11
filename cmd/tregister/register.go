package main

import (
	"bytes"
	"encoding/json"
	"github.com/trafero/tstack/tstackutil"
	"log"
	"net/http"
	"strings"
	"time"
)

type Post struct {
	RegistrationKey string
	DeviceType      string
}
type Reply struct {
	Name     string
	Password string
	Broker   string
	Ca       string
}

func Register(registrationService string, registrationKey string, deviceType string) (reply Reply, err error) {

	var resp *http.Response
	reply = Reply{}

	log.Printf("Connecting to registration service %s", registrationService)
	tstackutil.WaitForUrl(registrationService)

	post := Post{
		RegistrationKey: registrationKey,
		DeviceType:      deviceType,
	}
	post_data, err := json.Marshal(post)
	if err != nil {
		return reply, err
	}
	for {
		resp, err = http.Post(
			registrationService,
			"application/json; charset=utf-8",
			bytes.NewBuffer(post_data),
		)
		if err == nil {
			break
		}
		log.Printf("Error from registration service%s\n", err)
		time.Sleep(time.Second)
	}

	defer resp.Body.Close()
	if !strings.HasPrefix(resp.Status, "200") {
		log.Printf("Got a response code of %s. Hoping for a 200", resp.Status)
	}
	err = json.NewDecoder(resp.Body).Decode(&reply)
	if err != nil {
		return reply, err
	}
	log.Println("Registered a device name of", reply.Name)

	return reply, nil

}
