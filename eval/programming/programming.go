
package programming

import (
	"os/exec"
)

type Problem struct {
	Title    string
	Solution Code
	Limits   Constraints
	Tests    []Tester
}

type Code struct {
	Lang, Text string
}

type Constraints struct { 
	Memory, Time, FileSize int 
}

type Tester interface {
	SetUp(*context, *exec.Cmd) error
	CleanUp(*context) error
	Veredict() TestResult
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