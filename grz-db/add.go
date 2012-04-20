package main

import (
	"fmt"
	"github.com/pauek/garzon/db"
	"github.com/pauek/garzon/eval"
)

const u_add = `grz-db add [-R] <directory>

Options:
   -R,   Add recursively (all problems found under <directory>)
`

func add(args []string) {
	storeFunc = func(db *db.Database, id string, Problem *eval.Problem) error {
		rev, _ := db.Rev(id)
		if rev != "" {
			return fmt.Errorf("Problem '%s' already in the database", id)
		}
		if err := db.Put(id, Problem); err != nil {
			return err
		}
		return nil
	}
	addupdate("add", args)
}
