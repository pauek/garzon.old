package main

import (
	"flag"
	"fmt"
	"strings"
	"time"

	"garzon/grz-judge/client"
)

const u_submit = `grz submit <ProblemID> <filename>`

func submit(args []string) {
	var url string
	fset := flag.NewFlagSet("submit", flag.ExitOnError)
	fset.StringVar(&url, "judge", "", "URL for the Judge")
	fset.Parse(args)

	if url != "" {
		client.JudgeUrl = url
	}

	probid, filename := checkTwoArgs("submit", fset.Args())

	var err error
	client.AuthToken, err = readAuthToken()
	if err != nil {
		_errx("Cannot read Auth Token: %s", err)
	}
	resp, err := client.Submit(probid, filename)
	if err != nil {
		_errx("Submission error: %s", err)
	}
	if strings.HasPrefix(resp, "ERROR") {
		_errx("%s", resp)
	}
	id := resp

	for {
		status, err := client.Status(id)
		if err != nil {
			_errx("Cannot get status: %s", err)
		}
		if status == "Resolved" {
			break
		}
		fmt.Printf("\r                                         \r")
		fmt.Printf("%s...", status)
		time.Sleep(500 * time.Millisecond)
	}

	veredict, err := client.Veredict(id)
	if err != nil {
		_errx("Cannot get veredict: %s", err)
	}
	fmt.Printf("\r                                         \r")
	fmt.Print(veredict)
}
