package main

import (
	"fmt"
	"garzon/grz-judge/client"
)

const u_login = `grz login <user>`

func login(args []string) {
	login := checkOneArg("login", args)

	// Get password (two times)
	DisableEcho()
	// FIXME: Catch Ctrl-C
	var passwd string
	fmt.Printf("Password: ")
	fmt.Scanf("%s", &passwd)
	fmt.Printf("\n") // no echo -> no endl
	EnableEcho()

	var err error
	if err = client.Login(login, passwd); err != nil {
		_errx("Cannot login: %s", err)
	}

	if err = saveAuthToken(client.AuthToken); err != nil {
		_errx("Cannot save auth token: %s\n", err)
	}

	fmt.Println("Ok.")
}
