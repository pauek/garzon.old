package main

import (
	"fmt"
	"github.com/pauek/garzon/db"
)

const u_list = `grz-db list`

func list(args []string) {
	problems, err := db.GetDB("problems")
	if err != nil {
		_errx("Cannot get db 'problems': %s\n", err)
	}

	ids, err := problems.AllIDs()
	if err != nil {
		_errx("Cannot get all IDs from 'problems'")
	}
	for _, id := range ids {
		fmt.Printf("%s\n", id)
	}
}
