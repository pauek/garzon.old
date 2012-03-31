
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

type Account struct {
	user, host string
	port int
}

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

func killRemoteProcess(userhost string) {
	exec.Command("ssh", userhost, "pkill -9 grz-eval").Run()
}

func copyFiles(userhost string) {
	grzjail := findLocation("grz-jail")
	grzeval := findLocation("grz-eval")
	copyToRemote(grzjail, userhost, ".")
	copyToRemote(grzeval, userhost, ".")
}

func evaluate(A Account, done chan bool) {
	userhost := fmt.Sprintf("%s@%s", A.user, A.host)

	var e bytes.Buffer
	var cmd *exec.Cmd
	if true {
		killRemoteProcess(userhost)

		// copy files over
		if copyfiles {
			copyFiles(userhost)
		}

		// Run grz-eval there
		log.Printf("Executing 'grz-eval' at '%s'\n", userhost)
		cmdline := fmt.Sprintf("./grz-eval -p %d", remotePort)
		redir   := fmt.Sprintf("localhost:%d:localhost:%d", A.port, remotePort)
		cmd = exec.Command("ssh", "-L", redir, userhost, cmdline)
		cmd.Stderr = &e
		if err := cmd.Start(); err != nil {
			log.Fatalf("Couldn't run \"%s\" on account '%s': %s\n", cmdline, userhost, err)
		}
		
		// hack: wait for the ssh process
		time.Sleep(5 * time.Second)
	}	 

	// Submission loop
	log.Printf("Dialing RPC...\n")
	client, err := rpc.DialHTTP("tcp", fmt.Sprintf("localhost:%d", A.port))
	if err != nil {
		log.Printf("%s\n", e.String())
		log.Fatalf("Error dialing: %s\n", err)
	}
	log.Printf("Connected.\n")

	s := <- submissions
	var V eval.Veredict
	log.Printf("Submitting '%s'", s.Problem.Title)
	err = client.Call("Eval.Submit", s, &V)
	if err != nil {
		log.Printf("%s\n", e.String())
		log.Fatalf("Call failed: %s\n", err)
	}
	fmt.Printf("Result was %v\n", V)
	client.Close()
	if cmd != nil {
		cmd.Process.Kill()   // kill 'ssh'
	}
	killRemoteProcess(userhost)
	done <- true
}