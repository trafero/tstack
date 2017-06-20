package all

import (
	"github.com/trafero/tstack/auth"
)

// All is a test implementation, offering full access for everyone, and some
// rather non-existant means of saving users
type All struct {
}

const ALL_RIGHTS = ".*"

func New(endpoints []string) (a *All, err error) {
	a = &All{}
	return a, nil
}

func (t *All) User(username string) (u auth.User, err error) {
	u = auth.User{}
	u.Username = username
	u.Rights = ALL_RIGHTS
	return u, err
}

func (t *All) setUser(u auth.User) (err error) {
	return nil
}

func (t *All) Authenticate(username string, password string) bool {
	return true

}

func (t *All) AddOrUpdateUser(username string, password string) (err error) {
	return nil
}

func (t *All) SetRights(username string, rights string) (err error) {
	return nil
}

func (t *All) Rights(username string) (rights string) {
	return ALL_RIGHTS
}

func (t *All) UserExists(username string) bool {
	return true
}
