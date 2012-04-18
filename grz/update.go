package main

import (
	"garzon/db"
)

const u_update = `grz update [options] <directory>

Options:
  --path    Problem root directory

`

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
