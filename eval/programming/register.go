
package programming

import (
	"garzon/db"
)

func Register() {
	db.Register("prog.Evaluator", Evaluator{})
	db.Register("prog.test.Input", InputTester{})
	db.Register("prog.test.Files", FilesTester{})
}

