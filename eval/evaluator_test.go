
package eval

import (
	"os"
	"log"
	"testing"
)	

var E *Evaluator
var ID string

const P = `
#include <iostream>
using namespace std;

int main() { 
   int a, b;
   cin >> a >> b;
   cout << (a + b) << endl; 
}

`

func init() {
	// Prepare directory
	dir := os.TempDir() + "/grz-eval"
	if err := os.RemoveAll(dir); err != nil {
		log.Fatal("init: Cannot remove previous dir '%s'", dir)
		return
	}
	if err := os.Mkdir(dir, 0700); err != nil {
		log.Fatal("init: Cannot create dir '%s'", dir)
		return
	}

	E = NewEvaluator(dir)
}

func TestCompile(t *testing.T) {
	err := E.Compile(Program{lang:"c++", code: P}, &ID)
	if err != nil {
		t.Errorf("Cannot compile C++: %s", err)
		return
	}
}

func TestExecute(t *testing.T) {
	var output string
	err := E.Execute(Request{ ID: ID, input: "2 3\n" }, &output)
	if err != nil {
		t.Errorf("Cannot execute C++ program: %s", err)
	}
	if output != "5\n" {
		t.Errorf("Wrong output: '%s'", output)
	}
}

func TestDelete(t *testing.T) {
	var ok bool
	if err := E.Delete(ID, &ok); err != nil {
		t.Errorf("Cannot delete C++ program: %s", err)
	}
	if ! ok {
		t.Errorf("Deletion no ok")
	}
}
