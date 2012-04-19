package main

import (
	"fmt"
)

func help(args []string) {
	if len(args) == 0 {
		usage(0)
	}
	for _, cmd := range args {
		help1(cmd)
	}
}

func help1(cmd string) {
	for _, C := range commands {
		if C.name == cmd {
			fmt.Printf("%s\n", C.usage)
			return
		}
	}
	_errx("unknown command '%s'\n", cmd)
}
