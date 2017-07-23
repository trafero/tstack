# What is MQTT?

MQTT is a messaging protocol, especially popular for the Internet of Things.  The "tStack" project uses version 3.1.1 of the MQTT protocol which is [fully documented as an open standard](http://docs.oasis-open.org/mqtt/mqtt/v3.1.1/os/mqtt-v3.1.1-os.html).

A "messaging broker" (the MQTT server) will accept connections from clients who may (authorization permitting) *publish* to a given message topic, or *subscribe* to a message topic or topics.  *Publishing* is the method for sending messages to the MQTT broker, whereas *subscribing* is the method for reading published messages from the MQTT broker. Tstack's [tserve](tserve.md) is a message broker. This project also containers MQTT messaging subscribers and publishers.

A message *topic* is (sort of) the context of the message; for example a message of "23" might not mean anything on its own, but with a *topic* of "/outside/temperature", we can start to understand what that message means.

# What Makes MQTT Special

MQTT is an open standard, so we're all free to use it, and, as long as we conform to the standard, we're far more able to create systems which are compatible with each other.

There are a couple of nice features of MQTT worth mentioning:

* Messages are generally given to subscribers as-and-when they are published. However, *retained* messages are sent to new subscribers as soon as they connect.  This is a very useful feature if a subscriber needs to know the last known state when it starts up. For example, a heating system computer might be restarted, and would need to know if it should turn on or off the heating when first it connects to the message broker.
* *Last Will and Testament* is another useful feature. It can be used to tell the MQTT broker to send a message out when a device disconnects. This can be used, for example to tell subscribers that a device they are interested in has disconnected, and maybe give them one last piece of information.

# Some Uses of MQTT

* Internet of Things - Devices can send and receive messages, which can be picked up by central components, allowing us to build all kinds of connected things!
* Server monitoring - System information, such as CPU and disk space can be sent to a messaging broker. These messages could be read by some code which takes all these readings and puts them into a database, for later analysis. Graphite/Grafana works well for this. OK, I don't know of anyone (else) who does this, but it works really well!
* Building an online chat system, so users can send messages out on different topics, and subscribers can subscribe to those topics.


[Here are some more practical applications](https://github.com/mqtt/mqtt.github.io/wiki/Example-uses).


