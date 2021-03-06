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
	Submitted time.Time
	Resolved  time.Time
	Veredict  Veredict
}

func (S *Submission) Hora() string {
	t := S.Submitted
	return fmt.Sprintf("%d/%d/%d a les %d:%d", t.Day(), t.Month(), t.Year(), t.Hour(), t.Minute())
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
	Evaluate(Problem *Problem, Solution string, progress chan<- string) Veredict
}

type Response struct {
	Status   string
	Veredict *Veredict
}

type Eval bool

func Submit(S Submission, V *Veredict, progress chan<- string) {
	ev, ok := S.Problem.Evaluator.Obj.(Evaluator)
	if !ok {
		progress <- "Error: wrong evaluator"
		return
	}
	*V = ev.Evaluate(S.Problem, S.Solution, progress)
	progress <- "Resolved"
}

func init() {
	db.Register("eval.Problem", Problem{})
	db.Register("eval.Submission", Submission{})
	db.Register("eval.Veredict", Veredict{})
}
