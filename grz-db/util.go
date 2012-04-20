package main

import (
	"fmt"
	"os"
)

func _err(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "grz-db: "+format+"\n", args...)
}

func _errx(format string, args ...interface{}) {
	_err(format, args...)
	os.Exit(2)
}

func checkNArgs(n int, cmd string, iargs []string) (oargs []string) {
	if len(iargs) != n {
		_err("Wrong number of arguments")
		usageCmd(cmd, 2)
	}
	return iargs
}

func checkOneArg(cmd string, args []string) string {
	return checkNArgs(1, cmd, args)[0]
}

func checkTwoArgs(cmd string, iargs []string) (a, b string) {
	oargs := checkNArgs(2, cmd, iargs)
	return oargs[0], oargs[1]
}
