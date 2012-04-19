package main

import (
	"fmt"
	"garzon/db"
	"garzon/eval"
	_ "garzon/eval/programming"
	"log"
	"strings"
)

func getDB(name string) *db.Database {
	D, err := db.GetDB(name)
	if err != nil {
		log.Fatalf("Cannot get database '%s': %s\n", name, err)
	}
	return D
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

func getProblem(probid string) (P *eval.Problem, err error) {
	if probid == "" {
		return nil, fmt.Errorf("Problem ID is empty")
	}
	if localMode {
		problem, err := eval.ReadFromID(probid)
		if err != nil {
			return nil, fmt.Errorf("Couldn't read problem '%s': %s\n", probid, err)
		}
		return problem, nil
	}
	var problem eval.Problem
	_, err = problems.Get(probid, &problem)
	if err != nil {
		return nil, fmt.Errorf("Cannot get problem '%s': %s\n", probid, err)
	}
	return &problem, nil
}
