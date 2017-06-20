package etcd

import (
	"encoding/json"
	"github.com/coreos/etcd/client"
	"github.com/trafero/tstack/auth"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/net/context"
	"log"
	"time"
)

// String to represent no user rights
const NO_RIGHTS = "^$"

// Authentication with ETCD backend
type Etcd struct {
	etcdConfig client.Config
	etcdClient client.Client
	etcdApi    client.KeysAPI
}

// New returns a pointer to an Auth struct. Endpoints are an array of etcd
// endpoints
func New(endpoints []string) (a *Etcd, err error) {

	cfg := client.Config{
		Endpoints:               endpoints,
		Transport:               client.DefaultTransport,
		HeaderTimeoutPerRequest: time.Second,
	}
	c, err := client.New(cfg)
	if err != nil {
		return nil, err
	}
	a = &Etcd{
		etcdConfig: cfg,
		etcdClient: c,
		etcdApi:    client.NewKeysAPI(c),
	}
	return a, nil
}

// Returns user object for a given username
func (t *Etcd) User(username string) (u auth.User, err error) {
	var resp *client.Response

	// Create an empty user
	u = auth.User{}
	u.Username = username
	u.Rights = NO_RIGHTS

	// Get that user from etcd
	resp, err = t.etcdApi.Get(
		context.Background(),
		"/user/"+username,
		nil,
	)
	if err != nil {
		log.Printf("Could not retrieve user %s from the database: %s", username, err)
		return u, err
	}

	err = json.Unmarshal([]byte(resp.Node.Value), &u)
	return u, err

}

// Saves a user object
func (t *Etcd) setUser(u auth.User) (err error) {

	var userInfo []byte // Text to save to etcd

	userInfo, err = json.Marshal(u)
	if err != nil {
		return err
	}

	_, err = t.etcdApi.Set(
		context.Background(),
		"/user/"+u.Username,
		string(userInfo[:]),
		nil,
	)
	if err != nil {
		log.Printf("Error saving user %s: %s", u.Username, err)
		return err
	}
	log.Printf("User %s saved", u.Username)

	return nil
}

// Authenticate checks the given password with the hashed version stored in etcd
func (t *Etcd) Authenticate(username string, password string) bool {

	u, err := t.User(username)
	if err != nil {
		log.Printf("Error retrtieving user: %s", err)
		return false
	}

	// Password never set
	if u.Password == "" {
		log.Printf("No password set for user: %s", username)
		return false
	}
	// log.Printf("Got: %s", u.Password)

	// Compare hash from ETCD with given password (not hashed)
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
		log.Printf("Passwords do not match")
		return false
	}
	log.Printf("Passwords match for user %s", username)
	return true

}

// AddOrUdpdateUser adds or updates a user's password, storing a hash of the
// password in etcd under /passwd/<usename>
func (t *Etcd) AddOrUpdateUser(username string, password string) (err error) {
	log.Printf("Setting up user %s", username)
	var u auth.User
	u, err = t.User(username)
	u.Password = auth.Hash(password)
	err = t.setUser(u)
	return err
}

// SetRights sets user rights
func (t *Etcd) SetRights(username string, rights string) (err error) {
	log.Printf("Setting up user rights for user %s", username)
	var u auth.User
	u, err = t.User(username)
	if err != nil {
		return err
	}
	u.Rights = rights
	err = t.setUser(u)
	return err
}

// Rights returns a string of allowed access rights
func (t *Etcd) Rights(username string) (rights string) {
	log.Printf("Getting user rights for user %s", username)
	u, _ := t.User(username)
	return u.Rights
}

// UserExists returns true if the user exists and false for anything else
func (t *Etcd) UserExists(username string) bool {
	_, err := t.User(username)
	if err != nil {
		return false
	}
	return true
}
