
package program

import (
	"os"
	"fmt"
	"log"
	"strings"
	"testing"
)	

var ID string

func init() {
	dir := os.TempDir() + "/grz-eval"
	if err := os.RemoveAll(dir); err != nil {
		log.Fatal("init: Cannot remove previous dir '%s'", dir)
	}
	if err := os.Mkdir(dir, 0700); err != nil {
		log.Fatal("init: Cannot create dir '%s'", dir)
	}
	Evaluator.BaseDir = dir
}

func mkEvaluation(model, accused string) Evaluation {
	ev := new(Evaluation)
	ev.Model   = Program{Lang: "c++", Code: model}
	ev.Accused = Program{Lang: "c++", Code: accused}
	return *ev
}

var keepDir bool

func evalWithInputs(model, accused string, I []string) (R []Result, err error) {
	var id string
	var ok bool

	ev := mkEvaluation(model, accused)
	if err = Evaluator.StartEvaluation(ev, &id); err != nil {
		return nil, err
	}

	R = make([]Result, len(I))
	for i, input := range I {
		T := TestInfo{ id, &InputTester{ Input: input } }
		if err = Evaluator.RunTest(T, &R[i]); err != nil {
			R[i].Veredict = fmt.Sprintf("%s", err)
		}
	}

	if ! keepDir {
		if err = Evaluator.EndEvaluation(id, &ok); err != nil {
			return nil, err
		}
	}

	return R, nil
}

const Minimal = `int main() {}`
var OneEmptyInput = []string{""}

func TestMinimal(t *testing.T) {
	R, _ := evalWithInputs(Minimal, Minimal, OneEmptyInput)
	if R[0].Veredict != "Accept" {
		t.Fail()
	}
}

const PrintX = `#include <iostream>
int main() { std::cout << "%s" << std::endl; }`

func TestDifferentOutput(t *testing.T) {
	model   := fmt.Sprintf(PrintX, "A");
	accused := fmt.Sprintf(PrintX, "B");
	R, _ := evalWithInputs(model, accused, OneEmptyInput)
	if R[0].Veredict != "Wrong Answer" {
		t.Fail()
	}
}

const Wrong = `int main{}`

func TestModelDoesntCompile(t *testing.T) {
	_, err := evalWithInputs(Wrong, Minimal, OneEmptyInput)
	if err == nil {
		t.Errorf("Compilation should fail")
	}
	errmsg := fmt.Sprintf("%s", err)
	if ! strings.HasPrefix(errmsg, "Error compiling 'model':") {
		t.Errorf("Error is not \"Error compiling 'model'\"")
	}
}

func TestAccusedDoesntCompile(t *testing.T) {
	_, err := evalWithInputs(Minimal, Wrong, OneEmptyInput)
	if err == nil {
		t.Errorf("Compilation should fail")
	}
	errmsg := fmt.Sprintf("%s", err)
	if ! strings.HasPrefix(errmsg, "Error compiling 'accused':") {
		t.Errorf("Error is not \"Error compiling 'accused'\"")
	}
}

const SumAB = `#include <iostream>

int main() {
   int a, b;
   std::cin >> a >> b;
   std::cout << a + b << std::endl;
}
`

func TestSumAB(t *testing.T) {
	inputs := []string{"2 3\n", "4\n5", "1000 2000\n", "500000000 500000000\n"}
	Res, err := evalWithInputs(SumAB, SumAB, inputs)
	if err != nil {
		t.Errorf("Test failed: %s\n", err)
	}
	for i, r := range Res {
		if r.Veredict != "Accept" {
			inp := strings.Replace(inputs[i], "\n", `\n`, -1)
			t.Errorf("Failed test '%s'\n", inp)
		}
	}
}

const Echo = `#include <iostream>
int main() { int a; std::cin >> a; std::cout << a; }`
const EchoX = `#include <iostream>
int main() { int a; std::cin >> a; std::cout << (a == 3 ? -1 : a); }`

func TestEcho(t *testing.T) {
	inputs := []string{"0", "1", "2", "3", "4", "5"}
	Res, err := evalWithInputs(Echo, EchoX, inputs)
	if err != nil {
		t.Errorf("Test failed: %s\n", err)
	}
	for i, r := range Res {
		if i == 3 {
			if r.Veredict != "Wrong Answer" {
				t.Errorf("Veredict should be \"Wrong Answer\" (test %d)", i)
			} 
		} else {
			if r.Veredict != "Accept" {
				t.Errorf("Veredict should be \"Accept\" (test %d)", i)
			}
		}
	}
}

func testExecutionError(t *testing.T, model, accused string, expected string) {
	R, err := evalWithInputs(model, accused, OneEmptyInput)
	if err != nil {
		t.Errorf("Evaluation should not return error")
	}
	if R[0].Veredict != expected {
		t.Errorf("Veredict should be \"%s\" (is \"%s\")", expected, R[0].Veredict)
	}
}

func TestTimeLimitExceeded(t *testing.T) {
	infLoop := `int main() { while (1); }`
	testExecutionError(t, Minimal, infLoop, "Time Limit Exceeded")
}

func TestSegmentationFault(t *testing.T) {
	segFault := `int main() { int T[1]; T[100000000] = 1; }`
	testExecutionError(t, Minimal, segFault, "Segmentation Fault")
}

func TestMemoryLimit1(t *testing.T) {
	normal := "#include <vector>\nint main() { std::vector<int> v(1000000); }"
	abort  := "#include <vector>\nint main() { std::vector<int> v(100000000); }"
	testExecutionError(t, normal, abort, "Memory Limit Exceeded")
}

func TestMemoryLimit2(t *testing.T) {
	normal := `
#include <stdlib.h>

int main() {
   int i;
   void *data[64];
   for (i = 0; i < 64; i++) {
      data[i] = malloc(1024 * 1024); // 1 MB
	}
}
`
	wrong := `
#include <stdlib.h>

int main() {
   int i;
   void *data[8192]; // ~8Gb
   for (i = 0; i < 8192; i++) {
      data[i] = malloc(1024 * 1024); // 1 MB
	}
}
`
	testExecutionError(t, normal, wrong, "Memory Limit Exceeded")   
}

// TODO: Aborted
// TODO: Interrupted
// TODO: Out of Memory
// TODO: Forbidden Syscall (open)
// TODO: Forbidden Syscall (execve)
// TODO: Forbidden Syscall (fork)
// TODO: Forbidden Syscall (unlink)
// TODO: Forbidden Syscall (kill)


