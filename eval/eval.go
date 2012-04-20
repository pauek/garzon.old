package eval

import (
	"encoding/gob"
	"fmt"
	"github.com/pauek/garzon/db"
	"time"
)

type Submission struct {
	User      string `json:",omitempty"`
	ProblemID string
	Problem   *Problem `json:"-"`
	Solution  string
	Status    string
	Submitted time.Time
	Resolved  time.Time
	Veredict  Veredict
}

type Problem struct {
	Title       string
	StatementID string
	Solution    string
	Evaluator   db.Obj
}

type Veredict struct {
	Message string
	Details db.Obj
}

type Evaluator interface {
	Evaluate(Problem *Problem, Solution string) Veredict
}

type DirReader interface {
	ReadDir(dir string, Problem *Problem) error
}

type Eval bool

func (E *Eval) Submit(S Submission, V *Veredict) error {
	ev, ok := S.Problem.Evaluator.Obj.(Evaluator)
	if !ok {
		return fmt.Errorf("Wrong Evaluator")
	}
	*V = ev.Evaluate(S.Problem, S.Solution)
	return nil
}

// RPC
func init() {
	db.Register("eval.Problem", Problem{})
	db.Register("eval.Submission", Submission{})
	db.Register("eval.Veredict", Veredict{})
	gob.Register(Problem{})
	gob.Register(Submission{})
	gob.Register(Veredict{})
}
