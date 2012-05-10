package main

import (
	"fmt"
)

const u_passwd = `grz passwd`

func passwd(args []string) {
	checkZeroArgs("passwd", args)

	// Determine user

	if isOpen() {
		fmt.Println("No password in this Judge.")
		return
	}
	_errx("Unimplemented.")
}
