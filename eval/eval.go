
package eval

import (
	"garzon/db"
)

type Submission struct {
	Problem *Problem
	Solution string
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

type Eval bool

func (E *Eval) Submit(S Submission, V *Veredict) error {
	ev := S.Problem.Evaluator.Obj.(Evaluator)
	*V = ev.Evaluate(S.Problem, S.Solution)
	return nil
}

// RPC
func init() {
	db.Register("eval.Problem", Problem{})
}



