# tstack

Trafero Stack = tstack

* Requires Go >= 1.8
* CLI applications are under the cmd directory

## Getting Started

### Services

The fastest way to bring up a working stack is by using [the docker-compose file](https://github.com/trafero/tstack/blob/master/docker-compose.yml).

1. Copy the [env.dist file](https://github.com/trafero/tstack/blob/master/env.dist) to ```.env``` and modify it as required
1. Run ```docker-compose up```
1. Check the logs with ```docker-compose logs -f```

Now you have a working MQTT broker, listening on ports 1883 (insecure) and 8883 (secure).  You also have a registration service (called treg) running on port 8000, and an etcd service which is used to hold user data.

### Adding a normal user

Described here is a manual method for adding a use, simply to demonstrate the treg service. Usually the tregister command line tool would be used to simplify this process, saving the new user's configuration and the certificates for later use.

A call to the registration service creates a user and adds that user to the etcd authentication and authorization database (for use by tserve).  Each user (which could be a device in the world of IoT) is given rights to only access message topics that start with their own username.

Here is an example of a request to the service:

```
echo '{
	"RegistrationKey": "PLEASE_CHANGE_ME_TOO",
	"DeviceType": "Test"
}' |  curl -d @- localhost:8000/register.json
```

This returns some JSON that contains:

* ```Name``` - a new username
* ```Password``` - a new password (a bycrypt hasj of the password is stored in etcd)
* ```Broker``` - the URL of the broker service to connect to
* ```Cert``` - a TLS certificate that the client can use
* ```Key``` - a TLS key that the client can use
* ```Ca``` - a CA certificate that the client can use to authenticate the Broker


The user details can also been seen in etcd, as in the example below, which runs a command in the existing etcd container. Replace USER below with the Name from the returned JSON:

```
docker-compose exec etcd etcdctl get /user/USER
```

### Adding a user with greater privileges

The tuser command line tool can create users with custom privileges.  This is useful for creating users for tools such as consumers that read all messages and process the data for back end storage.

The tuser binary exists in the same image used for tserve, so can user docker-compose to bring up a new tserve container and run the tuser command.  Here's an example, creating a user called "MASTER" with a password of "PLEASE_CHANGE_ME", which you should, of course change to something more appropriate!

```
docker-compose run tserve /go/bin/tuser \
  -etcdhosts=http://etcd.tstack_default:2379 \
  -username=MASTER \
  -password=PLEASE_CHANGE_ME \
  -rights='.*'
```

## Securing The Services

Some of Tstack is designed to be internet facing. Some should not be.

* tserve (the MQTT broker) should have it's secure port (8883) exposed, but not its insecure port
* treg is an HTTP service. It should sit behind a proxy serving just HTTPS, such as [an NGINX server that proxies traffic from HTTPS to HTTP](https://hub.docker.com/r/dougg/nginx-letsencrypt-proxy/)
* the etcd service should not be public facing

## Connecting to the MQTT service

tregister or tuser can be used to create a user, but how do you use that new user to access the MQTT service?

If you wish to use golang, there are some examples in the src/client/examples directory, which make use of the configuration file provided by tregister. There are also a couple of read-made consumers in the src/consumer directory.

Other applications such as [Mosquitto](https://mosquitto.org/) can also be used, since tserve is a standard offers a standard MQTT service.

Here is an example of using Mosquitto to subscribe to all messages. Note that we're using the insecure communication port which should not be exposed beyond the docker host.
```
mosquitto_sub \
 -h localhost \
 -p 1883 \
 -v \
 -u MASTER -P PLEASE_CHANGE_ME \
 -t \# \
 -V mqttv311
```

Here is an example using the secure port, and the certificate that was created in /etc/trafero. The USERNAME and PASSWORD settings should be changed to the values found in /etc/trafero/settings.yml.

```
mosquitto_pub \
 --cafile /etc/trafero/ca.crt \
 --cert /etc/trafero/client.crt \
 --key /etc/trafero/client.key \
 -h localhost \
 -p 8883 \
 -u USERNAME -P 'PASSWORD' \
 -V mqttv311 \
 --insecure \
 -t CPW-639/test/message \
 -m "Hello world"
```

## Applications

### treg

treg is a registration service, which listens for HTTP requests for registrations and creates new uses for the registrations in etcd.

Each user is given permissions to only access message topics that start with its own user name.

### tserve

tserve is an MQTT broker, which uses etcd as its backend authentication service.

### tuser

tuser creates a user on the etcd service. Access permissions can be specified on the command line.

### tconsume

MQTT consumer with multiple backends. Currently implemented:

* stdout
* influxdb
* graphite

### tregister

tregister requests a new user from the registration service, then writes the new user's details to a YAML configuration file for use by other applications.  The registration service also gives it a client TLS certificate and CA certificate which applications can use to verify the identity of the tserve service.

tregister creates a configuration file in __/etc/trafero__, creating this directory if it needs to. If you're running as a non-privileged user, you'll need to create the configuration directory first:

```
sudo mkdir /etc/trafero && sudo chown $USER /etc/trafero
```
Now run tregister, as below, changing TSERVE_HOST for the hostname of the tserve service, and PLEASE_CHANGE_ME_TOO to the registration key found in your .env file.

```
docker run -v /etc/trafero:/etc/trafero trafero/tstack \
  tregister                                            \
  -regkey=PLEASE_CHANGE_ME_TOO                         \
  -regservice=http://TSERVE_HOST:8000/register.json    \
  -verifytls=false
```

/etc/trafero should now contain:

* ```settings.yml``` -  a settings file, containing login details to the tserve MQQT broker
* ```ca.crt``` - A CA certificate that can be used to check the authenticity of the tserve MQQT broker
*  ```client.crt``` - A TLS certificate, should a client service need one to connect to the broker
*  ```client.key``` - A TLS key, should the client need one
