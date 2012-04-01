
package main

import (
	"os"
	"fmt"
	"flag"
	"strings"
	"garzon/db"
	"garzon/eval"
	prog "garzon/eval/programming"
)

var addCommand Command = Command{
	help: `Add a problem to the Database`,
	usage: _addUsage,
	function: add,
}

const _addUsage = `usage: git add [options] <ProblemID>

Options:
  --path    Colon-separated list of directories to consider 
            as roots

`

var addPath string

func init() {
	prog.Register()
}

func addParse(args []string) string {
	fset := flag.NewFlagSet("add", flag.ExitOnError)
	fset.StringVar(&addPath, "path", "", "Problem path (colon separated)")
	fset.Parse(args)

	if addPath == "" {
		addPath = os.Getenv("GRZ_PATH")
	}
	
	// TODO: Check that no path in 'addPath' is prefix of the others!

	args = fset.Args()
	if len(args) != 1 {
		fmt.Fprint(os.Stderr, "Wrong number of arguments\n")
		usageCommand("add", 2)
	}

	// remove trailing '/'
	dir := args[0]
	if dir[len(dir)-1] == '/' {
		dir = dir[:len(dir)-1]
	}
	return dir
}

func _err(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format + "\n", args...)
	os.Exit(1)
}

func splitType(dir string) (base, typ string) {
	dot := strings.Index(dir, ".")
	if dot == -1 {
		_err(`Directory should end with ".<type>"`, dir)
	}
	return dir[:dot], dir[dot+1:]
}

func add(args []string) {
	dir := addParse(args)
	fmt.Printf("path: %v\n", splitPath(addPath))
	
	// Change to absolute path
	absdir := dir
	cwd, err := os.Getwd()
	if err != nil {
		_err("Cannot get current working directory")
	}
	if absdir[0] != '/' {
		absdir = cwd + "/" + dir
	}

	// Check that it is a directory
	info, err := os.Stat(absdir) 
	if err != nil {
		_err("Cannot stat '%s'", absdir)
	}
	if ! info.IsDir() {
		_err("'%s' is not a directory", absdir)
	}

	// Find the root
	var root, relative string
	for _, path := range splitPath(addPath) {
		if strings.HasPrefix(absdir, path) {
			root, relative = path, absdir[len(path)+1:]
			break
		}
	}
	if root == "" {
		if dir[0] != '/' {
			root, relative = cwd, dir
		} else {
			_err("Root directory not found")
		}
	}

	// Get the <type> of the problem
	base, typ := splitType(relative)

	// Get ID
	id := strings.Replace(base, "/", ".", -1)
	
	fmt.Printf("abs:  %s\n", absdir)
	fmt.Printf("dir:  %s\n", dir)
	fmt.Printf("root: %s\n", root)
	fmt.Printf("ID:   %s\n", id)

	// Lookup <type>.Evaluator
	ev := db.ObjFromType(typ + ".Evaluator")
	if ev == nil {
		_err(`Type '%s.Evaluator' not found`, typ)
	}

	Problem := &eval.Problem{Title: id, StatementID: ""}

	// Read directory
	E, ok := ev.(eval.Evaluator)
	if ! ok {
		_err("Retrieved object is not an Evaluator")
	}
	if err := E.ReadFrom(absdir, Problem); err != nil {
		_err("Coudln't read problem '%s': %s\n", id, err)
	}
	Problem.Evaluator = db.Obj{E}
	
	// Store in the database
	db, err := db.GetOrCreate("localhost:5984", "problems")
	if err != nil {
		_err("Cannot access database (http://localhost:5984/problems)")
	}
	if err := db.PutOrUpdate(id, Problem); err != nil {
		_err("Couldn't store in the database: %s\n", err)
	}
}
