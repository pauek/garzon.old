package programming

import (
	"bytes"
	"fmt"
	"github.com/pauek/garzon/db"
	"io/ioutil"
	"log"
	"os/exec"
	"strings"
)

// InputTester

// An InputTester tests a program by feeding it some input and
// checking that the output is the same as the model's output.
type InputTester struct {
	Input string
	state *InputTesterState
}

type InputTesterState struct {
	modelOut, accusedOut bytes.Buffer
}

func (I InputTester) Prepare(C *context) {
	C.State = new(InputTesterState)
}

func (I InputTester) SetUp(C *context, cmd *exec.Cmd) error {
	log.Printf("Testing input '%s'\n", prefix(I.Input, 20))
	cmd.Stdin = strings.NewReader(I.Input)
	state := C.State.(*InputTesterState)
	switch C.Mode() {
	case "model":
		cmd.Stdout = &state.modelOut
	case "accused":
		cmd.Stdout = &state.accusedOut
	default:
		return fmt.Errorf("Unknown mode '%s'\n", C.Mode())
	}
	return nil
}

func (I InputTester) CleanUp(*context) error {
	return nil
}

func noendl(s string) string {
	return strings.Replace(s, "\n", `\n`, -1)
}

func spacesVisible(a string) (s string) {
	replacements := []struct{ from, to string }{
		{" ", "\u2423"},
		{"\n", "\u21B2\n"},
	}
	s = a
	for _, r := range replacements {
		s = strings.Replace(s, r.from, r.to, -1)
	}
	return
}

func (I InputTester) Veredict(C *context) TestResult {
	S := C.State.(*InputTesterState)
	a, b := S.modelOut.String(), S.accusedOut.String()
	if a == b {
		return TestResult{Veredict: "Accepted"}
	}
	a = spacesVisible(a)
	b = spacesVisible(b)
	return TestResult{
		Veredict: "Wrong Answer",
		Reason:   db.Obj{&GoodVsBadReason{a, b}},
	}
}

func (I *InputTester) ReadFrom(path string) error {
	text, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("InputTester.ReadFrom: cannot read '%s': %s\n", path, err)
	}
	I.Input = string(text)
	return nil
}
