package main

import (
	"flag"
	"fmt"
"strings"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"code.google.com/p/go.net/websocket"

	"github.com/pauek/garzon/eval"
	prog "github.com/pauek/garzon/eval/programming"
)

const usage = `usage: grz-eval [options...]

Options:
	-p <port>,   Port to listen on (60000)
	-j <path>,   Location of 'grz-jail'
	-t,          Use temp directory
   -k,          Keep Files

`

func submissions(ws *websocket.Conn) {
	for {
		var sub eval.Submission
		err := websocket.JSON.Receive(ws, &sub)
		if err != nil {
			log.Printf("websocket.JSON.Receive error: %s", err)
			err := websocket.JSON.Send(ws, eval.Response{Status: "Error"})
			if err != nil {
				log.Printf("websocket.JSON.Send 'Error' error: %s", err)
			}
		}
		var V eval.Veredict
		progress := make(chan string)
		go eval.Submit(sub, &V, progress)
		for {
			msg := <- progress
			if msg == "Resolved" || strings.HasPrefix(msg, "Error") {
				break
			}
			err = websocket.JSON.Send(ws, eval.Response{Status: msg, Veredict: nil})
			if err != nil {
				log.Printf("websocket.JSON.Send '%s' error: %s", msg, err)
			}
		}
		err = websocket.JSON.Send(ws, eval.Response{Status: "Resolved", Veredict: &V})
		if err != nil {
			log.Printf("websocket.JSON.Send 'Resolved' error: %s", err)
		}
	}
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, usage)
	}
	port := flag.Int("p", 60000, "Port")
	keep := flag.Bool("k", false, "Keep Files")
	temp := flag.Bool("t", false, "Temp directory")
	grzjail := flag.String("j", "grz-jail", "Location of grz-jail")
	flag.Parse()

	prog.KeepFiles = *keep
	prog.GrzJail = *grzjail
	if *temp {
		tmpdir := filepath.Join(os.TempDir(), "grz-eval")
		_ = os.RemoveAll(tmpdir)
		if err := os.Mkdir(tmpdir, 0700); err != nil {
			log.Fatal("Couldn't make directory '%s'\n", tmpdir)
		}
		prog.BaseDir = tmpdir
	}
	log.Printf("grz-eval: starting server\n")
	http.Handle("/ws", websocket.Handler(submissions))
	err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
	if err != nil {
		log.Fatal("Listen error:", err)
	}
}
