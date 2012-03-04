
package eval

import (
	"os"
	"os/exec"
	"io"
	"io/ioutil"
	"fmt"
	"bytes"
	"crypto/sha1"
	"strings"
	"log"
)

type Language interface {
	Name() string
	Extension() string
	Compile(ID string) (string, error)
	Execute(ID string, input string) (string, error)
}

type Cpp string

func (L *Cpp) Name() string {
	return string(*L)
}

func (L *Cpp) Extension() string {
	return "cc"
}

func (L *Cpp) Compile(ID string) (string, error) {
	cmd := exec.Command("g++", "-o", "exe", "code.cc")
	var out bytes.Buffer
	cmd.Stderr = &out
	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
		return out.String(), fmt.Errorf("Compilation failed")
	}
	return "", nil
}

var languages map[string]Language

func init() {
	languages = make(map[string]Language)
	languages["c++"] = new(Cpp)
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

type Session struct {
	basedir, ID string
	lang Language
}

func (S *Session) Dir() string {
	return S.basedir + "/" + S.ID
}

func (S *Session) withinDir(F func()) {
	os.Chdir(S.Dir())
	defer os.Chdir(S.basedir)
	F()
}

func (S *Session) Execute(input string) (out string, err error) {
	S.withinDir(func () {
		out, err = S.lang.Execute(S.ID, input)
	})
	return
}

func (S *Session) Destroy() {
	cmd := exec.Command("rm", "-rf", S.Dir())
	if err := cmd.Run(); err != nil {
		log.Fatal("Couldn't erase session with ID: %s", S.ID)
	}
}

type Evaluator struct {
	BaseDir   string
	sessions  map[string]*Session
}

type Program struct {
	lang, code string
}

func NewEvaluator(basedir string) *Evaluator {
	ev := new(Evaluator)
	ev.BaseDir = basedir
	ev.sessions = make(map[string]*Session)
	return ev
}

func (E *Evaluator) Compile(prog Program, ID *string) error {
	lang, ok := languages[prog.lang]
	if ! ok {
		return fmt.Errorf("Unsupported language '%s'", prog.lang)
	}

	hash := sha1.New()
	io.WriteString(hash, prog.code)
	*ID = fmt.Sprintf("%x", hash.Sum(nil))
	os.Mkdir(E.BaseDir + "/" + *ID, 0700)
	filename := E.BaseDir + "/" + *ID + "/code." + lang.Extension()
	ioutil.WriteFile(filename, []byte(prog.code), 0600)	
	session := &Session{ 
	   basedir: E.BaseDir, 
	   ID: *ID, 
	   lang: lang,
   }
	var output string
	var err error
	session.withinDir(func () {
		output, err = lang.Compile(*ID)
	})
	if err != nil {
		return fmt.Errorf("Compilation error:\n%s", output)
	}
	E.sessions[*ID] = session
	return nil
}

type Request struct {
	ID, input string
}

func (E *Evaluator) Execute(req Request, output *string) error {
	S, ok := E.sessions[req.ID]; 
	if ! ok {
		return fmt.Errorf("Session '%s' not found", req.ID)
	}
	out, err := S.Execute(req.input); 
	if err != nil {
		return err
	}
	*output = out
	return nil
}

func (E *Evaluator) Delete(ID string, result *bool) error {
	S, ok := E.sessions[ID]; 
	if ! ok {
		*result = false
		return fmt.Errorf("Session '%s' not found", ID)
	}
	S.Destroy()
	*result = true
	return nil
}
