package programming

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/pauek/garzon/db"
	"github.com/pauek/garzon/eval"
	"github.com/pauek/garzon/eval/programming/lang"
)

// context /////////////////////////////////////////////////

type context struct {
	dir    string // working directory
	mode   string // current program: "model" or "accused"
	limits Constraints
	lang   map[string]string
	code   map[string]string
	stderr string // the stderr written by grz-jail

	State interface{}
}

func (C *context) Dir() string     { return C.dir }
func (C *context) ExecDir() string { return C.dir + "/eval" }
func (C *context) Mode() string    { return C.mode }

func newContext(dir string, model, accused Code, ev Evaluator) *context {
	C := new(context)
	C.dir = dir
	C.limits = ev.Limits
	C.lang = map[string]string{
		"model":   model.Lang,
		"accused": accused.Lang,
	}
	C.code = map[string]string{
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
		if err := os.Mkdir(C.dir+subdir, 0700); err != nil {
			return fmt.Errorf("Couldn't make directory '%s'", C.dir+subdir)
		}
	}
	return nil
}

func (C *context) WriteAndCompile(whom string) error {
	L := lang.ByExtension(C.lang[whom])
	if L == nil {
		return fmt.Errorf("Unsupported language '%s'", C.lang[whom])
	}
	codefile := fmt.Sprintf("%s/.%s/code.%s", C.dir, whom, L.Extensions[0])
	if err := ioutil.WriteFile(codefile, []byte(C.code[whom]), 0600); err != nil {
		return fmt.Errorf("Couldn't write %s file '%s'", whom, codefile)
	}
	exefile := fmt.Sprintf("%s/.%s/exe", C.dir, whom)
	log.Printf("Compiling '%s' ('%s')", codefile, prefix(C.code[whom], 30))
	if err := L.Functions.Compile(codefile, exefile); err != nil {
		os.RemoveAll(C.dir)
		return err
	}
	return nil
}

func (C *context) SwitchTo(whom string) error {
	if whom != "model" && whom != "accused" {
		return fmt.Errorf("Program '%s' not known")
	}
	from := fmt.Sprintf("%s/.%s/exe", C.dir, whom)
	to := fmt.Sprintf("%s/eval/exe", C.dir)
	cmd := exec.Command("cp", from, to)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Couldn't copy '%s' to '%s'", from, to)
	}
	C.mode = whom
	return nil
}

func (C *context) MakeCommand() (cmd *exec.Cmd) {
	args := []string{}
	addOption := func(flag string, val int) {
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
	args = append(args, C.dir+"/eval")
	cmd = exec.Command(GrzJail, args...)
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

var (
	BaseDir   string // base working directory 
	KeepFiles bool   // keep files after evaluation (debug)
	GrzJail   string // path of grz-jail
)

func init() {
	BaseDir = os.Getenv("HOME")
	KeepFiles = false
	GrzJail = "grz-jail" // assume its in the PATH
}

func getProgram(solution string) (program Code, ok bool) {
	i := strings.Index(solution, "\n")
	if i == -1 {
		return Code{}, false
	}
	program = Code{
		Lang: solution[:i],
		Text: solution[i+1:],
	}
	return program, true
}

func (E Evaluator) Evaluate(P *eval.Problem, Solution string, progress chan<- string) eval.Veredict {
	E.progress = progress
	log.Printf("Evaluate(%+v)", P.Evaluator.Obj)

	// FIXME: get lang from string
	if progress != nil {
		progress <- "Preparing"
	}
	// determine language
	code, ok := getProgram(Solution)
	if !ok {
		return eval.Veredict{Message: "Cannot determine language"}
	}

	// prepareContext (create dirs, compile)
	C, err := E.prepareContext(P, code)
	if err != nil {
		if comperr, ok := err.(*lang.CompilationError); ok {
			return eval.Veredict{
				Message: "Compilation Error",
				Details: db.Obj{comperr.Output},
			}
		} else {
			return eval.Veredict{Message: err.Error()}
		}
	}
	results := make([]TestResult, len(E.Tests))
	ver := make(map[string]bool)
	for i, dbobj := range E.Tests {
		tester := dbobj.Obj.(Tester)
		if progress != nil {
			progress <- fmt.Sprintf("Test %d", i+1)
		}
		E.runTest(C, tester, &results[i])
		ver[results[i].Veredict] = true
	}
	if !KeepFiles {
		C.Destroy()
	}
	message := "<No message>"
	for _, m := range []string{"Execution Error", "Wrong Answer", "Accepted"} {
		if ver[m] {
			message = m
			break
		}
	}
	return eval.Veredict{
		Message: message,
		Details: db.Obj{VeredictDetails{results}},
	}
}

func (E Evaluator) prepareContext(P *eval.Problem, accused Code) (C *context, err error) {
	id := hash(accused.Text)
	model, ok := getProgram(P.Solution)
	if !ok {
		return nil, fmt.Errorf("Cannot get model language")
	}
	C = newContext(BaseDir+"/"+id, model, accused, E)
	if err := C.CreateDirectory(); err != nil {
		return nil, err
	}
	if err := C.WriteAndCompile("model"); err != nil {
		switch err.(type) {
		case *lang.CompilationError:
			return nil, fmt.Errorf("Model doesn't compile!")
		default:
			return nil, err
		}
	}
	if err := C.WriteAndCompile("accused"); err != nil {
		return nil, err
	}
	return C, nil
}

func (E Evaluator) runTest(C *context, T Tester, R *TestResult) (err error) {
	runtest := func(whom string) bool {
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
				lines := strings.Split(stderr.String(), "\n")
				R.Veredict = lines[0]
				R.Reason.Obj = &SimpleReason{lines[1]}
			} else {
				panic("Internal error")
			}
			return false
		}
		C.stderr = stderr.String()
		if err = T.CleanUp(C); err != nil {
			return false
		}
		return true
	}
	T.Prepare(C)
	if !runtest("model") {
		return
	}
	if !runtest("accused") {
		return
	}
	V := T.Veredict(C)
	*R = V
	return nil
}

func getExitStatus(err error) int {
	exiterror, ok := err.(*exec.ExitError)
	if !ok {
		log.Printf("Cannot get ProcessState")
		log.Fatalf("Error was: %s\n", err)
	}
	status, ok := exiterror.Sys().(syscall.WaitStatus)
	if !ok {
		log.Fatalf("Cannot get syscall.WaitStatus")
	}
	return status.ExitStatus()
}

// ReadFrom reads an evaluator from a directory. It reads a text file
// with name 'solution.*', with extension depending on the programming
// language. Then reads all files 'test.N.<type>', where N is an integer
// using a polymorphic method 'ReadFrom' for each tester.
//
func (E *Evaluator) ReadDir(dir string, prob *eval.Problem) error {
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

	ext := filepath.Ext(sol)
	solstr, err := ioutil.ReadFile(sol)
	if err != nil {
		return fmt.Errorf("Cannot read file '%s': %s\n", sol, err)
	}
	// extension in the first line (to be cut later)
	prob.Solution = fmt.Sprintf("%s\n%s", ext, solstr)

	// Read limits
	E.Limits = readLimits(dir + "/limits")

	// Read Tests
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
		tester, ok := obj.(Reader)
		if !ok {
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

func readLimits(path string) (lims Constraints) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}
	var key string
	var value int
	for _, line := range strings.Split(string(data), "\n") {
		n, _ := fmt.Sscanf(line, "%s %d", &key, &value)
		if n == 2 {
			switch key {
			case "Memory": lims.Memory = value
			case "Time": lims.Time = value
			case "FileSize": lims.FileSize = value
			}
		}
	}
	return
}
