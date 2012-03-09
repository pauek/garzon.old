
package eval

import (
	"os"
	"os/exec"
	"io"
	"io/ioutil"
	"fmt"
	"crypto/sha1"
	"log"
	"bytes"
	"strings"
	
	"garzon/eval/lang"
)

func _sha1(s string) string {
	hash := sha1.New()
	io.WriteString(hash, s)
	return fmt.Sprintf("%x", hash.Sum(nil))
}

type Result int

type Test interface {
	Run(T ProgramEvaluation) (Result, error)
}

// InputTest /////////////////////////////////////////////////////////

type InputTest struct {
	Input string
}

func prefix(s string, length int) string {
	max := length
	suffix := "..."
	if len(s) < length {
      max = len(s)
		suffix = ""
   }
	return strings.Replace(s[:max], "\n", `↩`, -1) + suffix
}

func (I InputTest) Run(T ProgramEvaluation) (Result, error) {
	outputs  := make(map[string]*bytes.Buffer)
	for _, prog := range []string{"model", "accused"} {
		err := T.SwitchTo(prog) 
		if err != nil {
			return Result(-1), err
		}
		log.Printf("Running input test (%s): '%s'", prog, prefix(I.Input, 10));
		cmd := T.GetCommand()
		cmd.Stdin  = strings.NewReader(I.Input)
		var output bytes.Buffer
		cmd.Stdout = &output
		outputs[prog] = &output
		err = cmd.Run()
		if err != nil {
			return Result(-1), fmt.Errorf("Couldn't execute '%s': %v", prog, err)
		} else {
			log.Printf("Output: '%s'", prefix(output.String(), 10))
		}
	}
	if outputs["model"].String() != outputs["accused"].String() {
		return Result(1), nil
	} 
	return Result(0), nil
}

// ProgramEvaluation /////////////////////////////////////////////////

type Program struct {
	Lang, Code string
}

type ProgramEvaluation struct {
	Accused, Model Program
	Limits struct { Memory, Time, DiskSpace int }
	Tests []Test

	curr string // current program: "model" or "accused"
	dir  string // directory used for testing
}

type Results struct {
	Values []Result
}

func (T *ProgramEvaluation) cleanUp() {
}

func (T *ProgramEvaluation) GetCommand() *exec.Cmd {
	args := []string{}
	if T.Limits.Memory > 0 {
		args = append(args, "-m")
		args = append(args, fmt.Sprintf("%d", T.Limits.Memory))
	}
	if T.Limits.Time > 0 {
		args = append(args, "-t")
		args = append(args, fmt.Sprintf("%d", T.Limits.Time))
	}
	if T.Limits.DiskSpace > 0 {
		args = append(args, "-f")
		args = append(args, fmt.Sprintf("%d", T.Limits.DiskSpace))
	}
	if T.curr == "accused" {
		args = append(args, "-a")
	}
	args = append(args, "./exe")
   return exec.Command("grz-jail", args...)
}

func (T *ProgramEvaluation) SwitchTo(whom string) error {
	if whom != "model" && whom != "accused" {
		return fmt.Errorf("Program '%s' not known")
	}
	cmd := exec.Command("cp", whom, "exe")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Couldn't switch to '%s'", whom)
	} 
	// Borrar el fichero de syscalls cada vez!
	if whom == "model" {
		path := T.dir[1:] + "/exe"
		path  = strings.Replace(path, "/", "_", -1)
		path  = strings.Replace(path, "-", "_", -1)
		tracefilename := fmt.Sprintf("%s/.systrace/%s", os.Getenv("HOME"), path)
		_, err := os.Stat(tracefilename)
		if !os.IsNotExist(err) {
			cmd := exec.Command("rm", tracefilename)
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("Couldn't erase systrace file '%s'", tracefilename)
			} 
		}
	}
	T.curr = whom
	return nil
}

// ProgramEvaluator //////////////////////////////////////////////////

var ProgramEvaluator *ProgramEvaluatorType

type ProgramEvaluatorType struct {
	BaseDir string
}

func init() {
	ProgramEvaluator = new(ProgramEvaluatorType)
	ProgramEvaluator.BaseDir = os.Getenv("HOME")
}

func (E *ProgramEvaluatorType) Run(T ProgramEvaluation, R *Results) error {
	var aLang, mLang *lang.Language
	var err error
	var ok bool

	R = nil
	
	// 1. Determine languages
	aLang, ok = lang.Languages[T.Accused.Lang]
	if ! ok {
		return fmt.Errorf("Unsupported language '%s'", T.Accused.Lang)
	}
	mLang, ok = lang.Languages[T.Model.Lang]
	if ! ok {
		return fmt.Errorf("Unsupported language '%s'", T.Model.Lang)
	}

	// 2. Prepare Directory
	dir := E.BaseDir + "/" + _sha1(T.Accused.Code)
	T.dir = dir
	log.Printf("Preparing directory '%s'", dir)
	lastdir, _ := os.Getwd() // TODO: handle error?
	os.Mkdir(dir, 0700)
	os.Chdir(dir)
	defer func () {
		// 6. Clean Up
		os.Chdir(lastdir)
		/*
		log.Printf("Removing directory '%s'", dir)
		cmd := exec.Command("rm", "-rf", dir)
		if err := cmd.Run(); err != nil {
			log.Fatal("Couldn't remove directory %s", dir)
		}
      */
	} ()

	// 3. Write Accused and Model
	aFile := fmt.Sprintf("accused.%s", aLang.Extension)
	ioutil.WriteFile(aFile, []byte(T.Accused.Code), 0600)
	mFile := fmt.Sprintf("model.%s", mLang.Extension)
	ioutil.WriteFile(mFile, []byte(T.Model.Code), 0600)
	log.Printf("Written '%s' and '%s'", aFile, mFile)

	// 4. Compile Accused and Model
	log.Printf("Compiling %s ('%s')", aFile, prefix(T.Accused.Code, 30))
	err = aLang.Functions.Compile(aFile, "accused")
	if err != nil {
		return fmt.Errorf("Error compiling 'accused': %v", err)
	}
	log.Printf("Compiling %s ('%s')", mFile, prefix(T.Model.Code, 30))
	err = mLang.Functions.Compile(mFile, "model")
	if err != nil {
		return fmt.Errorf("Error compiling 'model': %v", err)
	}

	R = new(Results)
	R.Values = make([]Result, len(T.Tests))

	// 5. Run tests
	for i, test := range T.Tests {
		R.Values[i], err = test.Run(T)
		if err != nil {
			return fmt.Errorf("Error testing: %v", err)
		}
	} 

	// 6. Clean up [deferred on step 2]

	return nil		
}