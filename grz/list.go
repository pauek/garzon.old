package main

import (
	"fmt"
)

const u_list = `grz list`

func list(args []string) {
	ids, err := client.ProblemList(); 
	if err != nil {
		_errx("Cannot get problem list: %s", err)
	}
	for _, id := range ids {
		fmt.Printf("%s\n", id)
	}
}
