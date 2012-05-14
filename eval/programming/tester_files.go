package programming

import (
	"fmt"
	"github.com/pauek/garzon/db"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

func init() {
	db.Register("prog.test.Files", FilesTester{})
}

// A FilesTester creates some input files, and checks that some 
// output files are created by the program and have the same
// contents as the files created by the model program
type FilesTester struct {
	Input       string
	InputFiles  []FileInfo
	OutputFiles []FileInfo
	Options     map[string]bool `json:",omitempty"`
	state       *InputTesterState
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
	state.modelOutFiles = make([][]byte, n)
	state.accusedOutFiles = make([][]byte, n)
	C.State = state
}

func (I FilesTester) SetUp(C *context, cmd *exec.Cmd) error {
	log.Printf("Testing Files '%s'\n", C.Mode())
	state := C.State.(*FileTesterState)
	cmd.Stdin = strings.NewReader(I.Input)
	switch C.Mode() {
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

func sortLines(a string) string {
	if a[len(a)-1] == '\n' {
		a = a[:len(a)-1]
	}
	lines := strings.Split(a, "\n")
	sort.Strings(lines)
	return strings.Join(lines, "\n") + "\n"
}

func blocks(a string, sorted bool) string {
	blocks := strings.Split(a, "\n\n")
	if blocks[len(blocks)-1] == "" {
		blocks = blocks[:len(blocks)-1]
	}
	if sorted {
		for i := range blocks {
			blocks[i] = sortLines(blocks[i])
		}
	}
	return strings.Join(blocks, "\n\n") + "\n\n"
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
	a, b := state.modelOut.String(), state.accusedOut.String()
	if I.Options["blocks"] {
		a = blocks(a, I.Options["sort"])
		b = blocks(b, I.Options["sort"])
	} else if I.Options["sort"] {
		a = sortLines(a)
		b = sortLines(b)
	}
	if a != b {
		return TestResult{
			Veredict: "Wrong Answer",
			Reason:   db.Obj{&GoodVsBadReason{seeSpace(a), seeSpace(b)}},
		}
	}
	return TestResult{Veredict: "Accepted"}
}

func (I *FilesTester) ReadFrom(path string) (err error) {
	if fileExists(path + "/in") {
		text, err := ioutil.ReadFile(path + "/in")
		if err != nil {
			return fmt.Errorf("FilesTester.ReadFrom: cannot read '%s/in': %s\n", path, err)
		}
		I.Input = string(text)
	}
	I.InputFiles, err = readFiles(path, "in")
	if err != nil {
		return err
	}
	I.OutputFiles, err = readFiles(path, "out")
	if err != nil {
		return err
	}
	if fileExists(path + "/options") {
		text, err := ioutil.ReadFile(path + "/options")
		if err != nil {
			return fmt.Errorf("FilesTester.ReadFrom: cannot read '%s/options': %s\n", path, err)
		}
		I.Options = make(map[string]bool)
		lines := strings.Split(string(text), "\n")
		for _, line := range lines {
			opt := strings.Trim(line, " \t")
			if opt != "" {
				I.Options[opt] = true
			}
		}
	}
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
