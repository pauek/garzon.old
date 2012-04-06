
package main

import (
	"os"
	"fmt"
)

type Command struct {
	help  string
	usage string
	function func(args []string)	
}

var commands map[string]*Command

func init() {
	commands = map[string]*Command{
		"add":    &Command{`Add a problem to the Database`,    u_add,    add},
		"update": &Command{`Update a problem in the Database`, u_update, update},
		"delete": &Command{`Delete a problem in the Database`, u_delete, delette},
		"submit": &Command{`Submit a problem to the judge`,    u_submit, submit},
	}
}

const _usage_header = "usage: grz <command> [<args>]\n\nCommands:\n"
const _usage_footer= "\nSee 'grz help <command>' for more information.\n"

func usage(exitcode int) {
	fmt.Fprint(os.Stderr, _usage_header)
	for id, cmd := range commands {
		fmt.Fprintf(os.Stderr, "  %-10s%s\n", id, cmd.help)
	}
	fmt.Fprint(os.Stderr, _usage_footer)
	os.Exit(exitcode)
}

func usageCmd(cmd string, exitcode int) {
	fmt.Fprint(os.Stderr, commands[cmd].usage)
	os.Exit(exitcode)
}

func main() {
	if len(os.Args) < 2 {
		usage(2)
	}
	cmd := os.Args[1]
	if C, ok := commands[cmd]; ok {
		C.function(os.Args[2:])
	} else {
		fmt.Fprintf(os.Stderr, "grz: '%s' is not a grz command. See 'grz --help'\n", cmd)
		os.Exit(2)
	}
}