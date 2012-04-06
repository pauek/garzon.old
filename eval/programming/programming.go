
package programming

import (
	"fmt"
	"bytes"
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

func (vd *VeredictDetails) String() string {
	var b bytes.Buffer
	for i, r := range vd.Results {
		fmt.Fprintf(&b, "%d. %s\n", i, r)
	}
	return b.String()
}

type TestResult struct {
	Veredict string
	Reason db.Obj
}

func (tr TestResult) String() string {
	var b bytes.Buffer
	fmt.Fprintf(&b, "%s", tr.Veredict)
	if tr.Veredict != "Accept" {
		fmt.Fprintf(&b, ":\n%s\n", tr.Reason.Obj)
	}
	return b.String()
}