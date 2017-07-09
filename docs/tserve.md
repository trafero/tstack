# tserve

tserve is an MQTT broker with back end authentication using an etcd key-value store.  For adding users see the [treg](treg.md) service, [tregister](tregister.md), and [tuser](tuser.md).


## Limitations

Currently MQTT v. 3.1.1, QoS level 0.


## Command Line Usage

Start the service using these command line options:

```
  -addr string
    	Unencrypted listen address. e.g. 0.0.0.0:1883
  -addrTls string
    	Encrypted listen address. eg. 0.0.0.0:8883
  -cafile string
    	CA certificate (default "/certs/ca.crt")
  -certfile string
    	TLS certificate file (default "/certs/mqtt.crt")
  -etcdhosts string
    	list of etcd endpoints. e.g. 'http://etcd0:2379 http://etcd1:2379'
  -keyfile string
    	TLS key file (default "/certs/mqtt.key")
  -authentication bool (default true)
        Use authentication (default true). etcdhosts is not required if this is set to false.
```

### Example Usage

To run without encryption and no authentication:

```
tserve -addr=0.0.0.0:1883 -authentication=false
```

To run without encryption and using a local etcd key-value store:

```
tserve -addr=0.0.0.0:1883 -etcdhosts=http://localhost:2379
```

