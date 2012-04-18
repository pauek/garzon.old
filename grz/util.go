package main

import (
	"fmt"
	"garzon/db"
	"garzon/eval"
	"io/ioutil"
	"os"
	"strings"
)

func _err(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "grz: "+format+"\n", args...)
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

var GrzPath string

func setGrzPath(path string) {
	GrzPath = path
	if GrzPath == "" {
		// TODO: Check that no path in 'GrzPath' is prefix of the others!
		GrzPath = os.Getenv("GRZ_PATH")
	}
}

func splitPath(pathstr string) (path []string) {
	if pathstr == "" {
		return
	}
	for _, p := range strings.Split(pathstr, ":") {
		if p[len(p)-1] == '/' {
			p = p[:len(p)-1]
		}
		path = append(path, p)
	}
	return path
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
	if !info.IsDir() {
		_errx("'%s' is not a directory", absdir)
	}

	// Find the root
	if GrzPath == "" {
		_errx("No roots specified")
	}
	var root, relative string
	for _, path := range splitPath(GrzPath) {
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
		Title:       string(title),
		StatementID: "",
	}

	// Read directory
	R, ok := ev.(eval.DirReader)
	if !ok {
		fmt.Printf("%v\n", ev)
		_errx("Retrieved object is not a DirReader")
	}
	if err := R.ReadDir(absdir, Problem); err != nil {
		_errx("Coudln't read problem '%s': %s\n", id, err)
	}
	E, ok := ev.(eval.Evaluator)
	if !ok {
		_errx("Retrieved object is not an Evaluator")
	}
	Problem.Evaluator = db.Obj{E}
	return
}
