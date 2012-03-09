
package eval

import (
	"os"
	"log"
	"testing"
)	

var ID string

var Programs = []string{`
#include <iostream>
using namespace std;

int main() { 
   int a, b;
   cin >> a >> b;
   cout << (a + b) << endl; 
}

`,
`int main() {}`,
}

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

func test(p1, p2 string, inputs []string) *ProgramEvaluation {
	ev := new(ProgramEvaluation)
	ev.Model   = Program{Lang: "c++", Code: p1}
	ev.Accused = Program{Lang: "c++", Code: p2}
	ev.Tests = []Test{}
	for _, inp := range inputs {
		ev.Tests = append(ev.Tests, &InputTest{inp})
	}
	return ev
}

func TestSame(t *testing.T) {
	ev := test(Programs[0], Programs[0], []string{"1 1\n", "1 -2\n", "2 3\n"})
	var R Results
	err := ProgramEvaluator.Run(*ev, &R)
	if err != nil { t.Error(err) }
}

func TestVoid(t *testing.T) {
	ev := test(Programs[1], Programs[0], []string{"1 1\n", "1 -2\n", "2 3\n"})
	var R Results
	err := ProgramEvaluator.Run(*ev, &R)
	if err != nil { t.Error(err) }
}
