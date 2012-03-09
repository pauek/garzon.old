
package main

import (
	"fmt"
	"flag"
	"log"
	"net"
	"net/rpc"
	"net/http"
	
	"garzon/eval"
)

func init() {
	rpc.Register(eval.ProgramEvaluator)
}

func main() {
	port := flag.Int("port", 15001, "Port")
	flag.Parse()

	rpc.HandleHTTP()
	L, err := net.Listen("tcp", fmt.Sprintf(":%d", *port)); 
	if err != nil {
		log.Fatal("Listen error:", err)
	}
	http.Serve(L, nil)
}
