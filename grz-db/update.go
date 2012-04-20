package main

import (
	"fmt"
	"github.com/pauek/garzon/db"
	"github.com/pauek/garzon/eval"
)

const u_update = `grz-db update [-R] <directory>

Options:
   -R,   Update recursively (all problems found under <directory>)
`

func update(args []string) {
	storeFunc = func(db *db.Database, id string, Problem *eval.Problem) error {
		rev, _ := db.Rev(id)
		if rev == "" {
			return fmt.Errorf("Problem '%s' not in the database", id)
		}
		if err := db.Update(id, rev, Problem); err != nil {
			return err
		}
		return nil
	}
	addupdate("add", args)
}
