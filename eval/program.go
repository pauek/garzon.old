
package eval

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
	
	"garzon/eval/lang"
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
	SetUp(ProgramEvaluation, *exec.Cmd) error
	CleanUp(ProgramEvaluation) error
	Veredict() Result
}

type InputTester struct {
	Input string

	modelOut, accusedOut bytes.Buffer
}

func (I *InputTester) SetUp(E ProgramEvaluation, cmd *exec.Cmd) error {
	log.Printf("Testing input '%s'\n", prefix(I.Input, 20))
	cmd.Stdin  = strings.NewReader(I.Input)
	switch (E.Mode()) {
	case "model":
		cmd.Stdout = &I.modelOut
	case "accused":
		cmd.Stdout = &I.accusedOut
	default:
		return fmt.Errorf("Unknown mode '%s'\n", E.Mode())
	}
	return nil
}

func (I *InputTester) CleanUp(E ProgramEvaluation) error {
	
	return nil
}

func (I *InputTester) Veredict() Result {
	if I.modelOut.String() == I.accusedOut.String() {
		return Result{Veredict: "Accept"}
	} 
	return Result{Veredict: "Wrong Answer"}
}

// ProgramEvaluation /////////////////////////////////////////////////

type Program struct {
	Lang, Code string
}

type ProgramEvaluation struct {
	Accused, Model Program
	Limits struct { Memory, Time, FileSize int }
	
	mode string // current program: "model" or "accused"
	dir  string // working directory
}

func (E *ProgramEvaluation) EvaluationDir() string {
	return E.dir 
}

func (E *ProgramEvaluation) ExecutionDir() string { 
	return E.dir + "/eval"
}

func (E *ProgramEvaluation) Mode() string {
	return E.mode
}

func (E *ProgramEvaluation) SwitchTo(whom string) error {
	if whom != "model" && whom != "accused" {
		return fmt.Errorf("Program '%s' not known")
	}
	from := fmt.Sprintf("%s/.%s/exe", E.dir, whom)
	to   := fmt.Sprintf("%s/eval/exe", E.dir)
	cmd := exec.Command("cp", from, to)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Couldn't copy '%s' to '%s'", from, to)
	} 
	E.mode = whom
	return nil
}

func (E *ProgramEvaluation) MakeCommand() (cmd *exec.Cmd) {
	args := []string{}
	addOption := func (flag string, val int) {
		if val > 0 {
			args = append(args, flag)
			args = append(args, fmt.Sprintf("%d", val))
		}
	}
	addOption("-m", E.Limits.Memory)
	addOption("-t", E.Limits.Time)
	addOption("-f", E.Limits.FileSize)
	if E.mode == "accused" {
		args = append(args, "-a")
	}
	args = append(args, E.dir + "/eval")
   cmd = exec.Command("grz-jail", args...)
	cmd.Dir = E.ExecutionDir()
	return
}

// ProgramEvaluator //////////////////////////////////////////////////

var Evaluator *ProgramEvaluator

type ProgramEvaluator struct {
	BaseDir string
}

func init() {
	Evaluator = new(ProgramEvaluator)
	Evaluator.BaseDir = os.Getenv("HOME")
	evaluations = make(map[string]ProgramEvaluation)
}

var evaluations map[string]ProgramEvaluation

func (E *ProgramEvaluator) StartEvaluation(P ProgramEvaluation, ID *string) error {
	// 1. Determine languages
	Lang := make(map[string]string)
	Code := make(map[string]string)
	Lang["model"]   = P.Model.Lang
	Lang["accused"] = P.Accused.Lang
	Code["model"]   = P.Model.Code
	Code["accused"] = P.Accused.Code

	// 2. Prepare Directory
	id  := _sha1(Code["accused"])
	dir := E.BaseDir + "/" + id
	P.dir = dir
	log.Printf("Preparing directory '%s'", dir)
	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("Couldn't remove directory '%s'", dir)
	}
	for _, subdir := range []string{"", "/.model", "/.accused", "/eval"} {
		if err := os.Mkdir(dir + subdir, 0700); err != nil {
			return fmt.Errorf("Couldn't make directory '%s'", dir + subdir)
		}
	}

	// 3. Write and Compile Accused and Model
	writeAndCompile := func (whom string) error {
		language, ok := lang.Languages[Lang[whom]]
		if ! ok {
			return fmt.Errorf("Unsupported language '%s'", Lang[whom])
		} 
		codefile := fmt.Sprintf("%s/.%s/code.%s", dir, whom, language.Extension)
		exefile  := fmt.Sprintf("%s/.%s/exe", dir, whom)
		if err := ioutil.WriteFile(codefile, []byte(Code[whom]), 0600); err != nil {
			return fmt.Errorf("Couldn't write %s file '%s'", whom, codefile)
		}
		log.Printf("Compiling '%s' ('%s')", codefile, prefix(Code[whom], 30))
		if err := language.Functions.Compile(codefile, exefile); err != nil {
			os.RemoveAll(dir)
			return fmt.Errorf("Error compiling '%s': %v", whom, err)
		}
		return nil
	}
	if err := writeAndCompile("model"); err != nil { 
		return err 
	}
	if err := writeAndCompile("accused"); err != nil {
		return err
	}

	// 4. Store Evaluation object
	evaluations[id] = P
	*ID = id
	return nil
}

type TestInfo struct {
	EvaluationID string
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
	P, ok := evaluations[T.EvaluationID]
	if ! ok {
		return fmt.Errorf("Evaluation ID '%s' not found", T.EvaluationID)
	}
	
	runtest := func (whom string) bool {
		if err = P.SwitchTo(whom); err != nil { 
			return false 
		}
		cmd := P.MakeCommand()
		if err = T.Test.SetUp(P, cmd); err != nil { 
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
		if err = T.Test.CleanUp(P); err != nil { 
			return false 
		}
		return true
	}
	if ! runtest("model")   { return }
	if ! runtest("accused") { return }

	*R = T.Test.Veredict()
	return nil
}

func (E *ProgramEvaluator) EndEvaluation(EvaluationID string, ok *bool) error {
	*ok = false
	P, found := evaluations[EvaluationID]
	if ! found {
		return fmt.Errorf("Evaluation ID '%s' not found", EvaluationID)
	}
	if err := os.RemoveAll(P.dir); err != nil {
		return fmt.Errorf("Couldn't remove directory '%s': %s", P.dir, err)
	}
	log.Printf("Removed directory '%s'\n", P.dir)
	*ok = true
	return nil
}
