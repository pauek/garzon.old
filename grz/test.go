package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"garzon/grz-judge/client"
)

const u_test = `grz test <directory> <filename>

Options:
  --judge   URL for the judge
  --path    Problem root directory

`

func test(args []string) {
	var url string
	fset := flag.NewFlagSet("add", flag.ExitOnError)
	fset.StringVar(&Path, "path", "", "Problem root directory")
	fset.StringVar(&url, "judge", "", "URL for the Judge")
	fset.Parse(args)
	if Path == "" {
		Path = os.Getenv("GRZ_PATH")
	}
	if url != "" {
		client.JudgeUrl = url
	}
	args = fset.Args()
	if len(args) != 2 {
		_err("Wrong number of arguments")
		usageCmd("test", 2)
	}

	dir, filename := args[0], args[1]

	directory, err := filepath.Abs(dir)
	if err != nil {
		_errx("Error with dir '%s': %s\n", err)
	}
	_, problem := readProblem(directory)
	json, err := json.Marshal(problem)
	if err != nil {
		_errx("cannot Marshal: %s\n", err)
	}

	// send to judge
	resp, err := client.Test(string(json), filename)
	if err != nil {
		_errx("test error: %s\n", err)
	}
	if strings.HasPrefix(resp, "ERROR") {
		_errx("%s\n", resp)
	}
	id := resp

	for {
		status, err := client.Status(id)
		if err != nil {
			_errx("Cannot get status: %s\n", err)
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
		_errx("Cannot get veredict: %s\n", err)
	}
	fmt.Printf("\r                                         \r")
	fmt.Print(veredict)
}
