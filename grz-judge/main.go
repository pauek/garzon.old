package main

import (
	"code.google.com/p/go.net/websocket"
	"flag"
	"fmt"
	"github.com/pauek/garzon/db"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

const (
	ListenPort   = 50000
	MaxQueueSize = 50
)

var (
	Server string
	Mode   = make(map[string]bool)
	Modes  = []string{"copy", "debug", "local", "open", "nolog", "files"}

	Problems    *db.Database
	Users       *db.Database
	Submissions *db.Database
)

const usage = `usage: grz-judge [options...] [accounts...]

Options:
   --copy,    Copy 'grz-{eval,jail}' to remote accounts.
   --debug,   Tell 'grz-eval' to keep evaluation directories.
   --local,   Local mode (only listen to localhost:50000).
   --open,    Open mode (no authentication, submissions not stored).
   --nolog,   Do not store submissions in the database.
   --files,   Read problems from disk (implies --open and --nolog).

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
	if queue.Pending() > MaxQueueSize {
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
	ID := queue.Add(user, id, problem, string(solution))
	fmt.Fprintf(w, "%s", ID)
	return
}

func wsStatus(id string) websocket.Handler {
	return func(ws *websocket.Conn) {
		for {
			msg := queue.GetStatus(id)
			err := websocket.Message.Send(ws, []byte(msg))
			if err != nil {
				break
			}
			if msg == "Resolved" || strings.HasPrefix(msg, "Error") {
				break
			}
		}

	}
}

func status(w http.ResponseWriter, req *http.Request) {
	id := req.URL.Path[len("/status/"):]
	sub := queue.Get(id)
	if sub != nil {
		websocket.Handler(wsStatus(id)).ServeHTTP(w, req)
		return
	} else {
		fmt.Fprint(w, "Not found\n")
	}
}

func veredict(w http.ResponseWriter, req *http.Request) {
	id := req.URL.Path[len("/veredict/"):]
	sub := queue.Get(id)
	if sub == nil {
		fmt.Fprint(w, "Not found\n")
		return
	}
	V := sub.Veredict
	fmt.Fprintf(w, "%s\n", V.Message)
	if V.Message != "Accepted" && V.Details.Obj != nil {
		fmt.Fprintf(w, "\n%v", V.Details.Obj)
	}
	queue.Delete(id)
}

func handleAndShowFlags(flags map[string]*bool) {
	for name, active := range flags {
		if *active {
			Mode[name] = true
			log.Printf("Mode '--%s'", name)
		}
	}

	if Mode["files"] {
		if !Mode["open"] {
			Mode["open"] = true // --files => --open
		}
		if !Mode["nolog"] {
			Mode["nolog"] = true // --files => --nolog
		}
	}
	if !Mode["files"] {
		Problems = getDB("problems")
	}
	if !Mode["open"] {
		Users = getDB("users")
	}
	if !Mode["nolog"] {
		Submissions = getDB("submissions")
	}
	if Mode["local"] {
		Server = "localhost"
	}
	for name, _ := range Mode {
		log.Printf("Mode '--%s'", name)
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

	handleAndShowFlags(flags)

	queue.Init()
	launchEvaluators(accounts)
	http.HandleFunc("/open", open)
	http.HandleFunc("/login", login)
	http.HandleFunc("/logout", wAuth(logout))
	http.HandleFunc("/submit", wAuth(submit))
	http.HandleFunc("/list", wAuth(list))
	http.HandleFunc("/status/", status)
	http.HandleFunc("/veredict/", wAuth(veredict))

	Url := fmt.Sprintf("%s:%d", Server, ListenPort)
	err := http.ListenAndServe(Url, nil)
	if err != nil {
		log.Printf("ListenAndServe: %s\n", err)
	}

	queue.Close()
	waitForEvaluators()
}
