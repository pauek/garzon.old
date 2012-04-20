package main

import (
	"code.google.com/p/go.crypto/bcrypt"
	"fmt"
	"github.com/pauek/garzon/db"
)

const u_adduser = `grz-db adduser <login>`

func adduser(args []string) {
	login := checkOneArg("adduser", args)

	// Get DB
	users, err := db.GetOrCreateDB("users")
	if err != nil {
		_errx("Cannot connect to 'users' database: %s\n", err)
	}

	// Check if user exists
	rev, err := users.Rev(login)
	if err != nil {
		_errx("Cannot get rev for user '%s'", login, err)
	}
	if rev != "" {
		_errx("User '%s' is already in the database", login)
	}

	// Disable terminal echo
	ToggleEcho()
	defer ToggleEcho()

	// Get password (two times)
	var passwd [2]string
	for i := range passwd {
		if i == 1 {
			fmt.Printf("(repeat) ")
		}
		fmt.Printf("Password: ")
		fmt.Scanf("%s", &passwd[i])
		fmt.Printf("\n") // no echo -> no endl
	}
	if passwd[0] != passwd[1] {
		_errx("Passwords do not match")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(passwd[0]), 10)
	if err != nil {
		_errx("Cannot hash password: %s\n", err)
	}

	err = users.Put(login, &db.User{
		Login:   login,
		Hpasswd: string(hash),
	})
	if err != nil {
		_errx("Cannot save user '%s': %s\n", login, err)
	}

	fmt.Printf("Ok\n")
}
