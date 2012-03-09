
package eval

import (
	"os"
	"log"
	"testing"
)	

var ID string

const Wrong   = `int main{}`
const Minimal = `int main() {}`
const SigSEGV = `int T[1]; int main() { T[1000000] = 0; }`

const SumAB = `
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

func doTest(p1, p2 string, I []string) (Results, error) {
	ev := test(p1, p2, I)
	var R Results
	err := ProgramEvaluator.Run(*ev, &R)
	return R, err
}

var tests = []struct { 
	p1, p2 string
	inputs []string
} {
	{ Minimal, Minimal,  []string{""} },
	{ Wrong,   Minimal,  []string{""} },
	{ Minimal, Wrong,    []string{""} },
	{ SigSEGV, Minimal,  []string{""} },
	{ Minimal, SigSEGV,  []string{""} },
	{ Minimal, SumAB,    []string{""} },
	{ SumAB,   SumAB,    []string{""} },	
}

func TestEvaluator(t *testing.T) {
	for _, test := range tests {
		_, err := doTest(test.p1, test.p2, test.inputs)
		if err != nil {
			log.Printf("%v", err)
			t.Error(err)
		}
	}
}
