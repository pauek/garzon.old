package main

import (
	"flag"
	"garzon/db"
	"garzon/eval"
	"path/filepath"
)

const u_update = `grz update [options] <directory>

Options:
  --path    Problem root directory
`

func update(args []string) {
	var path string
	fset := flag.NewFlagSet("update", flag.ExitOnError)
	fset.StringVar(&path, "path", "", "Problem path (colon separated)")
	fset.Parse(args)

	dir := filepath.Clean(checkOneArg("update", fset.Args()))

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
