package main

import (
	"flag"
	"fmt"
	"strings"
	"io/ioutil"

	"github.com/pauek/garzon/grz-judge/client"
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

	maybeReadAuthToken()

	probid, filename := checkTwoArgs("submit", fset.Args())

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		_errx("Cannot read file '%s'", filename)
	}
	resp, err := client.Submit(probid, filename, data)
	if err != nil {
		_errx("Submission error: %s", err)
	}
	if strings.HasPrefix(resp, "ERROR") {
		_errx("%s", resp)
	}
	id := resp

	err = client.Status(id, func(status string) {
		fmt.Printf("\r                                         \r")
		fmt.Printf("%s...", status)
	})
	if err != nil {
		_errx("%s", err)
	}

	veredict, err := client.Veredict(id)
	if err != nil {
		_errx("Cannot get veredict: %s", err)
	}
	fmt.Printf("\r                                         \r")
	fmt.Print(veredict)
}
