
package eval

import (
	"log"
	"time"
	"encoding/gob"
	"garzon/db"
)

type Submission struct {
	Problem  *Problem
	Solution  string
	State     string
	Submitted time.Time
	Resolved  time.Time
}

type Problem struct {
	Title, StatementID, Solution string
	Evaluator db.Obj
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
	log.Printf("Received Problem: %+v\n", S.Problem)
	ev := S.Problem.Evaluator.Obj.(Evaluator)
	*V = ev.Evaluate(S.Problem, S.Solution)
	log.Printf("Result: %+v\n", V)
	return nil
}

// RPC
func init() {
	db.Register("eval.Problem", Problem{})
	gob.Register(Problem{})
	gob.Register(Veredict{})
	gob.Register(Submission{})
}
