
package programming

import (
	"os/exec"
	"garzon/eval"
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
	eval.Tester
	SetUp(*Context, *exec.Cmd) error
	CleanUp(*Context) error
}

type Submission struct {
	Problem *Problem
	Accused  Code
}