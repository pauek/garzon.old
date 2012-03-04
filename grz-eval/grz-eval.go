
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
)

func main() {
	dir := flag.String("basedir", os.Getenv("HOME"), "Base directory")
	port := flag.Int("port", 15001, "Port")
	flag.Parse()
	rpc.Register(eval.NewEvaluator(*dir))
	rpc.HandleHTTP()
	L, err := net.Listen("tcp", fmt.Sprintf(":%d", *port)); 
	if err != nil {
		log.Fatal("Listen error:", err)
	}
	http.Serve(L, nil)
}
