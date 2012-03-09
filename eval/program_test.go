
package eval

import (
	"os"
	"log"
	"testing"
)	

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

	ProgramEvaluator.BaseDir = dir
}

func TestRun(t *testing.T) {
	ev := new(ProgramEvaluation)
	ev.Model   = Program{Lang: "c++", Code: P}
	ev.Accused = Program{Lang: "c++", Code: P}
	ev.Tests = []Test{}
	for _, inp := range []string{"1 1\n", "2 2\n", "1000 -4\n"} {
		ev.Tests = append(ev.Tests, &InputTest{inp})
	}
	var R Results
	err := ProgramEvaluator.Run(*ev, &R)
	if err != nil {
		t.Error(err)
	}
}