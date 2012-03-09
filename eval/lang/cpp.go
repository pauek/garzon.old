
package lang

import (
	"fmt"
	"os/exec"
	"bytes"
	"strings"
)

type Cpp string

func (L *Cpp) Compile(infile, outfile string) error {
	cmd := exec.Command("g++", "-static", "-o", outfile, infile)
	var out bytes.Buffer
	cmd.Stderr = &out
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("Cpp.Compile: %v", err)
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
