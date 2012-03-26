
package programming

import (
	"os/exec"
	"fmt"
	"log"
	"strings"
	"bytes"

	"garzon/db"
	"garzon/eval"
)

// InputTester ///////////////////////////////////////////////////////

func init() {
	db.Register("Input", InputTester{})
}

type InputTester struct {
	Input string
	modelOut, accusedOut bytes.Buffer
}

func (I *InputTester) SetUp(C *Context, cmd *exec.Cmd) error {
	log.Printf("Testing input '%s'\n", prefix(I.Input, 20))
	cmd.Stdin  = strings.NewReader(I.Input)
	switch (C.Mode()) {
	case "model":
		cmd.Stdout = &I.modelOut
	case "accused":
		cmd.Stdout = &I.accusedOut
	default:
		return fmt.Errorf("Unknown mode '%s'\n", C.Mode())
	}
	return nil
}

func (I *InputTester) CleanUp(*Context) error {
	return nil
}

func (I *InputTester) Veredict() eval.Result {
	if I.modelOut.String() == I.accusedOut.String() {
		return eval.Result{Veredict: "Accept"}
	} 
	return eval.Result{Veredict: "Wrong Answer"}
}
