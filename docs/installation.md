# Installation Guide

## Docker

All the binaries can be found in the [trafer/tstack-mqtt](https://hub.docker.com/r/trafero/tstack-mqtt/) container.

For example:

```
docker run trafero/tstack-mqtt tserve --authentication=false -addr=0.0.0.0:1883
```

## Go

If you have a go environment configured you can build and install the binaries with:

```
go get github.com/trafero/tstack/cmd/...
```

## Ubuntu

Add the APT repository:

```
echo "deb [trusted=yes] https://repo.fury.io/zenly/ /" > \
  /etc/apt/sources.list.d/fury.list
```

Install the package (this includes tserve, treg, tuser, etc)

```
apt update && apt install trafero-tstack
```

The Ubuntu package comes with a systemd startup script for [tserve](tserve.md). Configuration for the startup script can be found at ```/etc/trafero/tserve```.

To start the tserve service:

```
service tserve start
```

# Requirements

For authentication, etcd will also need to be installed.
