package programming

import (
	"github.com/pauek/garzon/db"
)

func init() {
	db.Register("prog.Evaluator", Evaluator{})
	db.Register("prog.VeredictDetails", VeredictDetails{})
	db.Register("prog.test.Input", InputTester{})
	db.Register("prog.test.Files", FilesTester{})
	db.Register("prog.test.Result", TestResult{})
	db.Register("prob.SimpleReason", SimpleReason{})
	db.Register("prob.GoodVsBadReason", GoodVsBadReason{})
	db.Register("prog.test.[]Result", []TestResult{})
}
