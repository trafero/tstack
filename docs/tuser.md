# tuser

tuser registers new uses directly against the etcd key-value store. etcd access should be limited, as should access to tuser. This command line tool is suitable for creating special system users. For creating "normal" users, it is recommended to use [treg](treg.md).

## Usage

The following command line options are available:

```
  -etcdhosts string
    	list of etcd endpoints. e.g. 'http://etcd0:2379 http://etcd1:2379'
  -username string
    	Username for new user
  -password string
    	Password for new user
  -rights string
    	Access rights as topic expression
```


## Access rights

The access rights should follow the format in the [MQTT 3.1.1 specification](http://docs.oasis-open.org/mqtt/mqtt/v3.1.1/errata01/os/mqtt-v3.1.1-errata01-os-complete.html#_Toc442180919).

* Access rights are based on topic
* The same rights apply for both publish and subscribe
* "#" is used as a multi-level wildcard
* "+" is used as a single level wildcard

For example:

* "#" - access to all topics
* "ABC-123/#" - access to all topics starting with "ABC-123/"
* "ABC-123/+/temperature" - access to all topics that start with "ABC-123/", end with "temperature", and have one more level in-between

### Example Usage

The following creates a user called USERNAME, with a password of PASSWORD, and access to all topics.

```
tuser                               \
-etcdhosts=http://localhost:2379    \
-username=USERNAME                  \
-password=PASSWORD                  \
-rights="#"                         \
```