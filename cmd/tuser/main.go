package main

import (
	"flag"
	etcdauth "github.com/trafero/tstack/auth/etcd"
	"log"
	"strings"
)

var etcdhosts, username, password, rights string

func init() {
	flag.StringVar(&etcdhosts, "etcdhosts", "", "list of etcd endpoints. e.g. 'http://etcd0:2379 http://etcd1:2379'")
	flag.StringVar(&username, "username", "", "Username for new user")
	flag.StringVar(&password, "password", "", "Password for new user")
	flag.StringVar(&rights, "rights", "", "Access rights as topic expression")
	flag.Parse()
}

func main() {

	if etcdhosts == "" || username == "" || password == "" || rights == "" {
		flag.Usage()
		log.Fatal("Incorrect command line arguments")
	}

	log.Printf("Setting up user %s", username)
	log.Printf("ETCD hosts %s", etcdhosts)

	a, err := etcdauth.New(strings.Split(etcdhosts, " "))
	checkErr(err)

	err = a.AddOrUpdateUser(username, password)
	checkErr(err)

	err = a.SetRights(username, rights)
	checkErr(err)
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
