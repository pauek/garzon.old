package main

import (
	"garzon/db"
)

const u_add = `grz add [options] <directory>

Options:
  --path    Problem root directory

`

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
