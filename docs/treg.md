# treg

treg is an HTTP service to register users. A random username in the form "ABC-123" is created, along with a random password string.  The etcd backend is checked during username creation to ensure that the uesrname is not already used.

An authorization string along with the username and a bcrypt hash of the password are stored in the etcd backend. The authorization string defines access only topics whose first level is the new username.  This means that each new user can only read and write their own messages.

For wider authentication options, use the [tuser.md](tuser.md) command line tool.

## Command Line Usage

Start the service using these command line options:

```
  -cacertfile string
    	CA certificate location (may be blank if not required)
  -etcdhosts string
    	list of etcd endpoints. e.g. 'http://etcd0:2379 http://etcd1:2379'
  -mqtturl string
    	URL for MQTT broker (default "tcp://localhost:1883")
  -port string
    	Port to listen on (default "8000")
  -regkey string
    	Registration key
```

The registration key should be a random string of alpha-numeric characters. The same string should be given to users of treg for authentication.


## API Usage

Once the application is running, a RESTful service is available on port 8000

### POST /register.json

Creates users for tstack.

### Request

A JSON request of the following required values:


* RegistrationKey - Authentication key (see regkey command line option)
* DeviceType - Short string description of the device, so that common devices can be grouped together

### Response

A JSON response of the following values:

* Name - New username
* Password - New password
* Broker - URL of the broker (tserve) service. See mqtt url above.
* Ca - TLS certificate authority key of the broker service. This may be blank if not set above

### Example

Curl request:

```
curl -X POST \
  --data '{"RegistrationKey": "TODO_REGISTRATION_PW", "DeviceType": "test"}' \
   http://HOSTNAME:8000/register.json
```

Response:
```
{"Name":"ABC-123","Password":"PASSWORD","Broker":"ssl://HOSTNAM:8883","Ca":""}
```
