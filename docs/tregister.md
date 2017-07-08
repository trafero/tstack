# tregister

tregister is a command line tool which requests a new user using the [treg](treg.d) service then writes the new user's details to a YAML configuration file for use by other applications.

tregister creates a configuration file in __/etc/trafero__, creating this directory if it needs to.

## Command Line Usage

The application takes the following command line options:


```
  -devtype string
        Device type (e.g. testdevice) (default "unknown")
  -regkey string
        Registration key
  -regservice string
        Registration service (e.g. http://localhost:8000/register.json)
  -verifytls
        Verify MQTT server TLS certificate name (default true)
```


## Example

If you're running as a non-privileged user, you'll need to create the configuration directory first:

```
sudo mkdir /etc/trafero && sudo chown $USER /etc/trafero
```

The command line tool can be found in the [trafero/tstack-mqtt](https://hub.docker.com/r/trafero/tstack-mqtt/) docker image.

Change REGISTRATION_KEY and TSERVE_HOST in the following:

```
tregister                                              \
  -regkey=REGISTRATION_KEY                             \
  -regservice=http://TSERVE_HOST:8000/register.json    \
  -verifytls=false
```

/etc/trafero should now contain:

* ```settings.yml``` -  a settings file, containing login details to the tserve MQTT broker
* ```ca.crt``` - A CA certificate that can be used to check the authenticity of the tserve MQTT broker


