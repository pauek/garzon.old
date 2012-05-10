package main

import (
	"fmt"
)

const u_logout = `grz logout <user>`

func logout(args []string) {
	logout := checkOneArg("logout", args)
	
	if err := client.MaybeReadAuthToken(); err != nil {
		_errx("%s", err)
	}

	if err := client.Logout(logout); err != nil {
		_errx("Cannot logout: %s", err)
	}
	if err := client.RemoveAuthToken(); err != nil {
		_errx("Cannot remove auth token: %s")
	}
	
	fmt.Println("Ok.")
}
