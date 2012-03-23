
package db

import (
	// "fmt"
	"testing"
)

type Test1 struct {
	A string
	B int
	C bool
}

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

func TestPutMyData(t *testing.T) {
	d := MyData{First: "Groucho", Last: "Marx", Age: 55}
	_, rev, err := db.Get("groucho")
	if err == nil {
		_ = db.Delete("groucho", rev)
	}
	err = db.Put("groucho", d)
	if err != nil {
		t.Errorf("Cannot put: %s\n", err)
	}
}

func TestGetMyData(t *testing.T) {
	_d, _, err := db.Get("groucho")
	if err != nil {
		t.Errorf("Cannot get: %s\n", err)
	}
	d := _d.(*MyData)
	if d.First != "Groucho" ||
		d.Last  != "Marx" ||
		d.Age   != 55 {
		t.Errorf("Wrong data\n")
	}
}