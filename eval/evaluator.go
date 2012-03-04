
package eval

import (
	"os"
	"os/exec"
	"io"
	"io/ioutil"
	"fmt"
	"crypto/sha1"
	"log"
	
	"garzon/eval/lang"
)

type Session struct {
	basedir, ID string
	lang lang.Language
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
		out, err = S.lang.Functions.Execute(S.ID, input)
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
	Lang, Code string
}

func NewEvaluator(basedir string) *Evaluator {
	ev := new(Evaluator)
	ev.BaseDir = basedir
	ev.sessions = make(map[string]*Session)
	return ev
}

func (E *Evaluator) Compile(prog Program, ID *string) error {
	lang, ok := lang.Languages[prog.Lang]
	if ! ok {
		return fmt.Errorf("Unsupported language '%s'", prog.Lang)
	}

	hash := sha1.New()
	io.WriteString(hash, prog.Code)
	*ID = fmt.Sprintf("%x", hash.Sum(nil))
	os.Mkdir(E.BaseDir + "/" + *ID, 0700)
	filename := E.BaseDir + "/" + *ID + "/code." + lang.Extension
	ioutil.WriteFile(filename, []byte(prog.Code), 0600)	
	session := &Session{ 
	   basedir: E.BaseDir, 
	   ID: *ID, 
	   lang: lang,
   }
	var output string
	var err error
	session.withinDir(func () {
		output, err = lang.Functions.Compile(*ID)
	})
	if err != nil {
		return fmt.Errorf("Compilation error:\n%s", output)
	}
	E.sessions[*ID] = session
	return nil
}

type Request struct {
	ID, Input string
}

func (E *Evaluator) Execute(req Request, output *string) error {
	S, ok := E.sessions[req.ID]; 
	if ! ok {
		return fmt.Errorf("Session '%s' not found", req.ID)
	}
	out, err := S.Execute(req.Input); 
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
