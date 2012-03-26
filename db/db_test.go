
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
	Register(Problem{})
	Register(Test1{})
	Register(Test2{})
}

var db Database = Database{
   host: "localhost", 
   port: "5984", 
   db: "test",
}

func TestMyData(t *testing.T) {
	const pid = "Cpp.Intro.SumaEnteros"

	// Put
	P := &Problem{
	   Title: "Suma de Enteros",
      Tests: []Obj {
			Obj{&Test1{A: "Input test1"}},
			Obj{&Test2{B: 45}},
		},
	}
	rev, err := db.Rev(pid)
	if rev != "" {
		if err := db.Delete(pid, rev); err != nil {
			t.Errorf("Can't delete rev '%s' of '%s': %s\n", rev, pid, err)
			return
		}
	}
	err = db.Put(pid, P)
	if err != nil {
		t.Errorf("Cannot put: %s\n", err)
	}

	// Get
	obj, rev, err := db.Get(pid)
	if err != nil {
		t.Errorf("Cannot get: %s\n", err)
	}
	if ! reflect.DeepEqual(P, obj) {
		fmt.Printf("%#v\n", P)
		fmt.Printf("%#v\n", obj)
		t.Errorf("Different data\n")
	}

	// Delete
	if err := db.Delete(pid, rev); err != nil {
		t.Errorf("Cannot delete '%s': %s\n", pid, err)
	}
}
