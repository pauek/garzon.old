package main

import (
	"github.com/pauek/garzon/db"
	"github.com/pauek/garzon/eval"
	"sync"
	"time"
)

var queue Queue

type Queue struct {
	Channel    chan string
	Mutex      sync.Mutex
	inprogress map[string]*eval.Submission
	progress   map[string]chan string
}

func (Q *Queue) Init() {
	Q.inprogress = make(map[string]*eval.Submission)
	Q.progress = make(map[string]chan string, 1)
	Q.Channel = make(chan string, MaxQueueSize)
}

func (Q *Queue) Close() {
	close(Q.Channel)
}

func (Q *Queue) Pending() int {
	return len(Q.inprogress)
}

func (Q *Queue) Add(user string, pid string, problem *eval.Problem, sol string) (ID string) {
	ID = db.NewUUID()
	Q.Mutex.Lock()
	Q.inprogress[ID] = &eval.Submission{
		User:      user,
		ProblemID: pid,
		Problem:   problem,
		Solution:  sol,
		Submitted: time.Now(),
	}
	Q.progress[ID] = make(chan string, 1)
	Q.progress[ID] <- "In queue"
	Q.Mutex.Unlock()
	Q.Channel <- ID
	return
}

func (Q *Queue) Get(id string) (sub *eval.Submission) {
	sub, ok := Q.inprogress[id]
	if !ok {
		sub = nil
	}
	return
}

func (Q *Queue) SetStatus(id, state string) {
	Q.Mutex.Lock()
	Q.progress[id] <- state
	Q.Mutex.Unlock()
}

func (Q *Queue) GetStatus(id string) string {
	return <- Q.progress[id]
}

func (Q *Queue) Delete(id string) {
	Q.Mutex.Lock()
	delete(Q.inprogress, id)
	Q.Mutex.Unlock()
}
