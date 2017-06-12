# tconsume

tconsume is a command-line MQTT consumer. It is designed to run with [tserve](tserve.md), but could just as easily be used with over MQTT brokers.

A typical use case would be to consume data from an MQTT broker and save that data into a database.


## Limitations

MQTT v. 3.1.1

## Command Line Usage

Command line options are as follows:

```
  -ctype string
    	Consumer type. One of influxdb, graphite, stdout
  -username string
    	Username for MQTT broker
  -password string
    	Password for MQTT broker
  -mqtturl string
    	URL for MQTT broker (default "tcp://localhost:1883")
  -verifytls
    	Verify MQTT certificate (default true)
  -topic string
    	Topic to subscribe to. Defaults to USERNAME/#
  -useconfig
    	Use tstack configuration file
  -cacrtfile string
    	CA Cert file (default "/etc/trafero/ca.crt")
  -graphitehost string
    	Graphite hostname (default "localhost")
  -graphiteport int
    	Graphite port (default 2003)
  -influxdatabase string
    	InfluxDB database (default "device")
  -influxhost string
    	InfluxDB hostname (default "localhost")
  -influxport int
    	InfluxDB port (default 8086)

```

* Set useconfig to "true" to use a tstack configuration file (see [tregister](tregister.md))
* tconsume uses the mqtturl URL string to determine if a secure connection is required (URLs starting with "ssl"). TLS options; cacertfile and verifytls are only used for secure connections.
* graphiteport and grahitehost are only required when the ctype is set to "graphite"
* influxdatabse, influxhost and influxport are only required when the ctype is set to "influxdb"


## Example Usage

The following consumes all topics and prints the messages to STDOUT.

```
tconsume                                          \
  -username=USERNAME                              \
  -password=PASSWORD                              \
  -mqtturl=ssl://MY_HOST:8883                     \
  -ctype=stdout                                   \
  -topic="#"
```
