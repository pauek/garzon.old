
package db

import (
	"fmt"
	"testing"
	"reflect"
)

type Problem struct {
	Title string
	Tests []Obj
}

type Test1 struct {
	A string
}

type Test2 struct {
	B int
}

func init() {
	Register("Problem", Problem{})
	Register("Test1", Test1{})
	Register("Test2", Test2{})
}

func TestProblem(t *testing.T) {
	const pid = "Cpp.Intro.SumaEnteros"

	db, err := GetOrCreateDB("test-problem-0001")
	if err != nil {
		t.Fatalf("Cannot get or create database: %s\n", err)
	}

	// Put
	P := &Problem{
	   Title: "Suma de Enteros",
      Tests: []Obj {
			Obj{&Test1{A: "Input test1"}},
			Obj{&Test2{B: 45}},
		},
	}
	rev, err := db.Rev(pid + "1")
	if rev != "" {
		if err := db.Delete(pid, rev); err != nil {
			t.Errorf("Can't delete rev '%s' of '%s': %s\n", rev, pid, err)
			return
		}
	}
	err = db.Put(pid + "1", P)
	if err != nil {
		t.Errorf("Cannot put: %s\n", err)
	}
	err = db.Put(pid + "2", P)
	if err != nil {
		t.Errorf("Cannot put: %s\n", err)
	}

	// AllIDs
	ids, err := db.AllIDs()
	if err != nil {
		t.Errorf("AllIDs failed: %s\n", err)
	}
	if len(ids) != 2 {
		t.Errorf("AllIDs should have length 1")
	} else {
		if ids[0] != pid + "1" {
			t.Errorf("First ID should be '%s'", pid)
		}
		if ids[1] != pid + "2" {
			t.Errorf("First ID should be '%s'", pid)
		}
	}
	
	// Get
	var obj Problem
	rev, err = db.Get(pid + "1", &obj)
	if err != nil {
		t.Errorf("Cannot get: %s\n", err)
	}
	if ! reflect.DeepEqual(P, &obj) {
		fmt.Printf("%#v\n", P)
		fmt.Printf("%#v\n", &obj)
		t.Fatal("Different data\n")
	}

	// Delete
	if err := db.Delete(pid, rev); err != nil {
		t.Errorf("Cannot delete '%s': %s\n", pid, err)
	}

	DeleteDB(db)
}
