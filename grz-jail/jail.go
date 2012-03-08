
package main

import (
	"os"
	"os/exec"
	"fmt"
	"flag"
	"syscall"
)

var MaxCPUSeconds uint64
var MaxMemory     uint64 // in megabytes
var MaxFilesize   uint64 // in megabytes
var Accused bool
var limit syscall.Rlimit

func setrlimit(what int, max uint64) {
	limit.Cur = max
	limit.Max = max
	syscall.Setrlimit(what, &limit)
}

func usage() {
	fmt.Printf(`usage: grz-jail [options...] <exe>

Options:
   -m <mem>       Max megabytes
   -t <time>      Max seconds
   -f <size>      Max filesize
   -a             Accused mode

`)
}

func main() {
	flag.Uint64Var(&MaxCPUSeconds, "t", 2,  "")
	flag.Uint64Var(&MaxMemory,     "m", 64, "")
	flag.Uint64Var(&MaxFilesize,   "f", 5,  "")
	flag.BoolVar(&Accused, "a", false, "Launch in 'accused' mode")
	flag.Usage = usage
	flag.Parse()
	
	setrlimit(syscall.RLIMIT_CPU,   MaxCPUSeconds)
	setrlimit(syscall.RLIMIT_AS,    MaxMemory * 1024 * 1024)
	setrlimit(syscall.RLIMIT_FSIZE, MaxFilesize)

	args := flag.Args()
	if len(args) > 1 {
		fmt.Fprintf(os.Stderr, "Too many arguments\n")
		os.Exit(1)
	} else if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Too few arguments\n")
		os.Exit(1)
	}
	
	exe := args[0]
	
	a := "-A"
	if Accused {
		a = "-a"
	}
	cmd := exec.Command("systrace", a, exe)
	err := cmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}