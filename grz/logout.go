package main

import (
	"fmt"
	"garzon/grz-judge/client"
)

const u_logout = `grz logout <user>`

func logout(args []string) {
	logout := checkOneArg("logout", args)
	readAuthToken()

	if err := client.Logout(logout); err != nil {
		_errx("Cannot logout: %s", err)
	}
	if err := removeAuthToken(); err != nil {
		_errx("Cannot remove auth token: %s")
	}
	
	fmt.Println("Ok.")
}
