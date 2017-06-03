package influxdb

import (
	"database/sql"
	influx "github.com/influxdata/influxdb/client/v2"
	_ "github.com/lib/pq"
	"github.com/trafero/tstack/client/mqtt"
	"log"
	"strconv"
	"strings"
	"time"
)

type Influxdb struct {
	influxhost     string
	influxport     int
	influxdatabase string

	influxClient influx.Client
	db           *sql.DB
}

func New(
	influxhost string,
	influxport int,
	influxdatabase string,
) (c *Influxdb, err error) {

	c = &Influxdb{
		influxhost:     influxhost,
		influxport:     influxport,
		influxdatabase: influxdatabase,
	}

	err = c.InitInfluxdb()

	return c, err

}

func (c *Influxdb) ControlMessageHandler(msg mqtt.Message) {
	// log.Printf("Received topic: %s message: %s\n", msg.Topic, msg.Payload)
	err := c.sendToInfluxdb(msg.Topic, string(msg.Payload))
	if err != nil {
		log.Printf("ERROR: %s", err)
	}
}

func (c *Influxdb) sendToInfluxdb(topic string, payload string) (err error) {
	// Create a new point batch
	bp, err := influx.NewBatchPoints(influx.BatchPointsConfig{
		Database:  c.influxdatabase,
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
	if err := c.influxClient.Write(bp); err != nil {
		return err
	}

	return nil
}

func (c *Influxdb) InitInfluxdb() (err error) {
	c.influxClient, err = influx.NewHTTPClient(influx.HTTPConfig{
		Addr: "http://" + c.influxhost + ":" + strconv.Itoa(c.influxport),
	})
	if err != nil {
		return err
	}

	// Create database just in case it does nto exist already
	q := influx.Query{
		Command:  "CREATE DATABASE " + c.influxdatabase,
		Database: c.influxdatabase,
	}
	_, err = c.influxClient.Query(q)
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
