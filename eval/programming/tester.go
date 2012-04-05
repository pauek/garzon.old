
package programming

import (
	"os"
	"os/exec"
	"io/ioutil"
	"fmt"
	"log"
	"bytes"
	"strings"
	"path/filepath"
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

func (I InputTester) CleanUp(*context) error {
	return nil
}

func (I InputTester) Veredict(C *context) TestResult {
	state := C.State.(*InputTesterState)
	if state.modelOut.String() == state.accusedOut.String() {
		return TestResult{Veredict: "Accept"}
	} 
	return TestResult{Veredict: "Wrong Answer"}
}

func (I InputTester) ReadFrom(path string) error {
	text, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("InputTester.ReadFrom: cannot read '%s': %s\n", path, err)
	}
	I.Input = string(text)
	return nil
}

// A FilesTester creates some input files, and checks that some 
// output files are created by the program and have the same
// contents as the files created by the model program
type FilesTester struct {
	InputFiles  []FileInfo
	OutputFiles []FileInfo
	state *InputTesterState
}

type FileInfo struct {
	RelPath  string // relative path with respecto to 'exe'
	Contents string // contents of the file
}

type FileTesterState struct {
	InputTesterState
	modelOutFiles, accusedOutFiles [][]byte
}

func (I FilesTester) Prepare(C *context) {
	state := new(FileTesterState)
	n := len(I.OutputFiles)
	state.modelOutFiles   = make([][]byte, n)
	state.accusedOutFiles = make([][]byte, n)
	C.State = state
}

func (I FilesTester) SetUp(C *context, cmd *exec.Cmd) error {
	log.Printf("Testing Files '%s'\n", C.Mode())
	state := C.State.(*FileTesterState)
	switch (C.Mode()) {
	case "model":
		cmd.Stdout = &state.modelOut
	case "accused":
		cmd.Stdout = &state.accusedOut
	default:
		return fmt.Errorf("Unknown mode '%s'\n", C.Mode())
	}
	for _, finfo := range I.InputFiles {
		path := C.ExecDir() + "/" + finfo.RelPath
		if err := ioutil.WriteFile(path, []byte(finfo.Contents), 0600); err != nil {
			return fmt.Errorf("FilesTester: Cannot create file '%s': %s\n", path, err)
		}
	}
	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil // ??
}

func (I FilesTester) CleanUp(C *context) (err error) {
	state := C.State.(*FileTesterState)
	// Erase input files
	for _, finfo := range I.InputFiles {
		path := C.ExecDir() + "/" + finfo.RelPath
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("FilesTester: Cannot remove file '%s': %s\n", path, err)
		}
	}
	// Keep output files
	for i, finfo := range I.OutputFiles {
		path := C.ExecDir() + "/" + finfo.RelPath
		var b []byte
		if fileExists(path) {
			if b, err = ioutil.ReadFile(path); err != nil {
				return fmt.Errorf("FilesTester: Cannot read '%s': %s\n", path, err)
			}
		}
		switch C.Mode() {
		case "model": 
			state.modelOutFiles[i] = b
		case "accused": 
			state.accusedOutFiles[i] = b
		}
		// erase output file
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("FilesTester: Cannot remove file '%s': %s\n", path, err)
		}
	}
	return nil
}

func equalBytes(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func (I FilesTester) Veredict(C *context) TestResult {
	state := C.State.(*FileTesterState)
	n := len(I.OutputFiles)
	// TODO: Compare with content in I.OutputFiles[i].Content!
	for i := 0; i < n; i++ {
		a := state.modelOutFiles[i]
		b := state.accusedOutFiles[i]
		if !equalBytes(a, b) {
			return TestResult{Veredict: "Wrong Answer" /* TODO: Add Reason! */}
		} 
	}
	return TestResult{Veredict: "Accept"}
}

func (I FilesTester) ReadFrom(path string) (err error) {
	I.InputFiles,  err = readFiles(path, "in")
	if err != nil { return err }
	I.OutputFiles, err = readFiles(path, "out")
	if err != nil { return err }
	return nil
}

func readFiles(path, ext string) (Files []FileInfo, err error) {
	matches, err := filepath.Glob(path + "/*." + ext)
	if err != nil {
		err = fmt.Errorf("readFiles: cannot glob '*.%s': %s\n", ext, err)
		return
	}
	Files = make([]FileInfo, len(matches))
	for i, m := range matches {
		p1 := len(path) + 1
		p2 := strings.LastIndex(m, ext) - 1
		relpath := m[p1:p2]
		var contents []byte
		contents, err = ioutil.ReadFile(m)
		if err != nil {
			err = fmt.Errorf("readFiles: Cannot read '%s': %s\n", m, err)
			return
		}
		Files[i] = FileInfo{RelPath: relpath, Contents: string(contents)}
	}
	return
}
