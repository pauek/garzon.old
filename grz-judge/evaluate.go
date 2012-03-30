
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

func evaluate(A Account) {
	var cmd *exec.Cmd
	var err error

	// 1. Find out where grz-eval is.
	var b bytes.Buffer
	cmd = exec.Command("which", "grz-eval")
	cmd.Stdout = &b
	if err = cmd.Run(); err != nil {
		log.Fatalf("Couldn't not determine where 'grz-eval' is: %s\n", err)
	}
	grzeval  := strings.Replace(b.String(), "\n", "", -1)
	userhost := fmt.Sprintf("%s@%s", A.user, A.host)
	
	// 2. Kill the process if it exists
	exec.Command("ssh", userhost, "pkill -9 grz-eval").Run()

	// 3. Copy grz-eval over.
	cmd = exec.Command("scp", grzeval, userhost + ":")
	if err = cmd.Run(); err != nil {
		log.Fatalf("Couldn't copy '%s' to account '%s': %s\n", grzeval, userhost, err)
	}
	
	// 4. Run grz-eval there + establish tunnel
	cmdline := fmt.Sprintf("./grz-eval -p %d", remotePort)
	redir   := fmt.Sprintf("localhost:%d:%s:%d", A.port, A.host, remotePort)
	cmd = exec.Command("ssh", "-L", redir, userhost, cmdline)
	var e bytes.Buffer
	cmd.Stderr = &e
	if err = cmd.Start(); err != nil {
		log.Fatalf("Couldn't run \"%s\" on account '%s': %s\n", cmdline, userhost, err)
	}
	defer cmd.Process.Kill()
	
	// hack: wait for the ssh process
	time.Sleep(time.Second)

	// Submission loop
	client, err := rpc.DialHTTP("tcp", fmt.Sprintf("localhost:%d", A.port))
	if err != nil {
		log.Fatalf("Error dialing: %s\n", err)
	}

	for {
		s := <- submissions
		var V eval.Veredict
		client.Call("Evaluator.Submit", s, &V)
		fmt.Printf("Result was %v\n", V)
	}
}