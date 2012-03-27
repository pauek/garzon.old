
package programming

import (
	"os/exec"
	"fmt"
	"log"
	"strings"
	"bytes"

	"garzon/db"
)

// InputTester ///////////////////////////////////////////////////////

func init() {
	db.Register("Input", InputTester{})
}

type InputTester struct {
	Input string
	state *InputTesterState
}

type InputTesterState struct {
	modelOut, accusedOut bytes.Buffer
}

func (I *InputTester) Prepare(C *context) {
	C.State = new(InputTesterState)
}

func (I *InputTester) SetUp(C *context, cmd *exec.Cmd) error {
	log.Printf("Testing input '%s'\n", prefix(I.Input, 20))
	cmd.Stdin  = strings.NewReader(I.Input)
	state := C.State.(*InputTesterState)
	switch (C.Mode()) {
	case "model":
		cmd.Stdout = &state.modelOut
	case "accused":
		cmd.Stdout = &state.accusedOut
	default:
		return fmt.Errorf("Unknown mode '%s'\n", C.Mode())
	}
	return nil
}

func (I *InputTester) CleanUp(*context) error {
	return nil
}

func (I *InputTester) Veredict(C *context) TestResult {
	state := C.State.(*InputTesterState)
	if state.modelOut.String() == state.accusedOut.String() {
		return TestResult{Veredict: "Accept"}
	} 
	return TestResult{Veredict: "Wrong Answer"}
}
