package main

import (
	"flag"
	"fmt"
	"github.com/pauek/garzon/db"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

const (
	ListenPort   = 50000
	MaxQueueSize = 50
)

var (
	copyFiles bool
	debugMode bool
	localMode bool
	openMode  bool

	problems *db.Database
	users    *db.Database
)

const usage = `usage: grz-judge [options...] [accounts...]

Options:
	--copy,    Copy 'grz-{eval,jail}' to remote accounts
   --debug,   Tell 'grz-eval' to keep evaluation directories
   --local,   Local mode (read problems from files, don't touch DB)
   --open,    Open mode (anyone can submit, submissions not stored)
					
`

func login(w http.ResponseWriter, req *http.Request) {
	if localMode || openMode {
		fmt.Fprintf(w, "Ok\n")
		return
	}
	login := req.FormValue("login")
	passwd := req.FormValue("passwd")
	if !LoginCorrect(login, passwd) {
		log.Printf("Unauthorized '%s'", login)
		http.Error(w, "Unauthorized", 401)
		return
	}
	tok, err := CreateToken(login)
	if err != nil {
		http.Error(w, fmt.Sprintf("Internal Error: %s\n", err), 500)
		return
	}
	http.SetCookie(w, &http.Cookie{Name: "Auth", Value: tok, MaxAge: 600})
	fmt.Fprintf(w, "Ok\n")
	log.Printf("Authorized '%s'", login)
}

func logout(w http.ResponseWriter, req *http.Request) {
	if localMode || openMode {
		fmt.Fprintf(w, "Ok\n")
		return
	}
	login := req.Header.Get("user")
	log.Printf("Logged out '%s'", login)
	DeleteToken(login)
	fmt.Fprintf(w, "Ok\n")
}

func wAuth(fn http.HandlerFunc) http.HandlerFunc {
	return func (w http.ResponseWriter, req *http.Request) {
		ok, user := IsAuthorized(req); 
		if !ok {
			http.Error(w, "Unauthorized", 401)
			return
		}
		req.Header.Add("user", user) // ugly
		fn(w, req)
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
	user := req.Header.Get("user")
	ID := Queue.Add(user, id, problem, string(solution))
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
	switch sub := Queue.Get(id); {
	case sub != nil && sub.Status != "Resolved":
		fmt.Fprintf(w, "Not resolved\n")
	case sub != nil:
		fmt.Fprintf(w, "%s\n", sub.Veredict.Message)
		if sub.Veredict.Message != "Accepted" {
			fmt.Fprintf(w, "\n%s\n", sub.Veredict.Details.Obj)
		}
		Queue.Delete(id)
	default:
		fmt.Fprint(w, "Not found\n")
	}
}

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
		problems = getDB("problems")
		users = getDB("users")
	}

	Queue.Init()
	launchEvaluators(accounts)
	http.HandleFunc("/login", login)
	http.HandleFunc("/logout", wAuth(logout))
	http.HandleFunc("/submit", wAuth(submit))
	http.HandleFunc("/status/", wAuth(status))
	http.HandleFunc("/veredict/", wAuth(veredict))
	err := http.ListenAndServe(fmt.Sprintf(":%d", ListenPort), nil)
	if err != nil {
		log.Printf("ListenAndServe: %s\n", err)
	}
	Queue.Close()
	waitForEvaluators()
}
