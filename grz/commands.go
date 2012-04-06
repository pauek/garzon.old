
package main

import (
	"os"
	"fmt"
	"flag"
	"time"
	"strings"
	"io/ioutil"

	"garzon/db"
	"garzon/eval"
	"garzon/grz-judge/client"
	prog "garzon/eval/programming"
)


const u_add = `grz add [options] <directory>

Options:
  --path    Colon-separated list of directories to consider 
            as roots

`
const u_update = `grz update [options] <directory>

Options:
  --path    Colon-separated list of directories to consider 
            as roots

`
const u_delete = `grz delete <ProblemID>
`
const u_submit = `grz submit [options] <ProblemID> <filename>

Options:
  --judge    URL for the judge

`

func help(args []string) {
	if len(args) == 0 {
		usage(0)
	}
	for _, cmd := range args {
		help1(cmd)
	}
}

func help1(cmd string) {
	C, ok := commands[cmd]
	if ! ok {
		_errx("unknown command '%s'\n", cmd)
	}
	fmt.Print(C.usage)
}


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
		_err("Wrong number of arguments")
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
	dot := strings.LastIndex(dir, ".")
	if dot == -1 {
		_errx(`Directory should end with ".<type>"`)
	}
	return dir[:dot], dir[dot+1:]
}

func readProblem(dir string) (id string, Problem *eval.Problem) {
	// Change to absolute path
	absdir := dir
	cwd, err := os.Getwd()
	if err != nil {
		_errx("Cannot get current working directory")
	}
	if absdir[0] != '/' {
		absdir = cwd
		if dir != "." {
			absdir += "/" + dir
		}
	}

	// Check that it is a directory
	info, err := os.Stat(absdir) 
	if err != nil {
		_errx("Cannot stat '%s'", absdir)
	}
	if ! info.IsDir() {
		_errx("'%s' is not a directory", absdir)
	}

	// Find the root
	if addPath == "" {
		_errx("No roots specified")
	}
	var root, relative string
	for _, path := range splitPath(addPath) {
		if len(path) == 0 || path[0] != '/' {
			_errx("path '%s' is not absolute", path)
		}
		if strings.HasPrefix(absdir, path) {
			root, relative = path, absdir[len(path)+1:]
			break
		}
	}
	if root == "" {
		if dir[0] != '/' {
			root, relative = cwd, dir
		} else {
			_errx("Root directory not found")
		}
	}

	// Get the <type> of the problem + ID
	base, typ := splitType(relative)
	id = strings.Replace(base, "/", ".", -1)
	
	// Lookup <type>.Evaluator
	ev := db.ObjFromType(typ + ".Evaluator")
	if ev == nil {
		_errx(`Type '%s.Evaluator' not found`, typ)
	}

	// Read Title
	title, err := ioutil.ReadFile(absdir + "/title")
	if err != nil {
		_errx("Cannot read title")
	}

	// TODO: Read statement

	Problem = &eval.Problem{
		Title: string(title), 
		StatementID: "",
	}

	// Read directory
	R, ok := ev.(eval.DirReader)
	if ! ok {
		fmt.Printf("%v\n", ev)
		_errx("Retrieved object is not a DirReader")
	}
	if err := R.ReadDir(absdir, Problem); err != nil {
		_errx("Coudln't read problem '%s': %s\n", id, err)
	}
	E, ok := ev.(eval.Evaluator)
	if ! ok {
		_errx("Retrieved object is not an Evaluator")
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
		_errx("Cannot get db 'problems': %s\n", err)
	}
	rev, _ := problems.Rev(id)
	if rev != "" {
		_errx("Problem '%s' already in the database", id)
	}
	if err := problems.Put(id, Problem); err != nil {
		_errx("Couldn't add: %s\n", err)
	}
}

func update(args []string) {
	dir := addParseFlags(args)
	
	id, Problem := readProblem(dir)
	
	// Store in the database
	problems, err := db.GetOrCreateDB("problems")
	if err != nil {
		_errx("Cannot get database 'problems': %s\n", err)
	}
	rev, _ := problems.Rev(id)
	if rev == "" {
		_errx("Problem '%s' not found in the database", id)
	}
	if err := problems.Update(id, rev, Problem); err != nil {
		_errx("Couldn't update: %s\n", err)
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
		_errx("Wrong number of arguments")
	}

	resp, err := client.Submit(args[0], args[1])
	if err != nil {
		_errx("Submission error: %s\n", err)
	}
	if strings.HasPrefix(resp, "ERROR") {
		_errx("%s\n", resp)
	}
	id := resp

	for {
		status, err := client.Status(id)
		if err != nil {
			_errx("Cannot get status: %s\n", err)
		}
		if status == "Resolved" {
			break
		}
		fmt.Printf("\r                                         \r")
		fmt.Printf("%s...", status)
		time.Sleep(500 * time.Millisecond)
	}

	submissions, err := db.GetDB("submissions")
	if err != nil {
		_errx("Cannot get 'submissions': %s\n", err)
	}

	var sub eval.Submission
	_, err = submissions.Get(id, &sub)
	if err != nil {
		_errx("Cannot get submission '%s': %s\n", id, err)
	}

	fmt.Printf("\r                                         \r")
	fmt.Printf("%s\n\n", sub.Veredict.Message)
	if sub.Veredict.Message != "Accepted" {
		fmt.Printf("%s\n", sub.Veredict.Details.Obj)
	}
}

func delette(args []string) {
	if len(args) != 1 {
		_err("Wrong number of arguments")
		usageCmd("delete", 2)
	}

	id := args[0]
	problems, err := db.GetDB("problems")
	if err != nil {
		_errx("Cannot get db 'problems': %s\n", err)
	}
	var P eval.Problem
	rev, err := problems.Get(id, &P)
	if err != nil {
		_errx("Couldn't get problem '%s': %s\n", id, err)
	}

	// Store in 'problems-deleted'
	delproblems, err := db.GetOrCreateDB("problems-deleted")
	if err != nil {
		_errx("Cannot get db 'problems-deleted'")
	}
	salt := db.RandString(8)
	err = delproblems.Put(id + "-" + salt, &P)
	if err != nil {
		_errx("Cannot backup deleted problem '%s': %s\n", id, err)
	}
	
	// Delete
	err = problems.Delete(id, rev)
	if err != nil {
		_errx("Couldn't delete problem '%s': %s\n", id, err)
	}

	fmt.Printf("Problem '%s' deleted\n", id)
}
