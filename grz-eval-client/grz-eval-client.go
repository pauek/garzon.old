
package main

import (
	"fmt"
	"log"
	"flag"
	"net/rpc"
	"io/ioutil"
	
	"garzon/eval"
)

func main() {
	host := flag.String("host", "localhost", "Host to connect to")
	port := flag.Int("port", 15001, "Port")
	file := flag.String("sourcefile", "prog.cc", "Source program")
	flag.Parse()

	client, err := rpc.DialHTTP("tcp", fmt.Sprintf("%s:%d", *host, *port))
	if err != nil {
		log.Fatal("dialing:", err)
	}

	// Compile
	code, err := ioutil.ReadFile(*file)
	if err != nil {
		log.Fatalf("Cannot real code file '%s'", *file)
	}
	prog := eval.Program{ Lang: "c++", Code: string(code) }
	var id string
	err = client.Call("Evaluator.Compile", prog, &id)
	if err != nil {
		log.Fatal("eval.Compile error:", err)
	}
	
	// Execute
	req := eval.Request{ ID: id, Input: "2 3\n" }
	var output string
	err = client.Call("Evaluator.Execute", req, &output)
	if err != nil {
		log.Fatal("eval.Execute error:", err)
	}

	// Delete
	var ok bool
	err = client.Call("Evaluator.Delete", id, &ok)
	if ! ok {
		log.Fatal("Couldn't delete program")
	}
	if err != nil {
		log.Fatal("Couldn't delete program:", err)
	}

	fmt.Print(output)	
}
