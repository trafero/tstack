# Quick Start Guide

A docker image is build that contains the tstack components. Using this is the fastest way to get started.

## Building a local stack

Over the internet, an MQTT broker with authentication should always be set up with encryption. However, for local use, it's more simple to set it up without.

The following docker containers are required:

### Create a network

The hosts of the various services will need to to each other so we should create a network for them:

```
docker network create --driver bridge tstack
```

### Create an etcd service

An etcd cluster is required as the back end for user storage.

```
docker run --net tstack --name=etcd0 -d quay.io/coreos/etcd:v3.1.8 \
      /usr/local/bin/etcd             \
      -name etcd0                     \
      -advertise-client-urls http://127.0.0.1:2379 \
      -listen-client-urls http://0.0.0.0:2379      \
      -initial-advertise-peer-urls http://127.0.0.1:2380 \
      -listen-peer-urls http://0.0.0.0:2380              \
      -initial-cluster-token etcd-cluster-1              \
      -initial-cluster etcd0=http://127.0.0.1:2380       \
      -initial-cluster-state new                         \
      -data-dir=/etcd
```

### Start a treg service

[treg](treg.md) is the registration service, to add users to etcd. Here we create it, listening on port 8000 for new registration requests. "regkey" is set to "REGKEY". This key should is shared with anyone (or anything) that wishes to sign up to use tserve by using the treg service.

```
docker run              \
  --net=tstack          \
  --name=treg0          \
  -d                    \
  -p 8000:8000          \
  trafero/tstack        \
  treg                  \
    --etcdhosts=http://etcd0:2379 \
    --mqtturl=tcp://tserve0:1883  \
    --regkey=REGKEY               \
    --port=8000
```

### Run the tserve MQTT broker

[tserve](tserve.md) is the MQTT broker.  Here we listen on port 1883 for insecure (not encrypted) MQTT requests.

```
docker run        \
  --net=tstack    \
  --name=tserve0  \
  -p 1883:1883    \
  -d              \
  trafero/tstack  \
  tserve --etcdhosts=http://etcd0:2379 --addr=0.0.0.0:1883
```

### Create and Use a Superuser

A superuse may be used to consume messages and publish them to a database, or perhaps, route messages from one user to another.

To test, we can start an interactive bash session, using the tstack docker image. This will give us access to all the tstack commands, and put us on the same docker network as tserve, treg and etcd.

```
docker run -it   \
  --net=tstack   \
  trafero/tstack \
  bash
```

The [tuser](tuser.md) command can be used to create a new user if it has access to etcd (which is why etcd should not be exposed on a public interface). Here we create a new user with access to all MQTT topics:

```
tuser \
  --etcdhosts=http://etcd0:2379 \
  --username="CONSUME"          \
  --password="PASSWORD"         \
  --rights="#"
```

This user can now connect to the [MQTT server](tserve.md) and start consuming messages. [tconsume](tconsume.md) can write to different database formats, but here we just write all messages to STDOUT:

```
tconsume \
  --ctype=stdout \
  --mqtturl=tcp://tserve0:1883 \
  --username=CONSUMER          \
  --password=PASSWORD          \
  --topic='#'
```

### Create and Use a Normal User

The [treg](treg.md) service is used to create users with limited access rights; they can only read and write topics that start with their user id.  This should be all that a normal device needs.

To test our platform, we can start another interactive bash session:

```
docker run -it   \
  --net=tstack   \
  trafero/tstack \
  bash
```


At the docker image prompt, register against the treg service, using the [tregister](tregister.md) command line tool:

```
tregister --regkey=REGKEY --regservice=http://treg0:8000/register.json
```

This will create a configuration file called /etc/trafero/settings.yml containing  access information for a new tserve user:

```
cat /etc/trafero/settings.yml
```

[tpublish](tpublish.md) can make use of this configuration file. Make a note of the username from the configuration file, and change it in the example below. This will publish a message on the MQTT broker:

```
tpublish --topic=ABC-123/temperature --payload=42
```

If you still have the consumer running, you will see this message printed out. This proves that you have a working MQTT service.
