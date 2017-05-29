# tstack

Trafero Stack = tstack

* Requires Go >= 1.8
* CLI applications are under the cmd directory

## 2 Minute Guide

## Services

The fastest way to bring up a working stack is by using docker-compose.

1. Copy the env.dist file to ```.env``` and modify it as required
1. Run ```docker-compose up```
1. Check the logs with ```docker-compose logs -f```

Now you have a working MQTT broker, listening on ports 1883 (insecure) and 8883 (secure).  You also have a registration service (called treg) running on port 8000.

## Adding a user

TODO

## Connecting a device

TODO

## Applications

### treg

treg is a registration service, which listens for HTTP requests for registrations and creates new uses for the registrations in etcd.

Each user is given permissions to only access message topics that start with its own user name.

### tserve

tserve is an MQTT broker, which uses etcd as its backend authentication service.

### tuser

tuser creates a user on the etcd service. Access permissions can be specified on the command line.

### tregister

tregister requests a new user from the registration service, then writes the new user's details to a YAML configuration file for use by other applications.  The registration service also gives it a client TLS certificate and CA certificate (which applications can use to verify the identity of the tserve service.

tregister creates a configuration file in /etc/trafero, creating this directory if it needs to. If you're running as a non-privileged used, you'll need to create the configuration directory first:

```
sudo mkdir /etc/trafero && sudo chown $USER /etc/trafero
```
Now run tregister, for example:

```
tregister \
-regkey=PLEASE_CHANGE_ME_TOO \
-regservice=http://localhost:8000/register.json \
-verifytls=false
```