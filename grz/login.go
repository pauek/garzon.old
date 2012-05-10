package main

import (
	"fmt"
)

const u_login = `grz login <user>`

func login(args []string) {
	login := checkOneArg("login", args)

	if isOpen() {
		fmt.Println("No need to login.")
		return
	}

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
	client.SaveAuthToken()
	fmt.Println("Ok.")
}
