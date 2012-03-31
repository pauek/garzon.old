
package main

import (
	"log"
	"garzon/db"
	"garzon/eval"
	prog "garzon/eval/programming"
)

var submissions chan eval.Submission

func init() {
	submissions = make(chan eval.Submission)
	prog.Register()
}

func main() {
	done := make(chan bool)
	go evaluate(Account{user: "user", host: "garzon1", port: 50000}, done)

	const minimal = `int main() {}`

	problem := &eval.Problem{
		Title: "Doesn't matter...",
		Solution: minimal, // FIXME: prog.Code{Lang: "c++", Text: minimal},
	   Evaluator: db.Obj{
			&prog.Evaluator{
				Tests: []db.Obj{{&prog.InputTester{Input: ""}}},
			},
		},
	}
	submissions <- eval.Submission{
	   Problem: problem,
	   Solution: minimal, // FIXME: prog.Code{Lang: "c++", Text: minimal},
	}
	log.Printf("Submitted!")
	<- done
}