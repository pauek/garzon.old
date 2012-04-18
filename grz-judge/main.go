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

func getDBs() {
	var err error
	problems, err = db.GetDB("problems")
	if err != nil {
		log.Fatalf("Cannot get database 'problems': %s\n", err)
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

func getProblem(probid string) (P *eval.Problem, err error) {
	if !localMode {
		var problem eval.Problem
		_, err = problems.Get(probid, &problem)
		return &problem, nil
	}
	problem, err := eval.ReadFromID(probid)
	if err != nil {
		return nil, fmt.Errorf("Couldn't read problem '%s': %s\n", probid, err)
	}
	return problem, nil
}

func login(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "Unimplemented")
}

func logout(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "Unimplemented")
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
	id := req.FormValue("id")
	problem, err := getProblem(id)
	if err != nil {
		fmt.Fprintf(w, "ERROR: %s\n", err)
		return
	}
	file, _, err := req.FormFile("solution")
	if err != nil {
		fmt.Fprint(w, "ERROR: Cannot get solution file")
		return
	}
	solution, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Fprint(w, "ERROR: Cannot read solution file")
		return
	}
	ID := Queue.Add(id, problem, string(solution))
	fmt.Fprintf(w, "%s", ID)
	return
}

func status(w http.ResponseWriter, req *http.Request) {
	id := req.URL.Path[len("/status/"):]
	sub := Queue.Get(id)
	if sub != nil {
		fmt.Fprintf(w, "%s", sub.Status)
		return
	} else {
		fmt.Fprint(w, "Not found\n")
	}
}

func veredict(w http.ResponseWriter, req *http.Request) {
	id := req.URL.Path[len("/veredict/"):]
	sub := Queue.Get(id)
	if sub != nil {
		fmt.Fprintf(w, "%s\n", sub.Veredict.Message)
		if sub.Veredict.Message != "Accepted" {
			fmt.Fprintf(w, "\n%s\n", sub.Veredict.Details.Obj)
		}
	} else {
		fmt.Fprint(w, "Not found\n")
	}
}

var (
	copyFiles bool
	debugMode bool
	localMode bool
	openMode  bool
)

const usage = `usage: grz-judge [options...] [accounts...]

Options:
   --copy,    Copy files to remote accounts
   --debug,   Show debug information
   --local,   Local mode (read files, don't touch DB)
   --open,    Open mode (anyone can submit, submissions not stored)
					
`

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, usage)
	}
	flag.BoolVar(&copyFiles, "copy", false, "")
	flag.BoolVar(&debugMode, "debug", false, "")
	flag.BoolVar(&localMode, "local", false, "")
	flag.BoolVar(&openMode, "open", false, "")
	flag.Parse()

	accounts := flag.Args()
	if len(accounts) == 0 {
		accounts = []string{"local"}
	}

	if !localMode {
		getDBs()
	}

	Queue.Init()
	launchEvaluators(accounts)
	http.HandleFunc("/login/", login)
	http.HandleFunc("/logout/", logout)
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
