# Quick Start Guide

A [docker image](https://hub.docker.com/r/trafero/tstack/) contains the tstack components. Using docker-compose with this image is the fastest way to get started.

Note that this stack does not use a secure (encrypted) MQTT connection, so is not suitable for live use, but is a great way to get started locally.

For a more complete stack with encryption, check out [../docker-compose.yml.dist](../docker-compose.yml.dist) and it's associated [env file](../env.dist).

## 1: Docker Compose

Use the [docker compose file](docker-compose-quickstart.yml) (which must be renamed to docker-compose.yml) to bring up the MQTT broker stack.

The registration key is set to REGKEY_TO_CHANGE in the docker-compose file. This should be changed.

The compose file brings up a stack of:

* etcd back end storage for user credentials
* [tserve](tserve.md) MQTT broker
* [treg](treg.md) RESTful registration service for new users


## 2: Create and Use a Superuser

A superuser may be used to consume messages and publish them to a database, or perhaps, route messages from one user to another.

To test, we can start an interactive bash session, using the tstack docker image. This will give us access to all the tstack commands.

```
docker-compose run tserve bash
```

The [tuser](tuser.md) command can be used to create a new user if it has access to etcd (which is why etcd should not be exposed on a public interface). Here we create a new user with access to all MQTT topics:

```
tuser \
  --etcdhosts=http://etcd:2379  \
  --username=CONSUMER           \
  --password=PASSWORD           \
  --rights="#"
```

This user can now connect to the [MQTT server](tserve.md) and start consuming messages. [tconsume](tconsume.md) can write to different database formats, but here we just write all messages to STDOUT:

```
tconsume \
  --ctype=stdout \
  --mqtturl=tcp://tserve:1883  \
  --username=CONSUMER          \
  --password=PASSWORD          \
  --topic='#'
```

### Create and Use a Normal User

The [treg](treg.md) service is used to create users with limited access rights; they can only read and write topics that start with their user id.  This should be all that a normal device needs.

To test our platform, we can start another interactive bash session (while leaving the consumer running, so we can see messages coming through):

```
docker-compose run tserve bash
```

At the docker image prompt, register a new user with the [treg](treg.md) service using the [tregister](tregister.md) command line tool. Change the REGKEY_TO_CHANGE value if you changed it in the docker-compose file.

```
tregister --regkey=REGKEY_TO_CHANGE --regservice=http://treg:8000/register.json
```

This will create a configuration file called /etc/trafero/settings.yml containing  access information for a new tserve user:

```
cat /etc/trafero/settings.yml
```

[tpublish](tpublish.md) can use of this configuration file to determine how to connect to the MQTT broker.  Here's an example of using tpublish to write a single message:

```
USERNAME=$(grep username /etc/trafero/settings.yml | awk '{{print $2}}')
tpublish --topic=$USERNAME/temperature --payload=42
```

If you still have the consumer running you will see this message printed out, showing that you have a fully working MQTT message broker, with authentication.
