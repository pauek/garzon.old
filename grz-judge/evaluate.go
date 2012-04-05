
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

const remotePort = 50000

func findLocation(cmd string) (path string) {
	var b bytes.Buffer
	C := exec.Command("which", cmd)
	C.Stdout = &b
	if err := C.Run(); err != nil {
		log.Fatalf("Couldn't not determine where '%s' is: %s\n", cmd, err)
	}
	path = strings.Replace(b.String(), "\n", "", -1)
	log.Printf("'%s' is '%s'\n", cmd, path)
	return
}

func copyToRemote(path, userhost, remPath string) {
	log.Printf("Copying '%s' to '%s:%s' host\n", path, userhost, remPath)
	cmd := exec.Command("scp", path, userhost + ":" + remPath)
	if err := cmd.Run(); err != nil {
		log.Fatalf("Couldn't copy '%s': %s\n", path, err)
	}
}


type Evaluator struct {
	user, host string
	port    int
	sshcmd *exec.Cmd
	client *rpc.Client
	stderr  bytes.Buffer
}

func (E *Evaluator) userhost() string {
	return fmt.Sprintf("%s@%s", E.user, E.host)
}

func (E *Evaluator) copyFiles() {
	grzjail := findLocation("grz-jail")
	grzeval := findLocation("grz-eval")
	copyToRemote(grzjail, E.userhost(), ".")
	copyToRemote(grzeval, E.userhost(), ".")
}

func (E *Evaluator) startRemoteProcess() {
	log.Printf("Executing 'grz-eval' at '%s'\n", E.userhost())
	debugFlag := ""
	if debugMode { debugFlag = "-k" }
	cmdline := fmt.Sprintf("./grz-eval %s -p %d", debugFlag, remotePort)
	redir   := fmt.Sprintf("localhost:%d:localhost:%d", E.port, remotePort)
	E.sshcmd = exec.Command("ssh", "-L", redir, E.userhost(), cmdline)
	E.sshcmd.Stderr = &E.stderr
	if err := E.sshcmd.Start(); err != nil {
		log.Fatalf("Couldn't run \"%s\" on account '%s': %s\n", cmdline, E.userhost(), err)
	}
}

func (E *Evaluator) killRemoteProcess() {
	exec.Command("ssh", E.userhost(), "pkill -9 grz-eval").Run()
}

func (E *Evaluator) connectRPC() {
	log.Printf("Dialing RPC...\n")
	var err error
	E.client, err = rpc.DialHTTP("tcp", fmt.Sprintf("localhost:%d", E.port))
	if err != nil {
		log.Printf("\n\n%s\n\n", E.stderr.String())
		log.Fatalf("Error dialing: %s\n", err)
	}
	log.Printf("Connected.\n")
}

func (E *Evaluator) handleSubmissions() {
	for {
		id, ok := <- Queue.Channel
		if ! ok { break }
		sub := Queue.Get(id)
		var V eval.Veredict
		log.Printf("Submitting '%s'", sub.Problem.Title)
		err := E.client.Call("Eval.Submit", sub, &V)
		if err != nil {
			log.Printf("\n\n%s\n\n", E.stderr.String())
			log.Fatalf("Call failed: %s\n", err)
		}
		sub.Resolved = time.Now()
		sub.Veredict = V
		sub.Problem = nil
		if err = submissions.Put(id, &sub); err != nil {
			log.Fatalf("Cannot save submission in database: %s\n", err)
		}
		log.Printf("Saved submission '%s'\n", id)
		Queue.Delete(id)
	}
}

func (E *Evaluator) cleanUp() {
	E.client.Close()
	if E.sshcmd != nil {
		E.sshcmd.Process.Kill()
	}
	E.killRemoteProcess()
}

func (E *Evaluator) Run(done chan bool) {
	E.killRemoteProcess()
	if copyFiles { E.copyFiles() }
	E.startRemoteProcess()
	time.Sleep(3 * time.Second) // FIXME
	E.connectRPC()
	E.handleSubmissions()
	E.cleanUp()
	done <- true
}
