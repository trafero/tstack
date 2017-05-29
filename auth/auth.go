package auth

// User struct uses typical user access naming conventions to try and make this
// generic, however it can also fit the needs of devices
type User struct {
	Username string // Device name
	Password string // Password hash
	Group    string // Group may be device type (grouping)
	Rights   string // User rights - may be string reg ex for topics
}

type Auth interface {
	// Return a user struct for the given user
	User(username string) (u User, err error)

	// Authenticate checks the given password with the stored version
	Authenticate(username string, password string) bool

	// AddOrUdpdateUser adds or updates a user's password
	AddOrUpdateUser(username string, password string) (err error)

	// Set the user group (there is only one group)
	SetGroup(username string, group string) (err error)

	// Retrieve the group name
	Group(username string) (group string)

	// Set user rights
	SetRights(username string, rights string) (err error)

	// Retrieve the user rights
	Rights(username string) (rights string)

	// Check that a given username exists
	UserExists(username string) bool
}
