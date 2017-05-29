package main

import (
	"github.com/trafero/tstack/auth"
	"log"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letters = []rune("ABCDEFGHJKLMNPQRSTUVWXYZ") // No "I" or "O"
var numbers = []rune("23456789")                 // No One or Zero
var pwdchars = []rune("abcdefghijklmnopqrstuvqxyzABCDEFGHIJKLMNOPQRSTUVWXYZ12345678790-_#,.!$%^&()[]{}")

func randString(n int, pool []rune) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = pool[rand.Intn(len(pool))]
	}
	return string(b)
}

// NewId generates a new name for a device
func NewId(authService auth.Auth) (string, error) {

	var id string
	for {

		// Create randon ID
		id = randString(3, letters) + "-" + randString(3, numbers)

		// Check if it already exists
		if !authService.UserExists(id) {
			break
		}

		// It does exist so round and round we go
		log.Printf("ID %s was not as unique as we hoped. Trying another.\n", id)

		// Be kind to all the other services
		time.Sleep(1000 * time.Millisecond)
	}
	return id, nil

}

func NewPassword() string {
	return randString(24, pwdchars)
}
