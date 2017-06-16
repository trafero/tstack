package main

import (
	nettls "crypto/tls"
	"flag"
	etcdauth "github.com/trafero/tstack/auth/etcd"
	"github.com/trafero/tstack/mqtt"
	"github.com/trafero/tstack/tls"
	"github.com/trafero/tstack/tstackutil"
	"log"
	"net"
	"runtime"
	"strings"
)

var addr, addrTls, etcdhosts, certfile, keyfile, cafile string

func init() {

	flag.StringVar(&addr, "addr", "", "Unencrypted listen address. e.g. 0.0.0.0:1883")
	flag.StringVar(&addrTls, "addrTls", "", "Encrypted listen address. eg. 0.0.0.0:8883")
	flag.StringVar(&etcdhosts, "etcdhosts", "", "list of etcd endpoints. e.g. 'http://etcd0:2379 http://etcd1:2379'")
	flag.StringVar(&certfile, "certfile", "/certs/mqtt.crt", "TLS certificate file")
	flag.StringVar(&keyfile, "keyfile", "/certs/mqtt.key", "TLS key file")
	flag.StringVar(&cafile, "cafile", "/certs/ca.crt", "CA certificate")
	flag.Parse()
}

func main() {

	// Check command line arguments
	if etcdhosts == "" {
		flag.Usage()
		log.Fatal("etcdhosts argument missing")
	}

	if addr == "" && addrTls == "" {
		flag.Usage()
		log.Fatal("addr and addrTls cannot both be missing")
	}

	// Authentication using ETCD
	log.Printf("Using etcd hosts: %s", etcdhosts)
	a, err := etcdauth.New(strings.Split(etcdhosts, " "))
	checkErr(err)

	// MQQT subscriptions service (used for both of the MQTT servers)
	subs := mqtt.NewSubscriptions(runtime.GOMAXPROCS(0))

	// Unencrypted MQTT server
	if addr != "" {
		log.Printf("Running MQTT server on %s", addr)
		l, err := net.Listen("tcp", addr)
		checkErr(err)
		svr1 := mqtt.NewServer(a, l, subs)
		go svr1.Start()
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
		lTls, err := nettls.Listen("tcp", addrTls, tlsconfig)
		checkErr(err)
		svr2 := mqtt.NewServer(a, lTls, subs)
		svr2.Start()
	}

	// Wait forever
	select {}

}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
