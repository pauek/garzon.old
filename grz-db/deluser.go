package main

import (
	"fmt"
	"garzon/db"
)

const u_deluser = `grz deluser <login>`

func deluser(args []string) {
	login := checkOneArg("deluser", args)

	users, err := db.GetDB("users")
	if err != nil {
		_errx("Cannot connect to 'users' database: %s\n", err)
	}

	// Find revision
	rev, err := users.Rev(login)
	if err != nil {
		_errx("Cannot get rev for user '%s'", login, err)
	}
	if rev == "" {
		_errx("User '%s' not in the database", login)
	}

	// Confirm
	var login2 string
	fmt.Printf("Confirm username: ")
	fmt.Scanf("%s", &login2)
	if login != login2 {
		_errx("'%s' and '%s' do not match", login, login2)
	}

	// Delete
	err = users.Delete(login, rev)
	if err != nil {
		_errx("Cannot delete user '%s': %s\n", login, err)
	}

	fmt.Printf("Ok\n")
}
