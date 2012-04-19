package main

import (
	"fmt"
	"garzon/grz-judge/client"
)

const u_logout = `grz logout <user>`

func logout(args []string) {
	logout := checkOneArg("logout", args)

	var err error
	client.AuthToken, err = readAuthToken()
	if err != nil {
		_errx("Cannot read Auth Token: %s", err)
	}

	if err = client.Logout(logout); err != nil {
		_errx("Cannot logout: %s", err)
	}
	if err := removeAuthToken(); err != nil {
		_errx("Cannot remove auth token: %s")
	}
	
	fmt.Println("Ok.")
}
