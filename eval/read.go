package eval

import (
	"fmt"
	"github.com/pauek/garzon/db"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var GrzPath string

func init() {
	GrzPath = os.Getenv("GRZ_PATH")
}

func grzPath() []string {
	return filepath.SplitList(GrzPath)
}

func readFrom(abspath string) (P *Problem, err error) {
	abspath = filepath.Clean(abspath)
	base := filepath.Base(abspath)
	idx := strings.Index(base, ".")
	if idx == -1 {
		return nil, fmt.Errorf("'%s' doesn't match <Name>.<Type>", base)
	}
	typ := base[idx+1:]

	// Lookup <type>.Evaluator
	ev := db.ObjFromType(typ + ".Evaluator")
	if ev == nil {
		return nil, fmt.Errorf(`Type '%s.Evaluator' not found`, typ)
	}

	// Read Title
	title, err := ioutil.ReadFile(abspath + "/title")
	if err != nil {
		return nil, fmt.Errorf("Cannot read title")
	}

	// TODO: Read statement

	P = &Problem{
		Title:       string(title),
		StatementID: "",
	}

	// Read directory
	R, ok := ev.(DirReader)
	if !ok {
		fmt.Printf("%v\n", ev)
		return nil, fmt.Errorf("Retrieved object is not a DirReader")
	}
	if err := R.ReadDir(abspath, P); err != nil {
		return nil, fmt.Errorf("Couldn't read problem at '%s': %s\n", abspath, err)
	}
	E, ok := ev.(Evaluator)
	if !ok {
		return nil, fmt.Errorf("Retrieved object is not an Evaluator")
	}
	P.Evaluator = db.Obj{E}
	return P, nil
}

func ReadFromID(id string) (Problem *Problem, err error) {
	reldir := strings.Replace(id, ".", "/", -1)

	var dirs []string
	for _, root := range grzPath() {
		glob := filepath.Join(root, reldir) + ".*"
		dirs, err = filepath.Glob(glob)
		if err != nil {
			return nil, fmt.Errorf("Cannot glob '%s.*'\n", glob)
		}
		if len(dirs) > 0 {
			break
		}
	}
	if len(dirs) == 0 {
		return nil, fmt.Errorf("Problem with id '%s' not found in GRZ_PATH", id)
	}
	for i, d := range dirs {
		if i == 0 {
			Problem, err = readFrom(d)
		} else {
			fmt.Fprintf(os.Stderr, "warning: ignoring problem '%s'\n", d)
		}
	}
	return
}

func ReadFromDir(dir string) (id string, problem *Problem, err error) {
	abspath, err := filepath.Abs(dir)
	if err != nil {
		return "", nil, fmt.Errorf("Cannot get abs of '%s': %s\n", dir)
	}

	// Check that it is a directory
	info, err := os.Stat(abspath)
	if err != nil {
		return "", nil, fmt.Errorf("Cannot stat '%s'", abspath)
	}
	if !info.IsDir() {
		return "", nil, fmt.Errorf("'%s' is not a directory", abspath)
	}

	// Find root + relative
	_, relative, err := findRoot(dir)
	if err != nil {
		return "", nil, err
	}

	// Get the ID of the problem
	preid := strings.Split(relative, ".")[0]
	id = strings.Replace(preid, "/", ".", -1)
	problem, err = readFrom(abspath)
	return id, problem, err
}

func findRoot(dir string) (root, relative string, err error) {
	abspath, err := filepath.Abs(dir)
	if err != nil {
		return "", "", fmt.Errorf("Cannot get abs of '%s': %s\n", dir)
	}
	grzpath := grzPath()
	if len(grzpath) == 0 {
		return "", "", fmt.Errorf("No roots (GRZ_PATH empty)")
	}
	for _, path := range grzpath {
		if len(path) == 0 || !filepath.IsAbs(path) {
			return "", "", fmt.Errorf("path '%s' is not absolute", path)
		}
		if strings.HasPrefix(abspath, path) {
			if relative, err = filepath.Rel(path, abspath); err != nil {
				return "", "", err
			}
			root = path
			break
		}
	}
	if root == "" {
		if !filepath.IsAbs(dir) {
			pwd, err := os.Getwd()
			if err != nil {
				return "", "", fmt.Errorf("Cannot get working dir: %s\n", err)
			}
			root, relative = pwd, dir
		} else {
			return "", "", fmt.Errorf("Root directory '/' not allowed")
		}
	}
	return
}
