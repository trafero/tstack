# Performance

_Notes on Performance tuning and function/non-functional testing_

------------------

## Tuning for performance

### ulimit

You may well need to up the file handle ulimit, so that tserver can accept a decent number of simultaneous connections:

```
ulimit -Sn 50000
```

### Connections

If you're submitting data regularly to tserve, it's a good idea to keep the connection open.  This saves on doing more connection and authentication negotiations than you need.

### Topic Names

Possible matches for a given topic looks something like this:

* Topic: one/two
* Matches: one/two,+/two,one/+,+/+,one/#,+/#,#

The number of matches is 2 to the power of n, where n is the number of topics filters.  i.e. keep the number of topic filters low.

## Performance Statistics from tserve

tserve publishes its profiling data (using [pprof](https://golang.org/pkg/net/http/pprof/) over http. This can be accessed from [http://localhost:8070/debug/pprof/](http://localhost:8070/debug/pprof/).

## Functional Testing

The Eclipse Paho project has put together a very useful test pack for MQTT, v 3.1.1

First, start tserve:

```
tserve -authentication=false -addr=0.0.0.0:1883
```

Run the functional test pack from [paho.mqtt.testing](https://github.com/eclipse/paho.mqtt.testing):
```
python3 client_test.py localhost:1883
git clone git@github.com:eclipse/paho.mqtt.testing.git
cd paho.mqtt.testing/interoperability/
python3 client_test.py localhost:1883
```

If all tests are _not_ passing, then we have a problem!


## Load Testing With Gatling


Here's an example [gatling](http://gatling.io/) test script (don't forget to change the TODO_HOSTNAME for the hostname).  This is a tough test, as each user publishes and subscribes (subscriptions being the hard part).  If you want to go easy on tserve, simply comment out the subscriptions. You'll get some great performance figures, but not such a realistic test.

It runs without authentication, so tserve should be started [tserve](tserve.md) also without authentication:

```
tserve -authentication=false -addr=0.0.0.0:1883
```
[Wireshark](https://www.wireshark.org/) is a good way to tell if the test is behaving itself. You should see the same amount of publishes as subscriptions. Also, as the machine running tserve starts to queue up processes, memory use will start to rise. This is a good sign that it has reached its limit.

```
import com.github.jeanadrien.gatling.mqtt.Predef._
import io.gatling.core.Predef._
import scala.util.Random
import scala.concurrent.duration._


class MqttScenarioExample extends Simulation {
    val mqttConf = mqtt.host("tcp://TODO_HOSTNAME:1883").version("3.1.1")
	val feeder = Iterator.continually(
		Map("topic" -> (Random.alphanumeric.take(5).mkString + "/" + Random.alphanumeric.take(5).mkString))
	)
    val scn = scenario("MQTT Test")
    	.feed(feeder)
        .exec(connect)
        .exec(subscribe("${topic}").qosAtMostOnce)
        .during(30 minutes) {
            pace(1 second).exec(
                publish("${topic}", "myPayload".getBytes()).qosAtMostOnce // qosExactlyOnce
            )
        }
    setUp(
        scn.inject(rampUsers(3000) over (2 minutes)))
        .protocols(mqttConf)
}
```