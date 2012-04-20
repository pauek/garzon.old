package programming

import (
	"bytes"
	"fmt"
	"github.com/pauek/garzon/db"
	"os/exec"
)

type Evaluator struct {
	Limits Constraints
	Tests  []db.Obj
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

func (vd VeredictDetails) String() string {
	var b bytes.Buffer
	for i, r := range vd.Results {
		fmt.Fprintf(&b, "%d. %s\n", i+1, r)
	}
	return b.String()
}

type TestResult struct {
	Veredict string
	Reason   db.Obj
}

func (tr TestResult) String() string {
	var b bytes.Buffer
	fmt.Fprintf(&b, "%s", tr.Veredict)
	if tr.Veredict != "Accepted" {
		if tr.Reason.Obj != nil {
			fmt.Fprintf(&b, ":\n%v\n", tr.Reason.Obj)
		} else {
			fmt.Fprintf(&b, "\n")
		}
	}
	return b.String()
}

type SimpleReason struct {
	Message string
}

func (sr SimpleReason) String() string {
	return sr.Message
}
