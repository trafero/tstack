# tstack

MQTT Broker with an etcd authentication backend as well as MQTT consumers and tools, all written in Go.

* Requires Go >= 1.8
* CLI applications are under the cmd directory


To quickly bring up an MQTT 3.1.1 broker:

```
docker run trafero/tstack-mqtt tserve --authentication=false -addr=0.0.0.0:1883
```

Full documentation, including how to enable authentication and TLS can be found in [docs](https://github.com/trafero/tstack/tree/master/docs)
