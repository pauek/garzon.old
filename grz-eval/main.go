
package main

import (
	"os"
	"fmt"
	"flag"
	"log"
	"net"
	"net/rpc"
	"net/http"

	"garzon/eval"
	prog "garzon/eval/programming"
)

func init() {
	rpc.Register(new(eval.Eval))
	prog.Register()
}

const usage = `usage: grz-eval [options...]

Options:
	-p <port>,   Port to listen on (60000)
   -k,          Keep Files

`

func main() {
	flag.Usage = func () {
		fmt.Fprintf(os.Stderr, usage)
	}
	port := flag.Int("p", 60000, "Port")
	keep := flag.Bool("k", false, "Keep Files")
	flag.Parse()
	prog.KeepFiles = *keep

	rpc.HandleHTTP()
	L, err := net.Listen("tcp", fmt.Sprintf(":%d", *port)); 
	if err != nil {
		log.Fatal("Listen error:", err)
	}
	log.Printf("grz-eval: starting server\n")
	http.Serve(L, nil)
}
