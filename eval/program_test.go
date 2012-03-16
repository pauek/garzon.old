
package eval

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

func mkEvaluation(model, accused string) ProgramEvaluation {
	ev := new(ProgramEvaluation)
	ev.Model   = Program{Lang: "c++", Code: model}
	ev.Accused = Program{Lang: "c++", Code: accused}
	return *ev
}

func evalWithInputs(model, accused string, I []string) (R []Result, err error) {
	var id string
	var ok bool

	ev := mkEvaluation(model, accused)
	if err = Evaluator.StartEvaluation(ev, &id); err != nil {
		return nil, err
	}

	R = make([]Result, len(I))
	for i, input := range I {
		T := TestInfo{ id, &InputTester{ input, nil } }
		if err = Evaluator.RunTest(T, &R[i]); err != nil {
			Evaluator.EndEvaluation(id, &ok)
			return nil, err
		}
	}

	if err = Evaluator.EndEvaluation(id, &ok); err != nil {
		return nil, err
	}

	return R, nil
}

const Minimal = `int main() {}`

func TestMinimal(t *testing.T) {
	R, _ := evalWithInputs(Minimal, Minimal, []string{""})
	if R[0].Veredict != ACCEPT {
		t.Fail()
	}
}

const PrintX = `#include <iostream>
int main() { std::cout << "%s" << std::endl; }`

func TestDifferentOutput(t *testing.T) {
	model   := fmt.Sprintf(PrintX, "A");
	accused := fmt.Sprintf(PrintX, "B");
	R, _ := evalWithInputs(model, accused, []string{""})
	if R[0].Veredict != WRONG_ANSWER {
		t.Fail()
	}
	output, ok := R[0].Reason.(map[string]string)
	if ! ok {
		t.Fail()
	}
	if output["model"] != "A\n" {
		t.Errorf("Model output incorrect")
	}
	if output["accused"] != "B\n" {
		t.Errorf("Accused output incorrect")
	}
}

const Wrong = `int main{}`

func TestModelDoesntCompile(t *testing.T) {
	_, err := evalWithInputs(Wrong, Minimal, []string{""})
	if err == nil {
		t.Errorf("Compilation should fail")
	}
	errmsg := fmt.Sprintf("%s", err)
	if ! strings.HasPrefix(errmsg, "Error compiling 'model':") {
		t.Errorf("Error is not \"Error compiling 'model'\"")
	}
}

func TestAccusedDoesntCompile(t *testing.T) {
	_, err := evalWithInputs(Minimal, Wrong, []string{""})
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
		if r.Veredict != ACCEPT {
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
			if r.Veredict != WRONG_ANSWER {
				t.Errorf("Veredict should be WRONG_ANSWER (test %d)", i)
			} 
		} else {
			if r.Veredict != ACCEPT {
				t.Errorf("Veredict should be ACCEPT (test %d)", i)
			}
		}
	}
}