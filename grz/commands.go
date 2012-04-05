
package main

import (
	"os"
	"fmt"
	"flag"
	"strings"
	"io/ioutil"

	"garzon/db"
	"garzon/eval"
	"garzon/grz-judge/client"
	prog "garzon/eval/programming"
)


const u_add = `usage: git add [options] <directory>

Options:
  --path    Colon-separated list of directories to consider 
            as roots

`
const u_update = `usage: git update [options] <directory>

Options:
  --path    Colon-separated list of directories to consider 
            as roots

`
const u_submit = `usage: git submit [options] <ProblemID> <filename>

Options:
  --judge    URL for the judge

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
		usageCmd("add", 2)
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

	// Get the <type> of the problem + ID
	base, typ := splitType(relative)
	id = strings.Replace(base, "/", ".", -1)
	
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

	// TODO: Read statement

	Problem = &eval.Problem{
		Title: string(title), 
		StatementID: "",
	}

	// Read directory
	R, ok := ev.(eval.DirReader)
	if ! ok {
		_err("Retrieved object is not a DirReader")
	}
	if err := R.ReadDir(absdir, Problem); err != nil {
		_err("Coudln't read problem '%s': %s\n", id, err)
	}
	E, ok := ev.(eval.Evaluator)
	if ! ok {
		_err("Retrieved object is not an Evaluator")
	}
	Problem.Evaluator = db.Obj{E}
	return
}

func add(args []string) {
	dir := addParseFlags(args)
	
	id, Problem := readProblem(dir)
	
	// Store in the database
	problems, err := db.GetOrCreateDB("problems")
	if err != nil {
		_err("Cannot get db 'problems': %s\n", err)
	}
	rev, _ := problems.Rev(id)
	if rev != "" {
		_err("Problem '%s' already in the database", id)
	}
	if err := problems.Put(id, Problem); err != nil {
		_err("Couldn't add: %s\n", err)
	}
}

func update(args []string) {
	dir := addParseFlags(args)
	
	id, Problem := readProblem(dir)
	
	// Store in the database
	problems, err := db.GetOrCreateDB("problems")
	if err != nil {
		_err("Cannot get database 'problems': %s\n", err)
	}
	rev, _ := problems.Rev(id)
	if rev == "" {
		_err("Problem '%s' not found in the database", id)
	}
	if err := problems.Update(id, rev, Problem); err != nil {
		_err("Couldn't update: %s\n", err)
	}
}

func submit(args []string) {
	var url string
	fset := flag.NewFlagSet("submit", flag.ExitOnError)
	fset.StringVar(&url, "judge", "", "URL for the Judge")
	fset.Parse(args)

	if url != "" {
		client.JudgeUrl = url
	}

	if len(args) != 2 {
		fmt.Fprintf(os.Stderr, "Wrong number of arguments")
		usageCmd("submit", 2)
	}

	id, err := client.Submit(args[0], args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Submission error: %s\n", err)
		os.Exit(2)
	}
	fmt.Printf("ID: %s\n", id)

	status, err := client.Status(id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot get status: %s\n", err)
		os.Exit(2)
	}
	fmt.Printf("Status: %s\n", status)
	
}