package main

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"
	"code.google.com/p/go.net/websocket"

	"github.com/pauek/garzon/eval"
)

const remotePort = 60000

type Evaluator struct {
	user     string
	host     string
	location map[string]string
	port     int
	cmd      *exec.Cmd
	ws       *websocket.Conn
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
	cmd := exec.Command("scp", path, E.userhost()+":"+remPath)
	if err := cmd.Run(); err != nil {
		log.Fatalf("Couldn't copy '%s': %s\n", path, err)
	}
}

func (E *Evaluator) StartRemoteProcess() {
	log.Printf("Executing 'grz-eval' at '%s'\n", E.userhost())
	debugFlag := ""
	if Mode["debug"] {
		debugFlag = "-k"
	}
	cmdline := fmt.Sprintf("./grz-eval %s -p %d", debugFlag, remotePort)
	redir := fmt.Sprintf("localhost:%d:localhost:%d", E.port, remotePort)
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
	args := []string{"-p", port, "-j", grzjail}
	if Mode["debug"] {
		args = append(args, "-k")
	}
	if Mode["local"] {
		args = append(args, "-t") // use '/tmp'
	}
	E.cmd = exec.Command("grz-eval", args...)
	E.cmd.Stderr = &E.stderr
	if err := E.cmd.Start(); err != nil {
		log.Fatalf("Couldn't run 'grz-eval' locally: %s\n", err)
	}
}

func (E *Evaluator) ConnectWebSocket() {
	log.Printf("Dialing WebSocket...\n")
	var err error
	orig := "http://localhost/"
	url := fmt.Sprintf("ws://localhost:%d/ws", E.port)
	E.ws, err = websocket.Dial(url, "", orig)
	if err != nil {
		log.Printf("\n\n%s\n\n", E.stderr.String())
		log.Fatalf("Error dialing: %s\n", err)
	}
	log.Printf("Connected.\n")
}

func (E *Evaluator) HandleSubmissions() {
	for {
		id, ok := <-Queue.Channel
		if !ok {
			break
		}
		Queue.SetStatus(id, "In Process")
		sub := Queue.Get(id)
		log.Printf("Submitting '%s'", sub.Problem.Title)
		log.Printf("Problem: %v", sub.Problem)
		if err := websocket.JSON.Send(E.ws, sub); err != nil {
			log.Printf("\n\n%s\n\n", E.stderr.String())
			log.Fatalf("Call failed: %s\n", err)
			continue
		} 
		var err error
		var resp eval.Response
		for {
			if err = websocket.JSON.Receive(E.ws, &resp); err != nil {
				break
			}
			if resp.Status == "Resolved" {
				break
			}
			// TODO: Handle updates...
		}
		if err != nil {
			log.Printf("Error: %s %+v", err, resp)
			fmt.Printf("\nREMOTE:\n%s\nLOCAL:\n", E.stderr.String())
			E.stderr.Reset() // FIXME: se corta o algo
			continue
		}
		sub.Status = "Resolved"
		sub.Resolved = time.Now()
		sub.Veredict = *resp.Veredict
		sub.Problem = nil
		fmt.Printf("\nREMOTE:\n%s\nLOCAL:\n", E.stderr.String())
		E.stderr.Reset() // FIXME: se corta o algo
	}
}

func (E *Evaluator) CleanUp() {
	E.ws.Close()
	if E.cmd != nil {
		E.cmd.Process.Kill()
	}
	if !E.Local() {
		E.KillRemoteProcess()
	}
}

func (E *Evaluator) Run(done chan bool) {
	E.FindLocations()
	if !E.Local() {
		E.KillRemoteProcess()
		if Mode["copy"] {
			E.CopyFiles()
		}
		E.StartRemoteProcess()
		time.Sleep(3 * time.Second) // FIXME
	} else {
		E.StartLocalProcess()
		time.Sleep(500 * time.Millisecond) // FIXME
	}
	E.ConnectWebSocket()
	E.HandleSubmissions()
	E.CleanUp()
	done <- true
}

// Evaluator pool

var done chan bool
var evaluators []*Evaluator

func launchEvaluators(accounts []string) {
	N := len(accounts)
	if N < 1 {
		log.Fatal("Accounts must be 'user@host' (and >= 1)")
	}
	evaluators = make([]*Evaluator, N)
	done = make(chan bool)
	for i, acc := range accounts {
		user, host := parseUserHost(acc)
		evaluators[i] = &Evaluator{
			user: user,
			host: host,
			port: ListenPort + 1 + i,
		}
		go evaluators[i].Run(done)
	}
}

func waitForEvaluators() {
	for i := 0; i < len(evaluators); i++ {
		<-done
	}
}
