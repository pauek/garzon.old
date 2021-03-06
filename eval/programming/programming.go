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
	progress chan<- string
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

type Reader interface {
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

func (T *TestResult) GoodVsBad() (ok bool) {
	_, ok = T.Reason.Obj.(*GoodVsBadReason)
	return
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
	} else {
		fmt.Fprintf(&b, " [%+v]\n", tr.Reason)
	}
	return b.String()
}

type Performance struct {
	Seconds float32
	Megabytes float32
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
			a = la[i]
		}
		fmt.Fprintf(&buf, fmt.Sprintf("%%-%ds  |  ", aw), a)
		b := ""
		if i < len(lb) {
			b = lb[i]
		}
		fmt.Fprintf(&buf, fmt.Sprintf("%%-%ds\n", bw), b)
	}
	return buf.String()
}
