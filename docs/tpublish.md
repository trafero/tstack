# tpublish

tpublish publishes messages to an MQTT 3.1.1 broker


## Command Line Usage

To publish a message:

```
  -cacrtfile string
    	CA Cert file (default "/etc/trafero/ca.crt")
  -mqtturl string
    	URL for MQTT broker (default "tcp://localhost:1883")
  -password string
    	Password for MQTT broker
  -username string
    	Username for MQTT broker
  -verifytls
    	Verify MQTT certificate (default true)
  -useconfig
    	Use tstack configuration file (default true)
  -payload string
    	Payload to publish
  -topic string
    	Topic to publish to
```

## Example Usage

To publish a message, using the configuration file created by [tregister](tregister.md):

```
tpublish --topic=ABC-123/test --payload=Test
```

__Note that standard users created by tregister can only write to topics that start with their username.__
