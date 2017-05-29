package main

import (
	"database/sql"
	"flag"
	influx "github.com/influxdata/influxdb/client/v2"
	_ "github.com/lib/pq"
	"github.com/trafero/tstack/client/mqtt"
	"github.com/trafero/tstack/client/settings"
	"log"
	"strconv"
	"strings"
	"time"
)

var db *sql.DB
var influxClient influx.Client
var username, password, influxhost, influxport, mqtturl, influxdatabase, topic string

const (
	clientid = "consumer"
)

func init() {
	flag.StringVar(&username, "username", "", "Username for MQTT broker")
	flag.StringVar(&password, "password", "", "Password for MQTT broker")
	flag.StringVar(&influxhost, "influxhost", "localhost", "InfluxDB hostname")
	flag.StringVar(&influxport, "influxport", "8086", "InfluxDB port")
	flag.StringVar(&influxdatabase, "influxdatabase", "device", "InfluxDB database")
	flag.StringVar(&mqtturl, "mqtturl", "tcp://localhost:1883", "URL for MQTT broker")
	flag.StringVar(&topic, "topic", "#", "Topic to subscribe to")
	flag.Parse()
}

func main() {

	var err error

	log.Printf("Using broker %s", mqtturl)

	err = InitInfluxdb()
	checkErr(err)

	s := &settings.Settings{
		Username: username,
		Password: password,
		Broker:   mqtturl,
	}
	m, err := mqtt.NewInsecure(s)
	checkErr(err)

	m.SetHandler(controlMessageHandler)
	m.Subscribe(topic)

	// Go to forever land
	select {}
}

func controlMessageHandler(msg mqtt.Message) {
	// log.Printf("Received topic: %s message: %s\n", msg.Topic, msg.Payload)
	err := sendToInfluxdb(msg.Topic, string(msg.Payload))
	if err != nil {
		log.Printf("ERROR: %s", err)
	}
}

func sendToInfluxdb(topic string, payload string) (err error) {
	// Create a new point batch
	bp, err := influx.NewBatchPoints(influx.BatchPointsConfig{
		Database:  influxdatabase,
		Precision: "ms",
	})
	if err != nil {
		return err
	}

	fields := map[string]interface{}{}

	// Split the topic by / to form database fields
	// i.e. ONE/TWO/THREE becomes
	// t0=ONE t1=TWO, t3=THREE
	topicSplit := strings.Split(topic, "/")
	for i := 0; i < len(topicSplit); i++ {
		fields["t"+strconv.Itoa(i)] = topicSplit[i]
	}

	// Add payload as a database field
	fields["payload"] = payload

	pt, err := influx.NewPoint("reading", nil, fields, time.Now())
	if err != nil {
		return err
	}
	bp.AddPoint(pt)

	// Write the batch
	if err := influxClient.Write(bp); err != nil {
		return err
	}

	return nil
}

func InitInfluxdb() (err error) {
	influxClient, err = influx.NewHTTPClient(influx.HTTPConfig{
		Addr: "http://" + influxhost + ":" + influxport,
	})
	if err != nil {
		return err
	}

	// Create database just in case it does nto exist already
	q := influx.Query{
		Command:  "CREATE DATABASE " + influxdatabase,
		Database: influxdatabase,
	}
	_, err = influxClient.Query(q)
	if err != nil {
		return err
	}

	return nil

}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
