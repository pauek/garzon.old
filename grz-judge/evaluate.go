package main

import (
	"fmt"
	"log"
	"os/exec"
	"time"
	"bytes"
	"strings"
	"net/rpc"

	"garzon/eval"
)

const remotePort = 60000

type Evaluator struct {
	user     string
	host     string
	location map[string]string
	port     int
	cmd     *exec.Cmd
	client  *rpc.Client
	stderr   bytes.Buffer
}

func (E *Evaluator) init() {
	E.location = make(map[string]string)
}

func (E *Evaluator) Local() bool {
	return E.user == "" && E.host == "local"
}

func (E *Evaluator) userhost() string {
	return fmt.Sprintf("%s@%s", E.user, E.host)
}

func (E *Evaluator) findLocation(cmd string) {
	var b bytes.Buffer
	C := exec.Command("which", cmd)
	C.Stdout = &b
	if err := C.Run(); err != nil {
		log.Fatalf("Couldn't not determine where '%s' is: %s\n", cmd, err)
	}
	path := strings.Split(b.String(), "\n")[0]
	log.Printf("'%s' is '%s'\n", cmd, path)
	E.location[cmd] = path
}

var commands = []string{"grz-jail", "grz-eval"}

func (E *Evaluator) FindLocations() {
	E.location = make(map[string]string)
	for _, c := range commands { 
		E.findLocation(c) 
	}
}

func (E *Evaluator) CopyFiles() {
	for _, c := range commands { 
		E.CopyToRemote(E.location[c], ".") 
	}
}

func (E *Evaluator) CopyToRemote(path, remPath string) {
	log.Printf("Copying '%s' to '%s:%s' host\n", path, E.userhost(), remPath)
	cmd := exec.Command("scp", path, E.userhost() + ":" + remPath)
	if err := cmd.Run(); err != nil {
		log.Fatalf("Couldn't copy '%s': %s\n", path, err)
	}
}

func (E *Evaluator) StartRemoteProcess() {
	log.Printf("Executing 'grz-eval' at '%s'\n", E.userhost())
	debugFlag := ""
	if debugMode { debugFlag = "-k" }
	cmdline := fmt.Sprintf("./grz-eval %s -p %d", debugFlag, remotePort)
	redir   := fmt.Sprintf("localhost:%d:localhost:%d", E.port, remotePort)
	E.cmd = exec.Command("ssh", "-L", redir, E.userhost(), cmdline)
	E.cmd.Stderr = &E.stderr
	if err := E.cmd.Start(); err != nil {
		log.Fatalf("Couldn't run \"%s\" on account '%s': %s\n", cmdline, E.userhost(), err)
	}
}

func (E *Evaluator) KillRemoteProcess() {
	exec.Command("ssh", E.userhost(), "pkill -9 grz-eval").Run()
}

func (E *Evaluator) StartLocalProcess() {
	log.Printf("Executing 'grz-eval' locally\n")
	grzjail := E.location["grz-jail"]
	port := fmt.Sprintf("%d", E.port)
	if debugMode { 
		E.cmd = exec.Command("grz-eval", "-p", port, "-k", "-j", grzjail)
	} else {
		E.cmd = exec.Command("grz-eval", "-p", port, "-j", grzjail)
	}
	E.cmd.Stderr = &E.stderr
	if err := E.cmd.Start(); err != nil {
		log.Fatalf("Couldn't run 'grz-eval' locally: %s\n", err)
	}
}

func (E *Evaluator) ConnectRPC() {
	log.Printf("Dialing RPC...\n")
	var err error
	E.client, err = rpc.DialHTTP("tcp", fmt.Sprintf("localhost:%d", E.port))
	if err != nil {
		log.Printf("\n\n%s\n\n", E.stderr.String())
		log.Fatalf("Error dialing: %s\n", err)
	}
	log.Printf("Connected.\n")
}

func (E *Evaluator) HandleSubmissions() {
	for {
		id, ok := <- Queue.Channel
		if ! ok { break }
		Queue.SetStatus(id, "In Process")
		sub := Queue.Get(id)
		var V eval.Veredict
		log.Printf("Submitting '%s'", sub.Problem.Title)
		err := E.client.Call("Eval.Submit", sub, &V)
		if err != nil {
			log.Printf("\n\n%s\n\n", E.stderr.String())
			log.Fatalf("Call failed: %s\n", err)
		}
		sub.Status = "Resolved"
		sub.Resolved = time.Now()
		sub.Veredict = V
		sub.Problem = nil
		if err = submissions.Put(id, &sub); err != nil {
			log.Fatalf("Cannot save submission in database: %s\n", err)
		}
		log.Printf("Saved submission '%s'\n", id)
		fmt.Printf("\nREMOTE:\n%s\nLOCAL:\n", E.stderr.String())
		Queue.Delete(id)
	}
}

func (E *Evaluator) CleanUp() {
	E.client.Close()
	if E.cmd != nil {
		E.cmd.Process.Kill()
	}
	if ! E.Local() {
		E.KillRemoteProcess()
	}
}

func (E *Evaluator) Run(done chan bool) {
	E.FindLocations()
	if ! E.Local() {
		E.KillRemoteProcess()
		if copyFiles { E.CopyFiles() }
		E.StartRemoteProcess()
		time.Sleep(3 * time.Second) // FIXME
	} else {
		E.StartLocalProcess()
		time.Sleep(500 * time.Millisecond) // FIXME
	}
	E.ConnectRPC()
	E.HandleSubmissions()
	E.CleanUp()
	done <- true
}
