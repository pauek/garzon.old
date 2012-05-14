package lang

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

func init() {
	Register(&Language{"C++", ".cc", new(Cpp)})
}

type Cpp struct{}

func (L *Cpp) Compile(infile, outfile string) error {
	cmd := exec.Command("g++", "-static", "-o", outfile, infile)
	var out bytes.Buffer
	cmd.Stderr = &out
	err := cmd.Run()
	if err != nil {
		output := strings.Replace(out.String(), infile, "code.cc", -1)
		return &CompilationError{Output: output}
	}
	return nil
}

func (L *Cpp) Execute(filename, input string) (string, error) {
	cmd := exec.Command("./" + filename)
	cmd.Stdin = strings.NewReader(input)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("Cpp.Execute: %v", err)
	}
	return out.String(), nil
}
