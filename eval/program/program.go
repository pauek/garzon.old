
package program

import (
	"os"
	"os/exec"
	"syscall"
	"io"
	"io/ioutil"
	"fmt"
	"crypto/sha1"
	"log"
	"strings"
	"bytes"
	
	"garzon/eval/program/lang"
)

// utils //

func _sha1(s string) string {
	hash := sha1.New()
	io.WriteString(hash, s)
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func prefix(s string, length int) string {
	max := length
	suffix := "..."
	if len(s) < length {
      max = len(s)
		suffix = ""
   }
	return strings.Replace(s[:max], "\n", `\n`, -1) + suffix
}

// Tests /////////////////////////////////////////////////////////////

type Result struct {
	Veredict string
	Reason   interface{}
}

type ProgramTester interface {
	SetUp(*Context, *exec.Cmd) error
	CleanUp(*Context) error
	Veredict() Result
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

func (I *InputTester) Veredict() Result {
	if I.modelOut.String() == I.accusedOut.String() {
		return Result{Veredict: "Accept"}
	} 
	return Result{Veredict: "Wrong Answer"}
}

// Context /////////////////////////////////////////////////

type Program struct {
	Lang, Code string
}

type Constraints struct { 
	Memory, Time, FileSize int 
}

type Evaluation struct {
	Accused, Model Program
	Limits Constraints
}
	
type Context struct {
	dir  string // working directory
	mode string // current program: "model" or "accused"
	limits Constraints 
	lang map[string]string
	code map[string]string
}

func (C *Context) Dir()     string { return C.dir }
func (C *Context) ExecDir() string { return C.dir + "/eval" }
func (C *Context) Mode()    string { return C.mode }

func NewContext(dir string, ev *Evaluation) *Context {
	C := new(Context)
	C.dir = dir
	C.limits = ev.Limits
	C.lang = map[string]string {
		"model":   ev.Model.Lang,
		"accused": ev.Accused.Lang,
	}
	C.code = map[string]string {
		"model":   ev.Model.Code,
		"accused": ev.Accused.Code,
	}
	return C
}

func (C *Context) CreateDirectory() error {
	log.Printf("Creating directory '%s'", C.dir)
	if err := os.RemoveAll(C.dir); err != nil {
		return fmt.Errorf("Couldn't remove directory '%s'", C.dir)
	}
	for _, subdir := range []string{"", "/.model", "/.accused", "/eval"} {
		if err := os.Mkdir(C.dir + subdir, 0700); err != nil {
			return fmt.Errorf("Couldn't make directory '%s'", C.dir + subdir)
		}
	}
	return nil
}

func (C *Context)	WriteAndCompile(whom string) error {
	L, ok := lang.Languages[C.lang[whom]]
	if ! ok {
		return fmt.Errorf("Unsupported language '%s'", C.lang[whom])
	} 
	codefile := fmt.Sprintf("%s/.%s/code.%s", C.dir, whom, L.Extension)
	if err := ioutil.WriteFile(codefile, []byte(C.code[whom]), 0600); err != nil {
		return fmt.Errorf("Couldn't write %s file '%s'", whom, codefile)
	}
	exefile  := fmt.Sprintf("%s/.%s/exe", C.dir, whom)
	log.Printf("Compiling '%s' ('%s')", codefile, prefix(C.code[whom], 30))
	if err := L.Functions.Compile(codefile, exefile); err != nil {
		os.RemoveAll(C.dir)
		return fmt.Errorf("Error compiling '%s': %v", whom, err)
	}
	return nil
}

func (C *Context) SwitchTo(whom string) error {
	if whom != "model" && whom != "accused" {
		return fmt.Errorf("Program '%s' not known")
	}
	from := fmt.Sprintf("%s/.%s/exe", C.dir, whom)
	to   := fmt.Sprintf("%s/eval/exe", C.dir)
	cmd := exec.Command("cp", from, to)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Couldn't copy '%s' to '%s'", from, to)
	} 
	C.mode = whom
	return nil
}

func (C *Context) MakeCommand() (cmd *exec.Cmd) {
	args := []string{}
	addOption := func (flag string, val int) {
		if val > 0 {
			args = append(args, flag)
			args = append(args, fmt.Sprintf("%d", val))
		}
	}
	addOption("-m", C.limits.Memory)
	addOption("-t", C.limits.Time)
	addOption("-f", C.limits.FileSize)
	if C.mode == "accused" {
		args = append(args, "-a")
	}
	args = append(args, C.dir + "/eval")
   cmd = exec.Command("grz-jail", args...)
	cmd.Dir = C.ExecDir()
	return
}

func (C *Context) Destroy() error {
	if err := os.RemoveAll(C.dir); err != nil {
		return fmt.Errorf("Couldn't remove directory '%s': %s", C.dir, err)
	}
	log.Printf("Removed directory '%s'\n", C.dir)
	return nil
}

// ProgramEvaluator //////////////////////////////////////////////////

var Evaluator *ProgramEvaluator

type ProgramEvaluator struct {
	BaseDir string
	Contexts map[string]*Context
}

func init() {
	Evaluator = new(ProgramEvaluator)
	Evaluator.BaseDir  = os.Getenv("HOME")
	Evaluator.Contexts = make(map[string]*Context)
}

func (E *ProgramEvaluator) StartEvaluation(ev Evaluation, ID *string) error {
	id  := _sha1(ev.Accused.Code)
	C := NewContext(E.BaseDir + "/" + id, &ev)
	if err := C.CreateDirectory(); err != nil { 
		return err 
	}
	if err := C.WriteAndCompile("model"); err != nil { 
		return err 
	}
	if err := C.WriteAndCompile("accused"); err != nil {
		return err
	}
	E.Contexts[id] = C
	*ID = id
	return nil
}

type TestInfo struct {
	EvalID string
	Test ProgramTester
}

func getExitStatus(err error) int {
	exiterror, ok := err.(*exec.ExitError)
	if ! ok { 
		log.Fatalf("Cannot get ProcessState") 
	}
	status, ok := exiterror.Sys().(syscall.WaitStatus)
	if ! ok { 
		log.Fatalf("Cannot get syscall.WaitStatus") 
	}
	return status.ExitStatus()
}

func (E *ProgramEvaluator) RunTest(T TestInfo, R *Result) (err error) {
	C, ok := E.Contexts[T.EvalID]
	if ! ok {
		return fmt.Errorf("Evaluation ID '%s' not found", T.EvalID)
	}
	
	runtest := func (whom string) bool {
		if err = C.SwitchTo(whom); err != nil { 
			return false 
		}
		cmd := C.MakeCommand()
		if err = T.Test.SetUp(C, cmd); err != nil { 
			return false 
		}
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		log.Printf("Executing '%s'", whom)
		if err = cmd.Run(); err != nil {
			code := getExitStatus(err)
			if code == 1 { // Execution Failed
				err = nil
				R.Veredict = strings.Replace(stderr.String(), "\n", "", -1)
			}
			return false
		}
		if err = T.Test.CleanUp(C); err != nil { 
			return false 
		}
		return true
	}
	if ! runtest("model")   { return }
	if ! runtest("accused") { return }

	*R = T.Test.Veredict()
	return nil
}

func (E *ProgramEvaluator) EndEvaluation(EvalID string, ok *bool) error {
	*ok = false
	C, found := E.Contexts[EvalID]
	if ! found {
		return fmt.Errorf("Evaluation ID '%s' not found", EvalID)
	}
	if err := C.Destroy(); err != nil {
		return err
	}
	*ok = true
	return nil
}
