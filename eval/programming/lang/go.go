package lang

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

func init() {
	Register(&Language{"Go", []string{".go"}, new(Go)})
}

type Go struct{}

func (L *Go) Compile(infile, outfile string) error {
	cmd := exec.Command("go", "build", "-o", outfile, infile)
	var out bytes.Buffer
	cmd.Stderr = &out
	err := cmd.Run()
	if err != nil {
		output := strings.Replace(out.String(), infile, "code.go", -1)
		return &CompilationError{Output: output}
	}
	return nil
}

func (L *Go) Execute(filename, input string) (string, error) {
	cmd := exec.Command("./" + filename)
	cmd.Stdin = strings.NewReader(input)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("Go.Execute: %v", err)
	}
	return out.String(), nil
}
