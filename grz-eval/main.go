package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"path/filepath"

	"garzon/eval"
	prog "garzon/eval/programming"
)

func init() {
	rpc.Register(new(eval.Eval))
}

const usage = `usage: grz-eval [options...]

Options:
	-p <port>,   Port to listen on (60000)
	-j <path>,   Location of 'grz-jail'
	-t,          Use temp directory
   -k,          Keep Files

`

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

	rpc.HandleHTTP()
	L, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatal("Listen error:", err)
	}
	log.Printf("grz-eval: starting server\n")
	http.Serve(L, nil)
}
