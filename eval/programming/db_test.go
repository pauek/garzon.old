
package programming

import (
	"fmt"
	"reflect"
	"testing"
	"github.com/pauek/garzon/eval"
	"github.com/pauek/garzon/db"
)

// var filesprob *Problem [declared in programming_test.go]

func TestStoreProblem(t *testing.T) {
	const dbname = "this-database-shouldn-exist-at-all-in-the-face-of-the-earth-42"

	D, err := db.GetOrCreateDB(dbname)
	if err != nil {
		t.Errorf("Cannot get or create database: %s\n", err)
		return
	}

	const pid = "Cpp.Intro.SumaEnteros"

	// Put
	rev, err := D.Rev(pid)
	if rev != "" {
		t.Fatalf("Un-fuckin'-believable: this problem already existed!")
	}
	err = D.Put(pid, filesProb)
	if err != nil {
		t.Errorf("Cannot put: %s\n", err)
	}

	// Get
	var obj eval.Problem
	rev, err = D.Get(pid, &obj)
	if err != nil {
		t.Errorf("Cannot get: %s\n", err)
	}
	if ! reflect.DeepEqual(filesProb, &obj) {
		fmt.Printf("%#v\n", filesProb)
		fmt.Printf("%#v\n", obj)
		t.Errorf("Different data\n")
	}
	
	// Delete
	if err := D.Delete(pid, rev); err != nil {
		t.Errorf("Cannot delete '%s': %s\n", pid, err)
	}

	db.DeleteDB(D)
}