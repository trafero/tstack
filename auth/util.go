package auth

import (
	"golang.org/x/crypto/bcrypt"
	"log"
)

// hash returns a bcrypt "hash" of the given string
func Hash(s string) (hashed string) {
	hash, err := bcrypt.GenerateFromPassword([]byte(s), bcrypt.DefaultCost)
	// TODO handle error
	if err != nil {
		log.Fatal(err)
	}
	return string(hash)
}
