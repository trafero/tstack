package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/didip/tollbooth"
	"github.com/trafero/tstack/auth"
	"github.com/trafero/tstack/auth/etcd"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

var authService auth.Auth
var mqtturl, regkey, etcdhosts, port, cacertfile string

func init() {
	flag.StringVar(&regkey, "regkey", "", "Registration key")
	flag.StringVar(&etcdhosts, "etcdhosts", "", "list of etcd endpoints. e.g. 'http://etcd0:2379 http://etcd1:2379'")
	flag.StringVar(&mqtturl, "mqtturl", "tcp://localhost:1883", "URL for MQTT broker")
	flag.StringVar(&port, "port", "8000", "Port to listen on")
	flag.StringVar(&cacertfile, "cacertfile", "", "CA certificate location")
	flag.Parse()
}

func main() {
	var err error
	if etcdhosts == "" || regkey == "" {
		flag.Usage()
		log.Fatal("Incorrect command line arguments")
	}
	authService, err = etcd.New(strings.Split(etcdhosts, " "))
	checkErr(err)

	log.Printf("Listening for registration requests on port %s", port)

	http.Handle(
		"/register.json",
		tollbooth.LimitFuncHandler(tollbooth.NewLimiter(5, 20*time.Second), register))
	http.ListenAndServe(":"+port, nil)
}

type registration_reply struct {
	Name     string
	Password string
	Broker   string
	Ca       string
}
type registration_request struct {
	RegistrationKey string
	DeviceType      string
}

func register(w http.ResponseWriter, r *http.Request) {
	LogRequest(r)

	decoder := json.NewDecoder(r.Body)
	var req_data registration_request

	// Get request data
	err := decoder.Decode(&req_data)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("Bad registration request: ", err)
		return
	}

	// Check registration key
	if req_data.RegistrationKey != regkey {
		w.WriteHeader(http.StatusUnauthorized)
		log.Println("Unauthorized request key: ", req_data.RegistrationKey)
		return
	}

	// Create a new device ID
	id, err := NewId(authService)
	checkErr(err)

	password := NewPassword()

	err = authService.AddOrUpdateUser(id, password)
	checkErr(err)

	err = authService.SetRights(id, id+`/#`)
	checkErr(err)

	err = authService.SetGroup(id, req_data.DeviceType)
	checkErr(err)

	// Read TLS certs into struct for output
	ca := ""
	if cacertfile != "" {
		cabytes, err := ioutil.ReadFile(cacertfile)
		checkErr(err)
		ca = string(cabytes)
	}
	// Output
	w.Header().Set("Content-Type", "application/json")

	regData := registration_reply{
		Name:     id,
		Broker:   mqtturl,
		Ca:       ca,
		Password: password,
	}
	js, err := json.Marshal(regData)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(js)
}

func LogRequest(r *http.Request) {
	url := fmt.Sprintf("%v %v %v", r.Method, r.URL, r.Proto)
	log.Printf("%s\n", url)
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
