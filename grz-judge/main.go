
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
	"garzon/db/problems"
	"garzon/eval"
	prog "garzon/eval/programming"
)

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

func (S *Submissions) Add(Problem *eval.Problem, Solution string) (ID string) {
	ID = db.NewUUID()
	S.Mutex.Lock()
	S.inprogress[ID] = &eval.Submission{
		Problem:  Problem,
		Solution: Solution,
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

var Queue Submissions

func init() {
	prog.Register()
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
		evaluators[i] = &Evaluator{user: user, host: host, port: 50000 + i}
		go evaluators[i].Run(done)
	}
}

func waitForEvaluators() {
	for i := 0; i < len(evaluators); i++ {
		<- done
	}
}

func submit(w http.ResponseWriter, req *http.Request) {
	log.Printf("New submission: %s\n", req.FormValue("ID"))
	if req.Method != "POST" {
		fmt.Fprintf(w, "ERROR: Wrong method")
		return
	}
	if Queue.Pending() > MaxQueueSize {
		fmt.Fprint(w, "ERROR: Server too busy")
		return
	}
	id := req.FormValue("ID")
	problem, _ := problems.Get(id)
	if problem == nil {
		fmt.Fprint(w, "ERROR: Problem not found")
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
	ID := Queue.Add(problem, string(solution))
	fmt.Fprintf(w, "%s\n", ID)
	return
}

const vfmt = `
Submission ID:  %s
Problem Title:  %s
Time submitted: %s
Time resolved:  %s
Veredict:       %s
`

func veredict(w http.ResponseWriter, req *http.Request) {
	id := req.URL.Path[len("/veredict/"):]
	sub := Queue.Get(id)
	fmt.Fprintf(w, vfmt[1:], 
		id, sub.Problem.Title, sub.Submitted, sub.Resolved, sub.Veredict)
}


var debugMode, copyFiles bool

func main() {
	flag.BoolVar(&copyFiles, "copy",  false, "Copy files to remote accounts")
	flag.BoolVar(&debugMode, "debug", false, "Enable debug mode")
	flag.Parse()
	accounts := flag.Args()

	Queue.Init()
	launchEvaluators(accounts)
	http.HandleFunc("/submit/",   submit)
	http.HandleFunc("/veredict/", veredict)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Printf("ListenAndServe: %s\n", err)
	}
	Queue.Close()
	waitForEvaluators()
}