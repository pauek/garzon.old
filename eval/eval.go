package eval

import (
	"fmt"
	"github.com/pauek/garzon/db"
	"time"
)

type Submission struct {
	User      string `json:",omitempty"`
	ProblemID string
	Problem   *Problem
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

type Response struct {
	Status string
	Veredict *Veredict
}

type Eval bool

func Submit(S Submission, V *Veredict) error {
	ev, ok := S.Problem.Evaluator.Obj.(Evaluator)
	if !ok {
		return fmt.Errorf("Wrong Evaluator")
	}
	*V = ev.Evaluate(S.Problem, S.Solution)
	return nil
}

func init() {
	db.Register("eval.Problem", Problem{})
	db.Register("eval.Submission", Submission{})
	db.Register("eval.Veredict", Veredict{})
}
