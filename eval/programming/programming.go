
package programming

import (
	"os/exec"
	"garzon/db"
)

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

type Readable interface {
	ReadFrom(path string) error
}

type VeredictDetails struct {
	Results []TestResult
}

type TestResult struct {
	Veredict string
	Reason db.Obj
}