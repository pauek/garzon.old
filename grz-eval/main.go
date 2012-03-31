
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
	-p <port>,   Port to listen on (50000)

`

func main() {
	flag.Usage = func () {
		fmt.Fprintf(os.Stderr, usage)
	}
	port := flag.Int("p", 50000, "Port")
	flag.Parse()

	rpc.HandleHTTP()
	L, err := net.Listen("tcp", fmt.Sprintf(":%d", *port)); 
	if err != nil {
		log.Fatal("Listen error:", err)
	}
	log.Printf("grz-eval: starting server\n")
	http.Serve(L, nil)
}
