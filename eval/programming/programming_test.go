
package programming

import (
	"os"
	"fmt"
	"log"
	"strings"
	"regexp"
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

var keepDir bool

func evalWithInputs(model, accused string, I []string) (R *Result) {
	prob := &Problem{
	   Title: "Doesn't matter...",
      Solution: Code{Lang: "c++", Text: model},
      Tests: make([]Tester, len(I)),
	}
	sub := Submission{
	   Problem: prob, 
	   Accused: Code{Lang: "c++", Text: accused},
	}
	for i, input := range I {
		prob.Tests[i] = &InputTester{Input: input}
	}
	return Evaluator.Submit(sub)
}

const Minimal = `int main() {}`
var OneEmptyInput = []string{""}

func TestMinimal(t *testing.T) {
	R := evalWithInputs(Minimal, Minimal, OneEmptyInput)
	if R.Results[0].Veredict != "Accept" {
		t.Fail()
	}
}

const PrintX = `#include <iostream>
int main() { std::cout << "%s" << std::endl; }`

func TestDifferentOutput(t *testing.T) {
	model   := fmt.Sprintf(PrintX, "A");
	accused := fmt.Sprintf(PrintX, "B");
	R := evalWithInputs(model, accused, OneEmptyInput)
	if R.Results[0].Veredict != "Wrong Answer" {
		t.Fail()
	}
}

const Wrong = `int main{}`

func TestModelDoesntCompile(t *testing.T) {
	R := evalWithInputs(Wrong, Minimal, OneEmptyInput)
	errmsg := fmt.Sprintf("%s", R.Veredict)
	if ! strings.HasPrefix(errmsg, "Error compiling 'model':") {
		t.Errorf("Error is not \"Error compiling 'model'\" (is \"%s\")", errmsg)
	}
}

func TestAccusedDoesntCompile(t *testing.T) {
	R := evalWithInputs(Minimal, Wrong, OneEmptyInput)
	errmsg := fmt.Sprintf("%s", R.Veredict)
	if ! strings.HasPrefix(errmsg, "Error compiling 'accused':") {
		t.Errorf("Error is not \"Error compiling 'accused'\" (is \"%s\")", errmsg)
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
	Res := evalWithInputs(SumAB, SumAB, inputs)
	for i, r := range Res.Results {
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
	Res := evalWithInputs(Echo, EchoX, inputs)
	for i, r := range Res.Results {
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
	R := evalWithInputs(model, accused, OneEmptyInput)
	R0 := R.Results[0].Veredict
	found, err := regexp.MatchString(expected, R0)
	if err != nil {
		t.Errorf(`Error matching regexp "%s" against "%s"`, expected, R0)
	}
	if ! found {
		t.Errorf(`Wrong veredict "%s", should be "%s"`, R0, expected)
	}

}

func TestTimeLimitExceeded(t *testing.T) {
	infLoop := `int main() { while (1); }`
	testExecutionError(t, Minimal, infLoop, "Time Limit Exceeded")
}

func TestSegmentationFault(t *testing.T) {
	segFault := `int main() { int T[1]; T[1000000] = 1; }`
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

func TestForbiddenSyscall1(t *testing.T) {
	opener := `#include <fstream>
   int main() { std::ofstream F("file"); F << '\n'; }`
	testExecutionError(t, Minimal, opener, `Forbidden Syscall \[open\("file"\)\]`)
}

func TestForbiddenSyscall(t *testing.T) {
	execer := `#include <unistd.h>
   int main() { 
      char *argv[] = { NULL }, *envp[] = { NULL };
      execve("/bin/ls", argv, envp); 
   }`
	testExecutionError(t, Minimal, execer, 
		`Forbidden Syscall \[_execve\([0-9a-f]*,[0-9a-f]*,[0-9a-f]*\)\]`)
}

// TODO: Forbidden Syscall (fork) ?
// TODO: Forbidden Syscall (unlink) ?
// TODO: Forbidden Syscall (kill) ?


