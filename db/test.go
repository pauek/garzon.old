
package main

import (
	"io/ioutil"
	"net/http"
	"log"
	"fmt"
	"encoding/json"

	"garzon/eval"
	"garzon/eval/program"
)

const CouchURL = "http://localhost:5984"

func main() {
	P := &eval.Problem{
	ID: "Cpp.Intro.SumaEnteros",
	Title: "Suma de Enteros",
	Solution: "bla blah",
	Tests: []eval.Tester{ &program.InputTester{ Input: "1 2\n" } },
	}
	data, _ := json.MarshalIndent(P, "", "  ")
	fmt.Printf("%s\n", data)
}

func main2() {
	resp, err := http.Get(CouchURL + "/problems/Cpp.Intro.SumaEnteros")
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Couldn't read all: %v\n", err)
	}
	fmt.Printf("Doc: %s\n", body)
}