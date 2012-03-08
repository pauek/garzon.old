
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

func within(dir string, F func()) {
	lastdir, _ := os.Getwd() // TODO: handle error?
	os.Chdir(dir)
	defer os.Chdir(lastdir)
	F()
}

func _sha1(s string) string {
	hash := sha1.New()
	io.WriteString(hash, s)
	return fmt.Sprintf("%x", hash.Sum(nil))
}

type Evaluator struct {
	BaseDir   string
	sessions  map[string]*Session
}

type Session struct {
	dir  string
	lang *lang.Language
}

type Program struct {
	Lang, Code string
}

type Context struct {
	accused    Program
	model      Program
	pre, post *Program
}

func NewSession(dir string, Lang string) *Session {
	os.Mkdir(dir, 0700)
	return &Session{dir: dir, lang: lang.Get(Lang)}
}

func (S *Session) Write(name string, prog Program) error {
	lang, ok := lang.Languages[prog.Lang]
	if ! ok {
		return fmt.Errorf("Unsupported language '%s'", prog.Lang)
	}
	fnm := fmt.Sprintf("%s/%s.%s", S.dir, name, lang.Extension)
	ioutil.WriteFile(fnm, []byte(prog.Code), 0600)
	return nil
}

func (S *Session) Compile(infile, outfile string) (err error) {
	within(S.dir, func() {
		err = S.lang.Functions.Compile(infile, outfile)
	})
	return
}

func (S *Session) Execute(filename string, input string) (out string, err error) {
	within(S.dir, func () {
		out, err = S.lang.Functions.Execute(filename, input)
	})
	return
}

func (S *Session) Destroy() {
	cmd := exec.Command("rm", "-rf", S.dir)
	if err := cmd.Run(); err != nil {
		log.Fatal("Couldn't erase session %s", S.dir)
	}
}


func NewEvaluator(basedir string) *Evaluator {
	ev := new(Evaluator)
	ev.BaseDir = basedir
	ev.sessions = make(map[string]*Session)
	return ev
}

type SessionFunc func(S *Session) error

func (E *Evaluator) withSession(ID string, Fn SessionFunc) error {
	S, ok := E.sessions[ID]; 
	if ! ok {
		return fmt.Errorf("Session '%s' not found", ID)
	}
	return Fn(S)
}

func (E *Evaluator) Create(ctx Context, ID *string) error {
	*ID = _sha1(ctx.accused.Code)
	session := NewSession(E.BaseDir + "/" + *ID, ctx.accused.Lang)

	// Write programs
	session.Write("accused", ctx.accused)
	session.Write("model", ctx.model)
	if ctx.pre != nil {
		session.Write("pre", *ctx.pre)
	}
	if ctx.post != nil {
		session.Write("post", *ctx.post)
	}

	// Compile programs
	if err := session.Compile("model", "exe"); err != nil {
		return fmt.Errorf("The 'model' program doesn't compile")
	}
	if ctx.pre != nil {
		if err := session.Compile("pre", "pre"); err != nil {
			return fmt.Errorf("The 'pre' program doesn't compile")
		}
	}
	if ctx.post != nil {
		if err := session.Compile("post", "post"); err != nil {
			return fmt.Errorf("The 'post' program doesn't compile")
		}
	}
	return nil
}

func (E *Evaluator) Compile(ID string) error {
	return E.withSession(ID, func (S *Session) error {
		err := S.Compile("accused", "exe")
		if err != nil {
			return fmt.Errorf("Compilation error:\n%s", err)
		}
		return nil
	})
}
	
type Request struct {
	ID, Input string
}

func (E *Evaluator) Execute(req Request, output *string) error {
	return E.withSession(req.ID, func (S *Session) error {
		out, err := S.Execute("accused", req.Input); 
		if err != nil {
			return err
		}
		*output = out
		return nil
	})
}

func (E *Evaluator) Delete(ID string) error {
	return E.withSession(ID, func (S *Session) error {
		S.Destroy()
		return nil
	})
}
