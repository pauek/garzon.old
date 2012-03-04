
package lang

import (
	"fmt"
	"os/exec"
	"bytes"
	"strings"
)

type Cpp string

func (L *Cpp) Compile(ID string) (string, error) {
	cmd := exec.Command("g++", "-static", "-o", "exe", "code.cc")
	var out bytes.Buffer
	cmd.Stderr = &out
	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
		return out.String(), fmt.Errorf("Compilation failed")
	}
	return "", nil
}

func (L *Cpp) Execute(ID string, input string) (string, error) {
	cmd := exec.Command("./exe")
	cmd.Stdin = strings.NewReader(input)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("Execution failed: ", err)
	}
	return out.String(), nil
}
