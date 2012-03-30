
package main

import (
	"garzon/db"
	"garzon/eval"
	prog "garzon/eval/programming"
)

var submissions chan eval.Submission

func main() {
	go evaluate(Account{user: "user", host: "garzon", port: 50000})

	const minimal = `int main() {}`

	submissions <- eval.Submission{
	   Problem: &eval.Problem{
			Title: "Doesn't matter...",
			Solution: minimal, // FIXME: prog.Code{Lang: "c++", Text: minimal},
		   Evaluator: db.Obj{
				&prog.Evaluator{
					Tests: []db.Obj{{&prog.InputTester{Input: ""}}},
				},
			},
		},
	   Solution: minimal, // FIXME: prog.Code{Lang: "c++", Text: minimal},
	}
}