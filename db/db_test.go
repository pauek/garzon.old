
package db

import (
	// "fmt"
	"testing"
)

type MyData struct {
	First, Last string
	Age float64
}

func init() {
	Register(MyData{})
}

var db Database = Database{
   host: "localhost", 
   port: "5984", 
   db: "test",
}

func TestMyData(t *testing.T) {
	// Put
	d := &MyData{First: "Groucho", Last: "Marx", Age: 55}
	rev, err := db.Rev("groucho")
	if rev != "" {
		if err := db.Delete("groucho", rev); err != nil {
			t.Errorf("Can't delete rev '%s' of 'groucho': %s\n", rev, err)
			return
		}
	}
	err = db.Put("groucho", d)
	if err != nil {
		t.Errorf("Cannot put: %s\n", err)
	}
	// Get
	obj, rev, err := db.Get("groucho")
	if err != nil {
		t.Errorf("Cannot get: %s\n", err)
	}
	d, ok := obj.(*MyData)
	if ! ok {
		t.Errorf("Returned object is not of type 'MyData'")
		return
	}
	if d.First != "Groucho" ||
		d.Last  != "Marx" ||
		d.Age   != 55 {
		t.Errorf("Wrong data\n")
	}
	// Delete
	if err := db.Delete("groucho", rev); err != nil {
		t.Errorf("Cannot delete 'groucho': %s\n", err)
	}
}

