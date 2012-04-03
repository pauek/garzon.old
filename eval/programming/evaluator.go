
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
	"path/filepath"

	"garzon/db"
	"garzon/eval"
	"garzon/eval/programming/lang"
)


// context /////////////////////////////////////////////////

type context struct {
	dir  string // working directory
	mode string // current program: "model" or "accused"
	limits Constraints 
	lang map[string]string
	code map[string]string

	State interface{}
}

func (C *context) Dir()     string { return C.dir }
func (C *context) ExecDir() string { return C.dir + "/eval" }
func (C *context) Mode()    string { return C.mode }

func newContext(dir string, model, accused Code, ev Evaluator) *context {
	C := new(context)
	C.dir = dir
	C.limits = ev.Limits
	C.lang = map[string]string {
		"model":   model.Lang,
		"accused": accused.Lang,
	}
	C.code = map[string]string {
		"model":   model.Text,
		"accused": accused.Text,
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
   cmd = exec.Command(os.Getenv("HOME") + "/grz-jail", args...)
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

// Evaluator //////////////////////////////////////////////////

var BaseDir string
var KeepFiles bool

func init() {
   BaseDir   = os.Getenv("HOME")
	KeepFiles = false
}

func (E Evaluator) Evaluate(P *eval.Problem, Solution string) eval.Veredict {
	// FIXME: get lang from string
	C, err := E.prepareContext(P, Code{Text: Solution, Lang:"c++"}) 
	if err != nil {
		return eval.Veredict{Message: fmt.Sprintf("%s\n", err)}
	}
	results := make([]TestResult, len(E.Tests))
	for i, dbobj := range E.Tests {
		tester := dbobj.Obj.(Tester)
		E.runTest(C, tester, &results[i])
	}
	if !KeepFiles {
		C.Destroy()
	}
	return eval.Veredict{Message: "Accept", Details: db.Obj{results}}
}

func (E Evaluator) prepareContext(P *eval.Problem, accused Code) (C *context, err error) {
	id  := hash(accused.Text)
	// FIXME: Get Lang from the string itself
	model := Code{Text: P.Solution, Lang: "c++"}
	C = newContext(BaseDir + "/" + id, model, accused, E)
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

func (E Evaluator) runTest(C *context, T Tester, R *TestResult) (err error) {
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
	T.Prepare(C)
	if ! runtest("model")   { return }
	if ! runtest("accused") { return }
	*R = T.Veredict(C)
	return nil
}

func getExitStatus(err error) int {
	exiterror, ok := err.(*exec.ExitError)
	if ! ok { 
		log.Printf("Cannot get ProcessState") 
		log.Fatalf("Error was: %s\n", err)
	}
	status, ok := exiterror.Sys().(syscall.WaitStatus)
	if ! ok { 
		log.Fatalf("Cannot get syscall.WaitStatus") 
	}
	return status.ExitStatus()
}

// ReadFrom reads an evaluator from a directory. It reads a text file
// with name 'solution.*', with extension depending on the programming
// language. Then reads all files 'test.N.<type>', where N is an integer
// using a polymorphic method 'ReadFrom' for each tester.
//
func (E *Evaluator) ReadFrom(dir string, prob *eval.Problem) error {
	// Read solution
	matches, err := filepath.Glob(dir + "/solution.*")
	if err != nil {
		return fmt.Errorf("Cannot look for 'solution.*': %s\n")
	}
	
	var sol string
	for i, m := range matches {
		if i == 0 {
			sol = m
		} else {
			fmt.Fprintf(os.Stderr, "Warning: ignoring solution '%s'\n", m)
		}
	}

	// TODO: Handle more programming languages
	solstr, err := ioutil.ReadFile(sol)
	if err != nil {
		return fmt.Errorf("Cannot read file '%s': %s\n", sol, err)
	}
	// TODO: Put extension in the first line:
	//   prob.Solution = fmt.Sprintf("c++\n%s", solstr)
	prob.Solution = string(solstr)
	
	// path/filepath.glob: "New matches are added in 
	//   lexicographical order" (we use that for now)
	matches, err = filepath.Glob(dir + "/test.*.*")
	if err != nil {
		return fmt.Errorf("Cannot look for 'test.*': %s\n")
	}
	E.Tests = []db.Obj{}
	for _, m := range matches {
		typ := getType(m)
		obj := db.ObjFromType("prog.test." + typ)
		tester, ok := obj.(Readable)
		if ! ok {
			return fmt.Errorf("Type '%s' is not a programming.Tester", typ)
		}
		if err := tester.ReadFrom(m); err != nil {
			return fmt.Errorf("Couldn't read test '%s': %s\n", m, err)
		}
		E.Tests = append(E.Tests, db.Obj{tester})
	}
	return nil
}

func getType(path string) string {
	i := strings.LastIndex(path, ".")
	if i == -1 {
		panic("Assumed I would find '.' in path")
	}
	return path[i+1:]
}
