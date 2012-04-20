package main

import (
	"flag"
	"fmt"
	"github.com/pauek/garzon/db"
	"github.com/pauek/garzon/eval"
	"os"
	"path/filepath"
)

const u_add = `grz-db add [-R] <directory>

Options:
   -R,   Add recursively (all problems found under <directory>)
`

var storeFunc func(db *db.Database, id string, Problem *eval.Problem) error

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

func addupdate(who string, args []string) {
	var recursive bool
	fset := flag.NewFlagSet(who, flag.ExitOnError)
	fset.BoolVar(&recursive, "R", false, "")
	fset.Parse(args)

	dir := filepath.Clean(checkOneArg(who, fset.Args()))

	if recursive {
		_addupdaterecursive(dir)
	} else {
		err := _addupdate(dir)
		if err != nil {
			_errx(fmt.Sprintf("%s\n", err))
		}
	}
}

func _addupdaterecursive(dir string) {
	if eval.GrzPath == "" {
		fmt.Printf("No roots.\n")
		return
	}
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if ok, _, _ := eval.IsProblem(path); ok {
			if _, rel, err := eval.SplitRootRelative(path, path); err == nil {
				e := _addupdate(path)
				if e != nil {
					fmt.Printf("Error: %s\n", e)
				} else {
					fmt.Printf("%s\n", eval.IdFromDir(rel))
				}
			} else {
				fmt.Printf("Error: %s\n", err)
			}
		}
		return nil
	})
}

func _addupdate(dir string) error {
	id, Problem, err := eval.ReadFromDir(dir)
	if err != nil {
		return fmt.Errorf("Cannot read problem at '%s': %s\n", dir, err)
	}

	// Store in the database
	problems, err := db.GetOrCreateDB("problems")
	if err != nil {
		_errx("Cannot get db 'problems': %s\n", err)
	}
	if err := storeFunc(problems, id, Problem); err != nil {
		return err
	}
	return nil
}
