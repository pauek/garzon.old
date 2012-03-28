
package programming

import (
	"os/exec"
	"garzon/db"
)

func init() {
	db.Register("prog.Problem", Problem{})
}

type Problem struct {
	Title    string
	Solution Code
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

type Submission struct {
	Problem *Problem
	Accused  Code
}

type Result struct {
	Veredict string
	Results []TestResult
}

type TestResult struct {
	Veredict string
	Reason interface{}
}