
package programming

import (
	"os"
	"fmt"
	"log"
	"strings"
	"regexp"
	"testing"

	"github.com/pauek/garzon/db"
	"github.com/pauek/garzon/eval"
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
	BaseDir = dir
}

var keepDir bool

func evalWithInputs(model, accused string, I []string) eval.Veredict {
	ev := &Evaluator{
		Tests: make([]db.Obj, len(I)),
	}
	for i, input := range I {
		ev.Tests[i] = db.Obj{&InputTester{Input: input}}
	}
	prob := &eval.Problem{
	   Title: "Doesn't matter...",
      Solution: model, // FIXME: Code{Lang: "c++", Text: model},
		Evaluator: db.Obj{ev},
	}
	return ev.Evaluate(prob, accused, nil)
}

const Minimal = `.cc
int main() {}`
var OneEmptyInput = []string{""}

func results(V eval.Veredict) []TestResult {
	return V.Details.Obj.(VeredictDetails).Results
}

func firstRes(V eval.Veredict) string {
	return results(V)[0].Veredict
}

func TestMinimal(t *testing.T) {
	V := evalWithInputs(Minimal, Minimal, OneEmptyInput)
	fmt.Println(V)
	if results(V)[0].Veredict != "Accepted" {
		t.Fail()
	}
}

const PrintX = `.cc
#include <iostream>
int main() { std::cout << "%s" << std::endl; }`

func TestDifferentOutput(t *testing.T) {
	model   := fmt.Sprintf(PrintX, "A");
	accused := fmt.Sprintf(PrintX, "B");
	V := evalWithInputs(model, accused, OneEmptyInput)
	if firstRes(V) != "Wrong Answer" {
		t.Fail()
	}
}

const Wrong = `.cc
int main{}`

func TestModelDoesntCompile(t *testing.T) {
	V := evalWithInputs(Wrong, Minimal, OneEmptyInput)
	if V.Message != "Model doesn't compile!" {
		t.Errorf(`Error is not "Model doesn't compile" (is '%s')"`, V.Message)
	}
}

func TestAccusedDoesntCompile(t *testing.T) {
	V := evalWithInputs(Minimal, Wrong, OneEmptyInput)
	if V.Message != "Compilation Error" {
		t.Errorf(`Error should be "Compilation Error" (not '%s')`, V.Message)
	}
}

const SumAB = `.cc
#include <iostream>

int main() {
   int a, b;
   std::cin >> a >> b;
   std::cout << a + b << std::endl;
}
`

func TestSumAB(t *testing.T) {
	inputs := []string{"2 3\n", "4\n5", "1000 2000\n", "500000000 500000000\n"}
	V := evalWithInputs(SumAB, SumAB, inputs)
	for i, r := range results(V) {
		if r.Veredict != "Accepted" {
			inp := strings.Replace(inputs[i], "\n", `\n`, -1)
			t.Errorf("Failed test '%s'\n", inp)
		}
	}
}

const Echo = `.cc
#include <iostream>
int main() { int a; std::cin >> a; std::cout << a; }`
const EchoX = `.cc
#include <iostream>
int main() { int a; std::cin >> a; std::cout << (a == 3 ? -1 : a); }`

func TestEcho(t *testing.T) {
	inputs := []string{"0", "1", "2", "3", "4", "5"}
	V := evalWithInputs(Echo, EchoX, inputs)
	for i, r := range results(V) {
		if i == 3 {
			if r.Veredict != "Wrong Answer" {
				t.Errorf("Veredict should be \"Wrong Answer\" (test %d)", i)
			} 
		} else {
			if r.Veredict != "Accepted" {
				t.Errorf("Veredict should be \"Accept\" (test %d)", i)
			}
		}
	}
}

func testExecutionError(t *testing.T, model, accused string, expected string) {
	V := evalWithInputs(model, accused, OneEmptyInput)
	R0 := results(V)[0]
	if R0.Veredict != "Execution Error" {
		t.Errorf("Should be 'Execution Error'")
	}
	reason, ok := R0.Reason.Obj.(*SimpleReason)
	if !ok {
		t.Errorf("Reason is no a SimpleReason")
	}
	found, err := regexp.MatchString(expected, reason.Message)
	if err != nil {
		t.Errorf(`Error matching regexp "%s" against "%s"`, expected, R0)
	}
	if ! found {
		t.Errorf(`Wrong veredict "%s", should be "%s"`, R0, expected)
	}

}

func TestTimeLimitExceeded(t *testing.T) {
	infLoop := ".cc\nint main() { while (1); }"
	testExecutionError(t, Minimal, infLoop, "Time Limit Exceeded")
}

func TestSegmentationFault(t *testing.T) {
	segFault := ".cc\nint main() { int T[1]; T[1000000] = 1; }"
	testExecutionError(t, Minimal, segFault, "Segmentation Fault")
}

func TestMemoryLimit1(t *testing.T) {
	normal := ".cc\n#include <vector>\nint main() { std::vector<int> v(1000000); }"
	abort  := ".cc\n#include <vector>\nint main() { std::vector<int> v(100000000); }"
	testExecutionError(t, normal, abort, "Memory Limit Exceeded")
}

func TestMemoryLimit2(t *testing.T) {
	normal := `.cc
#include <stdlib.h>

int main() {
   int i;
   void *data[64];
   for (i = 0; i < 64; i++) {
      data[i] = malloc(1024 * 1024); // 1 MB
	}
}
`
	wrong := `.cc
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
	opener := `.cc
#include <fstream>
int main() { std::ofstream F("file"); F << '\n'; }`
	testExecutionError(t, Minimal, opener, `Forbidden Syscall 'open\("file"\)'`)
}

func TestForbiddenSyscall(t *testing.T) {
	execer := `.cc
#include <unistd.h>
int main() { 
   char *argv[] = { NULL }, *envp[] = { NULL };
   execve("/bin/ls", argv, envp); 
}`
	testExecutionError(t, Minimal, execer, 
		`Forbidden Syscall '_execve\([0-9a-f]*,[0-9a-f]*,[0-9a-f]*\)'`)
}

// TODO: Forbidden Syscall (fork) ?
// TODO: Forbidden Syscall (unlink) ?
// TODO: Forbidden Syscall (kill) ?

const sumABFiles = `.cc
#include <fstream>
using namespace std;

int main() {
  ifstream A("A"), B("B");
  ofstream C("C");
  int a, b;
  A >> a; B >> b;
  C << a + b;
}
`

const wrongFiles1 = `.cc
#include <fstream>
using namespace std;

int main() {
  ifstream A("A"), B("B");
  ofstream C("D");
  int a, b;
  A >> a; B >> b;
  C << a + b;
}
`

const wrongAnswer1 = `.cc
#include <fstream>
using namespace std;

int main() {
  ifstream A("A"), B("B");
  ofstream C("C");
  int a, b;
  A >> a; B >> b;
  C << a * b;
}
`

var filesEv *Evaluator = &Evaluator{
   Tests: []db.Obj{
		{&FilesTester{
			InputFiles: []FileInfo{
					FileInfo{RelPath: "A", Contents: "13"},
					FileInfo{RelPath: "B", Contents: "17"},
				},
			OutputFiles: []FileInfo{
					FileInfo{RelPath: "C"},
				},
		}},
	},
}

var filesProb *eval.Problem = &eval.Problem{
   Title: "Simple One with Files",
   Solution: sumABFiles, // FIXME: Code{Lang: "c++", Text: sumABFiles},
   Evaluator: db.Obj{filesEv},
}

func TestFileTester(t *testing.T) {
	var V eval.Veredict

	// Good
	V = filesEv.Evaluate(filesProb, sumABFiles, nil)
	if firstRes(V) != "Accepted" {
		t.Errorf("Test should be accepted")
	}

	// Creates a file named 'D' instead of 'C'
	V = filesEv.Evaluate(filesProb, wrongFiles1, nil)
	R0 := results(V)[0]
	if R0.Veredict != "Execution Error" {
		t.Errorf("Should be 'Execution Error'")
	}
	reason, ok := R0.Reason.Obj.(*SimpleReason)
	if !ok {
		t.Errorf("Reason is no a SimpleReason")
	}
	if reason.Message != "Forbidden Syscall 'open(\"D\")'" {
		t.Errorf("Wrong Veredict")
	}

	// Doesn't compute sum
	V = filesEv.Evaluate(filesProb, wrongAnswer1, nil)
	if res := firstRes(V); res != "Wrong Answer" {
		t.Errorf("Wrong Veredict ('%s')", res)
	}
}
