
package programming

import (
	"os"
	"fmt"
	"log"
	"bytes"
	"strings"
	"os/exec"
	"syscall"
	"io/ioutil"

	"garzon/eval/programming/lang"
)


// context /////////////////////////////////////////////////

type context struct {
	dir  string // working directory
	mode string // current program: "model" or "accused"
	limits Constraints 
	lang map[string]string
	code map[string]string
}

func (C *context) Dir()     string { return C.dir }
func (C *context) ExecDir() string { return C.dir + "/eval" }
func (C *context) Mode()    string { return C.mode }

func newContext(dir string, sub *Submission) *context {
	C := new(context)
	C.dir = dir
	C.limits = sub.Problem.Limits
	C.lang = map[string]string {
		"model":   sub.Problem.Solution.Lang,
		"accused": sub.Accused.Lang,
	}
	C.code = map[string]string {
		"model":   sub.Problem.Solution.Text,
		"accused": sub.Accused.Text,
	}
	return C
}

func (C *context) CreateDirectory() error {
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

func (C *context)	WriteAndCompile(whom string) error {
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

func (C *context) SwitchTo(whom string) error {
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

func (C *context) MakeCommand() (cmd *exec.Cmd) {
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

func (C *context) Destroy() error {
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
}

func init() {
	Evaluator = new(ProgramEvaluator)
	Evaluator.BaseDir  = os.Getenv("HOME")
}

func (E *ProgramEvaluator) Submit(sub Submission) (R *Result) {
	C, err := E.prepareContext(sub)
	if err != nil {
		return &Result{Veredict: fmt.Sprintf("%s\n", err)}
	}
	numTests := len(sub.Problem.Tests)
	R = &Result{Results: make([]TestResult, numTests)}
	for i, tester := range sub.Problem.Tests {
		E.runTest(C, tester, &R.Results[i])
	}
	C.Destroy()
	return
}

func (E *ProgramEvaluator) prepareContext(sub Submission) (C *context, err error) {
	id  := hash(sub.Accused.Text)
	C = newContext(E.BaseDir + "/" + id, &sub)
	if err := C.CreateDirectory(); err != nil { 
		return nil, err 
	}
	if err := C.WriteAndCompile("model"); err != nil { 
		return nil, err 
	}
	if err := C.WriteAndCompile("accused"); err != nil {
		return nil, err
	}
	return C, nil
}

func (E *ProgramEvaluator) runTest(C *context, T Tester, R *TestResult) (err error) {
	runtest := func (whom string) bool {
		if err = C.SwitchTo(whom); err != nil { 
			return false 
		}
		cmd := C.MakeCommand()
		if err = T.SetUp(C, cmd); err != nil { 
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
		if err = T.CleanUp(C); err != nil { 
			return false 
		}
		return true
	}
	if ! runtest("model")   { return }
	if ! runtest("accused") { return }

	*R = T.Veredict()
	return nil
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
