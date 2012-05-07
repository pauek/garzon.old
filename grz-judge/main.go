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
	Server string
	Mode  = make(map[string]bool)
	Modes = []string{"copy", "debug", "local", "open", "files"}

	Problems *db.Database
	Users    *db.Database
)

const usage = `usage: grz-judge [options...] [accounts...]

Options:
   --copy,    Copy 'grz-{eval,jail}' to remote accounts
   --debug,   Tell 'grz-eval' to keep evaluation directories
   --local,   Local mode (only listen to localhost:50000)
   --open,    Open mode (no authentication, submissions not stored)
   --files,   Doesn't touch DB, reads problems from disk (implies --open)

Environment:
	GRZ_PATH   List of colon-separated directories with problems (for --files)
					
`

func open(w http.ResponseWriter, req *http.Request) {
	if Mode["open"] {
		fmt.Fprintf(w, "yes")
	} else {
		fmt.Fprintf(w, "no")
	}
}

func login(w http.ResponseWriter, req *http.Request) {
	if Mode["open"] {
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
	if Mode["open"] {
		fmt.Fprintf(w, "Ok\n")
		return
	}
	login := req.Header.Get("user")
	log.Printf("Logged out '%s'", login)
	DeleteToken(login)
	fmt.Fprintf(w, "Ok\n")
}

func wAuth(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ok, user := IsAuthorized(req)
		if !ok {
			http.Error(w, "Unauthorized", 401)
			return
		}
		req.Header.Add("user", user) // ugly
		fn(w, req)
	}
}

func list(w http.ResponseWriter, req *http.Request) {
	if Mode["files"] {
		listProblems(w)
		return
	} 
	ids, err := Problems.AllIDs()
	if err != nil {
		fmt.Fprintf(w, "ERROR: Cannot get AllIDs")
	}
	for _, id := range ids {
		fmt.Fprintf(w, "%s\n", id)
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
		V := sub.Veredict
		fmt.Fprintf(w, "%s\n", V.Message)
		if V.Message != "Accepted" && V.Details.Obj != nil {
			fmt.Fprintf(w, "\n%v", V.Details.Obj)
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
	flags := make(map[string]*bool)
	for _, f := range Modes {
		flags[f] = flag.Bool(f, false, "")
	}
	flag.Parse()

	accounts := flag.Args()
	if len(accounts) == 0 {
		accounts = []string{"local"}
	}

	for name, active := range flags {
		if *active {
			Mode[name] = true
			log.Printf("Mode '--%s'", name)
		}
	}

	if Mode["files"] { 
		if !Mode["open"] {
			Mode["open"] = true // --files => --open
			log.Printf("Mode '--open'")
		}
	}
	if !Mode["files"] {
		Problems = getDB("problems")
		if !Mode["open"] {
			Users = getDB("users")
		}
	}
	if Mode["local"] {
		Server = "localhost"
	}

	Queue.Init()
	launchEvaluators(accounts)
	http.HandleFunc("/open", open)
	http.HandleFunc("/login", login)
	http.HandleFunc("/logout", wAuth(logout))
	http.HandleFunc("/submit", wAuth(submit))
	http.HandleFunc("/list", wAuth(list))
	http.HandleFunc("/status/", wAuth(status))
	http.HandleFunc("/veredict/", wAuth(veredict))

	Url := fmt.Sprintf("%s:%d", Server, ListenPort)
	err := http.ListenAndServe(Url, nil)
	if err != nil {
		log.Printf("ListenAndServe: %s\n", err)
	}

	Queue.Close()
	waitForEvaluators()
}
