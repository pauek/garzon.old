
package main

import (
	"fmt"
	"log"
	"flag"
	"sync"
	"time"
	"strings"
	"net/http"
	"io/ioutil"

	"garzon/db"
	"garzon/eval"
	prog "garzon/eval/programming"
)

const ListenPort = 50000
const MaxQueueSize = 50

type Submissions struct {
	Channel chan string
	Mutex   sync.Mutex
	inprogress map[string]*eval.Submission
}

func (S *Submissions) Init() {
	S.inprogress = make(map[string]*eval.Submission)
	S.Channel    = make(chan string, MaxQueueSize)
}

func (S *Submissions) Close() { 
	close(S.Channel) 
}

func (S *Submissions) Pending() int {
	return len(S.inprogress)
}

func (S *Submissions) Add(ProblemID string, Problem *eval.Problem, Solution string) (ID string) {
	ID = db.NewUUID()
	S.Mutex.Lock()
	S.inprogress[ID] = &eval.Submission{
	   ProblemID: ProblemID, 
		Problem:   Problem,
		Solution:  Solution,
		Submitted: time.Now(),
		State: "In Queue",
	}
	S.Mutex.Unlock()
	S.Channel <- ID
	return
}

func (S *Submissions) Get(id string) (sub *eval.Submission) {
	sub, ok := S.inprogress[id]
	if ! ok { sub = nil }
	return
}

func (S *Submissions) Delete(id string) {
	delete(S.inprogress, id)
}

var Queue Submissions
var problems *db.Database
var submissions *db.Database

func init() {
	prog.Register()
	problems    = db.Problems()
	submissions = db.Submissions()
}

func parseUserHost(userhost string) (user, host string) {
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
		<- done
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
		fmt.Fprint(w, "ERROR: Cannot get problem")
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
	fmt.Fprintf(w, "%s\n", ID)
	return
}

func status(w http.ResponseWriter, req *http.Request) {
	id := req.URL.Path[len("/status/"):]
	sub := Queue.Get(id)
	if sub != nil {
		fmt.Fprintf(w, "%s\n", sub.State)
		return
	}
	fmt.Printf("Status for ID: %s\n", id)
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

var debugMode, copyFiles bool

func main() {
	flag.BoolVar(&copyFiles, "copy",  false, "Copy files to remote accounts")
	flag.BoolVar(&debugMode, "debug", false, "Enable debug mode")
	flag.Parse()
	accounts := flag.Args()

	Queue.Init()
	launchEvaluators(accounts)
	http.HandleFunc("/submit/", submit)
	http.HandleFunc("/status/", status)
	err := http.ListenAndServe(fmt.Sprintf(":%d", ListenPort), nil)
	if err != nil {
		log.Printf("ListenAndServe: %s\n", err)
	}
	Queue.Close()
	waitForEvaluators()
}