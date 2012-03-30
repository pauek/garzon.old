
package programming

import (
	"os/exec"
	"garzon/db"
)

func init() {
	db.Register("prog.Evaluator", Evaluator{})
}

type Evaluator struct {
	Limits   Constraints
	Tests    []db.Obj
}

type Code struct {
	Lang, Text string
}

type Constraints struct { 
	Memory, Time, FileSize int 
}

type Tester interface {
	Prepare(*context)
	SetUp(*context, *exec.Cmd) error
	CleanUp(*context) error
	Veredict(*context) TestResult
}

type VeredictDetails struct {
	Results []TestResult
}

type TestResult struct {
	Veredict string
	Reason interface{}
}