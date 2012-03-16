
package eval

import (
	"os"
	"os/exec"
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

// Veredicts /////////////////////////////////////////////////////////

const (
	ACCEPT = 0
   COMPILATION_ERROR = 1
   EXECUTION_ERROR = 2
   WRONG_ANSWER = 3
)

// Tests /////////////////////////////////////////////////////////////

type Result struct {
	Veredict int
	Reason   interface{}
}

type ProgramTester interface {
	SetUp(ProgramEvaluation) error
	CleanUp(ProgramEvaluation) error
	Run(ProgramEvaluation, *exec.Cmd) error
	Veredict() Result
}

type InputTester struct {
	Input string
	output map[string]string
}

func (I *InputTester) SetUp(E ProgramEvaluation) error {
	return nil
}

func (I *InputTester) CleanUp(E ProgramEvaluation) error {
	return nil
}

func (I *InputTester) Run(E ProgramEvaluation, cmd *exec.Cmd) error {
	if (E.Mode() == "model") {
		I.output = make(map[string]string)
	}
	var output, status bytes.Buffer
	cmd.Stdin  = strings.NewReader(I.Input)
	cmd.Stdout = &output
	cmd.Stderr = &status
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s", status.String())
		return fmt.Errorf("Couldn't execute '%s': %s", E.Mode(), err)
	} 
	I.output[E.Mode()] = output.String()
	return nil
}

func (I *InputTester) Veredict() Result {
	if I.output["model"] == I.output["accused"] {
		return Result{Veredict: ACCEPT}
	} 
	return Result{Veredict: WRONG_ANSWER, Reason: I.output}
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

func (E *ProgramEvaluation) RunCommand() (cmd *exec.Cmd) {
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

func (E *ProgramEvaluator) RunTest(T TestInfo, Veredict *Result) error {
	P, ok := evaluations[T.EvaluationID]
	if ! ok {
		return fmt.Errorf("Evaluation ID '%s' not found", T.EvaluationID)
	}
	if err := P.SwitchTo("model");   err != nil { return err }
	if err := T.Test.SetUp(P);       err != nil { return err }
	cmd := P.RunCommand()
	log.Printf("Executing 'model'")
	if err := T.Test.Run(P, cmd);    err != nil { return err }
	if err := T.Test.CleanUp(P);     err != nil { return err }
	if err := P.SwitchTo("accused"); err != nil { return err }
	if err := T.Test.SetUp(P);       err != nil { return err }
	cmd  = P.RunCommand()
	log.Printf("Executing 'accused'")
	if err := T.Test.Run(P, cmd);    err != nil { return err }
	if err := T.Test.CleanUp(P);     err != nil { return err }
	*Veredict = T.Test.Veredict()
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
