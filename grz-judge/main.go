package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"garzon/db"
	"garzon/eval"
	_ "garzon/eval/programming"
)

const ListenPort = 50000
const MaxQueueSize = 50

var Queue Submissions
var problems *db.Database
var submissions *db.Database

func init() {
	var err error
	problems, err = db.GetDB("problems")
	if err != nil {
		log.Fatalf("Cannot get database 'problems': %s\n", err)
	}
	submissions, err = db.GetOrCreateDB("submissions")
	if err != nil {
		log.Fatalf("Cannot get database 'submissions': %s\n", err)
	}
}

func parseUserHost(userhost string) (user, host string) {
	if userhost == "local" {
		return "", "local"
	}
	parts := strings.Split(userhost, "@")
	if len(parts) != 2 {
		log.Fatal("Wrong user@host = '%s'\n", userhost)
	}
	return parts[0], parts[1]
}

var done chan bool
var evaluators []*Evaluator

func launchEvaluators(accounts []string) {
	N := len(accounts)
	if N < 1 {
		log.Fatal("Accounts must be 'user@host' (and >= 1)")
	}
	evaluators = make([]*Evaluator, N)
	done = make(chan bool)
	for i, acc := range accounts {
		user, host := parseUserHost(acc)
		evaluators[i] = &Evaluator{
			user: user,
			host: host,
			port: ListenPort + 1 + i,
		}
		go evaluators[i].Run(done)
	}
}

func waitForEvaluators() {
	for i := 0; i < len(evaluators); i++ {
		<-done
	}
}

func submit(w http.ResponseWriter, req *http.Request) {
	log.Printf("New submission: %s\n", req.FormValue("id"))
	if req.Method != "POST" {
		fmt.Fprintf(w, "ERROR: Wrong method")
		return
	}
	if Queue.Pending() > MaxQueueSize {
		fmt.Fprint(w, "ERROR: Server too busy")
		return
	}
	probid := req.FormValue("id")
	var problem eval.Problem
	_, err := problems.Get(probid, &problem)
	if err != nil {
		fmt.Fprintf(w, "ERROR: Problem '%s' not found", probid)
		return
	}
	file, _, err := req.FormFile("solution")
	if err != nil {
		fmt.Fprint(w, "Cannot get solution file")
	}
	solution, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Fprint(w, "Cannot read solution file")
	}
	ID := Queue.Add(probid, &problem, string(solution))
	fmt.Fprintf(w, "%s", ID)
	return
}

func status(w http.ResponseWriter, req *http.Request) {
	id := req.URL.Path[len("/status/"):]
	sub := Queue.Get(id)
	if sub != nil {
		fmt.Fprintf(w, "%s", sub.Status)
		return
	}
	rev, err := submissions.Rev(id)
	if err != nil {
		log.Printf("Cannot query submission '%s'\n", id)
	}
	if rev != "" {
		fmt.Fprint(w, "Resolved\n")
	} else {
		fmt.Fprint(w, "Not found\n")
	}
}

func veredict(w http.ResponseWriter, req *http.Request) {
	id := req.URL.Path[len("/veredict/"):]
	var sub eval.Submission
	_, err := submissions.Get(id, &sub)
	if err != nil {
		fmt.Fprintf(w, "Cannot get submission '%s': %s\n", id, err)
		return
	}
	fmt.Fprintf(w, "%s\n", sub.Veredict.Message)
	if sub.Veredict.Message != "Accepted" {
		fmt.Fprintf(w, "\n%s\n", sub.Veredict.Details.Obj)
	}
}

var debugMode, copyFiles bool

const usage = `usage: grz-judge [options...] <accounts>*

Options:
   --copy,    Copy files to remote accounts
   --debug,   Enable debug mode

`

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, usage)
	}
	flag.BoolVar(&copyFiles, "copy", false, "Copy files to remote accounts")
	flag.BoolVar(&debugMode, "debug", false, "Enable debug mode")
	flag.Parse()
	accounts := flag.Args()

	Queue.Init()
	launchEvaluators(accounts)
	http.HandleFunc("/submit/", submit)
	http.HandleFunc("/status/", status)
	http.HandleFunc("/veredict/", veredict)
	err := http.ListenAndServe(fmt.Sprintf(":%d", ListenPort), nil)
	if err != nil {
		log.Printf("ListenAndServe: %s\n", err)
	}
	Queue.Close()
	waitForEvaluators()
}
