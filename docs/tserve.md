# tserve

tserve is an MQTT broker with back end authentication using an etcd key-value store.  For adding users see the [treg](treg.md) service, [tregister](tregister.md), and [tuser](tuser.md).


## Limitations

Currently MQTT v. 3.1.1, QoS level 0.


## Command Line Usage

Start the service using these command line options:

```
  -addr string
    	Unencrypted listen address. Can be blank to disable (default "0.0.0.0:1883")
  -addrTls string
    	Encrypted listen address. Can be blank to disable (default "0.0.0.0:8883")
  -etcdhosts string
    	list of etcd endpoints. e.g. 'http://etcd0:2379 http://etcd1:2379'
  -cafile string
    	CA certificate (default "/certs/ca.crt")
  -certfile string
    	TLS certificate file (default "/certs/mqtt.crt")
  -keyfile string
    	TLS key file (default "/certs/mqtt.key")
```

### Example Usage

To run without encryption (addrTls set to blank), and using a local etcd key-value store:

```
tserve -addrTls="" -etcdhosts=http://localhost:2379
```
