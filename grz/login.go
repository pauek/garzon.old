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
	saveAuthToken()
	fmt.Println("Ok.")
}
