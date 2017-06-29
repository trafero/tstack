package main

import (
	nettls "crypto/tls"
	"flag"
	"github.com/trafero/tstack/auth"
	authall "github.com/trafero/tstack/auth/all"
	etcdauth "github.com/trafero/tstack/auth/etcd"
	"github.com/trafero/tstack/serve"
	"github.com/trafero/tstack/tls"
	"github.com/trafero/tstack/tstackutil"
	"log"
	"net"
	"strings"

	"net/http"
	_ "net/http/pprof"
)

var addr, addrTls, etcdhosts, certfile, keyfile, cafile string
var authentication bool

var broker *serve.Broker
var authenticator auth.Auth

func init() {

	flag.StringVar(&addr, "addr", "", "Unencrypted listen address. e.g. 0.0.0.0:1883")
	flag.StringVar(&addrTls, "addrTls", "", "Encrypted listen address. eg. 0.0.0.0:8883")
	flag.StringVar(&etcdhosts, "etcdhosts", "", "list of etcd endpoints. e.g. 'http://etcd0:2379 http://etcd1:2379'")
	flag.StringVar(&certfile, "certfile", "/certs/mqtt.crt", "TLS certificate file")
	flag.StringVar(&keyfile, "keyfile", "/certs/mqtt.key", "TLS key file")
	flag.StringVar(&cafile, "cafile", "/certs/ca.crt", "CA certificate")
	flag.BoolVar(&authentication, "authentication", true, "Use authentication")
	flag.Parse()
}

func main() {

	// For pprof profiler (at http://localhost:8070/debug/pprof/ )
	go http.ListenAndServe("localhost:8070", nil) // For pprof

	var err error

	if addr == "" && addrTls == "" {
		flag.Usage()
		log.Fatal("addr and addrTls cannot both be missing")
	}

	if authentication {
		if etcdhosts == "" {
			flag.Usage()
			log.Fatal("etcdhosts argument missing")
		}
		// Authentication using ETCD
		log.Printf("Using etcd hosts: %s", etcdhosts)
		authenticator, err = etcdauth.New(strings.Split(etcdhosts, " "))
		checkErr(err)
	} else {
		// Authentication using dummy authenticator which allows all and gives
		// everyone '#' rights (not access to topics starting with $)
		authenticator, _ = authall.New()
	}

	// MQTT broker back end
	broker = serve.NewBroker()

	// Unencrypted MQTT server
	if addr != "" {
		log.Printf("Running MQTT server on %s", addr)
		l, err := net.Listen("tcp", addr)
		checkErr(err)
		go handleServer(l)
		defer l.Close()
	}

	// Encrypted MQTT server
	if addrTls != "" {

		// Wait for the certificates to be created (by another service)
		tstackutil.WaitForFile(cafile)
		tstackutil.WaitForFile(certfile)
		tstackutil.WaitForFile(keyfile)

		log.Printf("Running encypted MQTT server on %s", addrTls)
		tlsconfig, err := tls.TLSConfig(cafile, certfile, keyfile)
		checkErr(err)
		l, err := nettls.Listen("tcp", addrTls, tlsconfig)
		checkErr(err)
		go handleServer(l)
		defer l.Close()
	}

	// Wait forever
	select {}

}

func handleServer(l net.Listener) {

	for {
		connection, err := l.Accept()
		checkErr(err)
		client := serve.NewClient(authenticator, broker, connection)
		go client.HandleConnection()
	}
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
