
package programming

import (
	"encoding/gob"
	"garzon/db"
)

func Register() {
	db.Register("prog.Evaluator", Evaluator{})
	db.Register("prog.VeredictDetails", VeredictDetails{})
	db.Register("prog.test.Input", InputTester{})
	db.Register("prog.test.Files", FilesTester{})
	db.Register("prog.test.Result", TestResult{})
	gob.Register(Evaluator{})
	gob.Register(InputTester{})
	gob.Register(FilesTester{})
	gob.Register(VeredictDetails{})
	gob.Register(TestResult{})
	gob.Register([]TestResult{})
}

