
package main

import (
	"log"
	"flag"
	"strings"
	"garzon/db"
	"garzon/eval"
	prog "garzon/eval/programming"
)

var submissions chan eval.Submission

func init() {
	submissions = make(chan eval.Submission)
	prog.Register()
}

func parseUserHost(userhost string) Account {
	parts := strings.Split(userhost, "@")
	if len(parts) != 2 {
		log.Fatal("Wrong user@host = '%s'\n", userhost)
	}
	return Account{user: parts[0], host: parts[1]}
}

func submitTestProblem() {
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
}

func submitDatabaseProblem(probid string) {
	problems, err := db.Get("localhost:5984", "problems")
	if err != nil {
		log.Fatalf("db.Get: %s\n", err)
	}
	P, _, err := problems.Get(probid)
	if err != nil {
		log.Fatalf("problems.Get: %s\n", err)
	}
	problem := P.(*eval.Problem)
	submissions <- eval.Submission{
		Problem: problem,
		Solution: problem.Solution,
	}
}

var copyfiles bool

func main() {
	flag.BoolVar(&copyfiles, "copy", false, "Copy files to remote accounts")
	flag.Parse()
	accounts := flag.Args()
	if len(accounts) < 1 {
		log.Fatal("Accounts must be 'user@host' (and >= 1)")
	}
	done := make(chan bool)
	for i, acc := range accounts {
		A := parseUserHost(acc)
		A.port = 50000 + i
		go evaluate(A, done)
	}

	for i := 0; i < 10; i++ {
		submitTestProblem()
		submitDatabaseProblem("cpp.ficheros.SumaEnteros")
	}
	close(submissions)

	for i := 0; i < len(accounts); i++ {
		<- done
	}
}