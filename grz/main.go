package main

import (
	"fmt"
	_ "github.com/pauek/garzon/eval/programming"
	"os"
)

type Command struct {
	name     string
	help     string
	usage    string
	function func(args []string)
}

var commands []*Command
var authToken string

func init() {
	commands = []*Command{
		&Command{"login", `Login to the Judge`, u_login, login},
		&Command{"logout", `Logout from the Judge`, u_logout, logout},
		&Command{"passwd", `Change your Judge's password`, u_passwd, passwd},
		&Command{"list", `List all problems`, u_list, list},
		&Command{"search", `Search for problems`, u_search, search},
		&Command{"show", `Show the problem's statement`, u_show, show},
		&Command{"submit", `Submit a problem to the judge`, u_submit, submit},
		// grz config??
		&Command{"help", ``, "", help},
	}
}

const _usage_header = "usage: grz <command> [<args>]\n\nCommands:\n"
const _usage_footer = `
Environment: 
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
	fmt.Fprintf(os.Stderr, "usage: %s\n", C.usage)
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
