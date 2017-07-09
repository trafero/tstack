#!/bin/bash -e

VERSION=0.1

cd "$( dirname "${BASH_SOURCE[0]}" )"
PACKAGE_DIR=`pwd`/trafero-tstack-$VERSION
rm -rf $PACKAGE_DIR
mkdir -p $PACKAGE_DIR/DEBIAN


# Post install script
cp postinst $PACKAGE_DIR/DEBIAN/.
chmod +x $PACKAGE_DIR/DEBIAN/postinst

mkdir -p $PACKAGE_DIR/usr/bin
mkdir -p $PACKAGE_DIR/etc/trafero
mkdir -p $PACKAGE_DIR/lib/systemd/system/


# Create executables
GOBIN=$PACKAGE_DIR/usr/bin go install github.com/trafero/tstack/cmd/...

# Configuration file
echo '
# Environment file for tserve.service
OPTS=" --authentication=false --addr=0.0.0.0:1883"
' > $PACKAGE_DIR/etc/trafero/tserve

# Systemd startup
echo '
[Unit]
Description=Trafero tserve MQTT messsage broker

[Service]
Type=simple
User=trafero
EnvironmentFile=/etc/trafero/tserve
ExecStart=/usr/bin/tserve $OPTS
' > $PACKAGE_DIR/lib/systemd/system/tserve.service
 
# Package build config
echo "Package: trafero-tstack
Version: $VERSION
Section: base
Priority: optional
Architecture: amd64
Depends: dh-systemd
Maintainer: Douglas Gibbons <doug@trafero.io>
Description: Trafero tstack MQTT messsage broker and tools
" > $PACKAGE_DIR/DEBIAN/control

# Create the package
dpkg-deb --build $PACKAGE_DIR

echo OK

