
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
	Compile(ID string) (string, error)
	Execute(ID string, input string) (string, error)
}

type Cpp string

func (L *Cpp) Name() string {
	return string(*L)
}

func (L *Cpp) Compile(ID string) (string, error) {
	cmd := exec.Command("g++", "-o", "exe", "code")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return out.String(), fmt.Errorf("Compilation failed")
	}
	return "", nil
}

func (L *Cpp) Execute(ID string, input string) (string, error) {
	cmd := exec.Command("exe")
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
	languages map[string]Language
	sessions  map[string]*Session
}

type Program struct {
	lang, code string
}

func (E *Evaluator) Compile(prog Program, ID *string) error {
	hash := sha1.New()
	io.WriteString(hash, prog.code)
	*ID = string(hash.Sum(nil))
	os.Mkdir(E.BaseDir + "/" + *ID, 0700)
	ioutil.WriteFile(E.BaseDir + "/" + *ID + "/code", []byte(prog.code), 0600)	
	lang, ok := E.languages[prog.lang]
	if ! ok {
		return fmt.Errorf("Unsupported language '%s'", prog.lang)
	}
	output, err := lang.Compile(*ID)
	if err != nil {
		return fmt.Errorf("Compilation error:\n%s", output)
	}
	E.sessions[*ID] = &Session{ 
	   basedir: E.BaseDir, 
	   ID: *ID, 
	   lang: lang,
   }
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
