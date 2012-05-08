package programming

import (
	"bytes"
	"fmt"
	"github.com/pauek/garzon/db"
	"os/exec"
	"strings"
)

type Evaluator struct {
	Limits   Constraints
	Tests    []db.Obj
	progress chan string
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

type GoodVsBadReason struct {
	Good, Bad string
}

func sizes(lines []string) (int, int) {
	w, h := 0, 0
	for i, ln := range lines {
		h = i
		if len(ln) > w {
			w = len(ln)
		}
	}
	return w, h
}

func (r GoodVsBadReason) String() string {
	la := strings.Split(r.Good, "\n")
	lb := strings.Split(r.Bad, "\n")
	aw, ah := sizes(la)
	bw, bh := sizes(lb)
	h := ah
	if bh > h {
		h = bh
	}
	var buf bytes.Buffer
	fmt.Fprintf(&buf, fmt.Sprintf("\n%%-%ds     %%-%ds\n\n", aw, bw), "Good", "Bad")
	for i := 0; i < h; i++ {
		a := ""
		if i < len(la) {
			a = strings.Replace(la[i], " ", "\u2423", -1)
		}
		fmt.Fprintf(&buf, fmt.Sprintf("%%-%ds  |  ", aw), a)
		b := ""
		if i < len(lb) {
			b = strings.Replace(lb[i], " ", "\u2423", -1)
		}
		fmt.Fprintf(&buf, fmt.Sprintf("%%-%ds\n", bw), b)
	}
	return buf.String()
}
