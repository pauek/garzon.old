package main

import (
	"flag"
	"garzon/db"
	"garzon/eval"
	"path/filepath"
)

const u_add = `grz add [options] <directory>

Options:
  --path    Problem root directory

`

func add(args []string) {
	var path string
	fset := flag.NewFlagSet("add", flag.ExitOnError)
	fset.StringVar(&path, "path", "", "Problem path (colon separated)")
	fset.Parse(args)

	dir := filepath.Clean(checkOneArg("add", fset.Args()))

	if path != "" {
		eval.GrzPath = path
	}
	id, Problem, err := eval.ReadFromDir(dir)
	if err != nil {
		_errx("Cannot read problem at '%s': %s\n", dir, err)
	}

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
