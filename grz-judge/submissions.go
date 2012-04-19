package main

import (
	"garzon/db"
	"garzon/eval"
	"sync"
	"time"
)

var Queue Submissions

type Submissions struct {
	Channel    chan string
	Mutex      sync.Mutex
	inprogress map[string]*eval.Submission
}

func (S *Submissions) Init() {
	S.inprogress = make(map[string]*eval.Submission)
	S.Channel = make(chan string, MaxQueueSize)
}

func (S *Submissions) Close() {
	close(S.Channel)
}

func (S *Submissions) Pending() int {
	return len(S.inprogress)
}

func (S *Submissions) Add(user string, pid string, problem *eval.Problem, sol string) (ID string) {
	ID = db.NewUUID()
	S.Mutex.Lock()
	S.inprogress[ID] = &eval.Submission{
		User:      user,
		ProblemID: pid,
		Problem:   problem,
		Solution:  sol,
		Submitted: time.Now(),
		Status:    "In Queue",
	}
	S.Mutex.Unlock()
	S.Channel <- ID
	return
}

func (S *Submissions) Get(id string) (sub *eval.Submission) {
	sub, ok := S.inprogress[id]
	if !ok {
		sub = nil
	}
	return
}

func (S *Submissions) SetStatus(id, state string) {
	S.Mutex.Lock()
	S.inprogress[id].Status = state
	S.Mutex.Unlock()
}

func (S *Submissions) Delete(id string) {
	S.Mutex.Lock()
	delete(S.inprogress, id)
	S.Mutex.Unlock()
}
