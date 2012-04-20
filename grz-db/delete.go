package main

import (
	"fmt"
	"github.com/pauek/garzon/db"
	"github.com/pauek/garzon/eval"
)

const u_delete = `grz delete <ProblemID>`

func delette(args []string) {
	id := checkOneArg("delete", args)

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
	err = delproblems.Put(id+"-"+salt, &P)
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
