
package main

import (
	"os"
	"fmt"
	"flag"
	"strings"
	"io/ioutil"

	"garzon/db"
	"garzon/eval"
	prog "garzon/eval/programming"
)

var addCommand Command = Command{
	help: `Add a problem to the Database`,
	usage: _addUsage,
	function: add,
}

var updateCommand Command = Command{
	help: `Update a problem in the Database`,
	usage: _updateUsage,
	function: update,
}

const _addUsage = `usage: git add [options] <directory>

Options:
  --path    Colon-separated list of directories to consider 
            as roots

`
const _updateUsage = `usage: git update [options] <directory>

Options:
  --path    Colon-separated list of directories to consider 
            as roots

`

var addPath string

func init() {
	prog.Register()
}

func addParseFlags(args []string) string {
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

func splitType(dir string) (base, typ string) {
	dot := strings.Index(dir, ".")
	if dot == -1 {
		_err(`Directory should end with ".<type>"`)
	}
	return dir[:dot], dir[dot+1:]
}

func readProblem(dir string) (id string, Problem *eval.Problem) {
	// TODO: Read statement

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
	id = strings.Replace(base, "/", ".", -1)
	
	// fmt.Printf("abs:  %s\n", absdir)
	// fmt.Printf("dir:  %s\n", dir)
	// fmt.Printf("root: %s\n", root)
	// fmt.Printf("ID:   %s\n", id)

	// Lookup <type>.Evaluator
	ev := db.ObjFromType(typ + ".Evaluator")
	if ev == nil {
		_err(`Type '%s.Evaluator' not found`, typ)
	}

	// Read Title
	title, err := ioutil.ReadFile(absdir + "/title")
	if err != nil {
		_err("Cannot read title")
	}

	Problem = &eval.Problem{Title: string(title), StatementID: ""}

	// Read directory
	E, ok := ev.(eval.Evaluator)
	if ! ok {
		_err("Retrieved object is not an Evaluator")
	}
	if err := E.ReadFrom(absdir, Problem); err != nil {
		_err("Coudln't read problem '%s': %s\n", id, err)
	}
	Problem.Evaluator = db.Obj{E}
	return
}

func add(args []string) {
	dir := addParseFlags(args)
	
	id, Problem := readProblem(dir)
	
	// Store in the database
	db, err := db.GetOrCreate("localhost:5984", "problems")
	if err != nil {
		_err("Cannot access database (http://localhost:5984/problems)")
	}
	rev, _ := db.Rev(id)
	if rev != "" {
		_err("Problem '%s' already in the database", id)
	}
	if err := db.Put(id, Problem); err != nil {
		_err("Couldn't add: %s\n", err)
	}
}

func update(args []string) {
	dir := addParseFlags(args)
	
	id, Problem := readProblem(dir)
	
	// Store in the database
	db, err := db.GetOrCreate("localhost:5984", "problems")
	if err != nil {
		_err("Cannot access database (http://localhost:5984/problems)")
	}
	rev, _ := db.Rev(id)
	if rev == "" {
		_err("Problem '%s' not found in the database", id)
	}
	if err := db.Update(id, rev, Problem); err != nil {
		_err("Couldn't update: %s\n", err)
	}
}
