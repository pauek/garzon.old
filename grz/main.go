package main

import (
	"fmt"
	_ "garzon/eval/programming"
	"os"
)

type Command struct {
	name     string
	help     string
	usage    string
	function func(args []string)
}

var commands []*Command

func init() {
	commands = []*Command{
		&Command{"add", `Add a problem to the Database`, u_add, add},
		&Command{"update", `Update a problem in the Database`, u_update, update},
		&Command{"delete", `Delete a problem in the Database`, u_delete, delette},
		&Command{"submit", `Submit a problem to the judge`, u_submit, submit},
		&Command{"help", ``, "", help},
	}
}

const _usage_header = "usage: grz <command> [<args>]\n\nCommands:\n"
const _usage_footer = `
Environment: 
  GRZ_PATH    List of colon-separated roots for problems
  GRZ_JUDGE   URL of the Judge (including port)

See 'grz help <command>' for more information.
`

func findCmd(cmd string) *Command {
	for _, C := range commands {
		if C.name == cmd {
			return C
		}
	}
	return nil
}

func usage(exitcode int) {
	fmt.Fprint(os.Stderr, _usage_header)
	for _, cmd := range commands {
		if cmd.name != "help" {
			fmt.Fprintf(os.Stderr, "  %-12s%s\n", cmd.name, cmd.help)
		}
	}
	fmt.Fprint(os.Stderr, _usage_footer)
	os.Exit(exitcode)
}

func usageCmd(cmd string, exitcode int) {
	C := findCmd(cmd)
	if C == nil {
		panic(fmt.Sprintf("command '%s' not found", cmd))
	}
	fmt.Fprint(os.Stderr, C.usage)
	os.Exit(exitcode)
}

func main() {
	if len(os.Args) < 2 {
		usage(2)
	}
	cmd := os.Args[1]
	if C := findCmd(cmd); C != nil {
		C.function(os.Args[2:])
	} else {
		if cmd == "--help" {
			usage(0)
		} else {
			_errx("grz: '%s' is not a grz command. See 'grz help'", cmd)
		}
	}
}
