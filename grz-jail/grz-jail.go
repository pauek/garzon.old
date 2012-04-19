package main

/* 

#include "grz-jail.h"

*/
import "C"

import (
	"flag"
	"fmt"
	"os"
)

var (
	MaxCpuSeconds int
	MaxMemory     int
	MaxFileSize   int
	AccusedMode   bool
)

const s_usage = `usage: grz-jail [options...] <directory>

Options:
   -m <mem>   Max megabytes of memory
   -t <mem>   Max seconds
   -f <mem>   Max megabytes for files
   -a         Accused mode

`

func usage() {
	fmt.Fprintf(os.Stderr, s_usage)
}

func main() {
	flag.Usage = usage
	flag.IntVar(&MaxCpuSeconds, "t", 2, "<dummy>")
	flag.IntVar(&MaxMemory, "m", 64*1024*1024, "<dummy>")
	flag.IntVar(&MaxCpuSeconds, "f", 1024, "<dummy>")
	flag.BoolVar(&AccusedMode, "b", false, "<dummy>")
	flag.Parse()
	args := flag.Args()
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "Wrong number of arguments")
		os.Exit(3)
	}
	C.max_cpu_seconds = C.int(MaxCpuSeconds)
	C.max_memory = C.int(MaxMemory)
	C.max_file_size = C.int(MaxFileSize)
	if AccusedMode {
		C.accused_mode = C.int(1)
	}
	C.grzjail(C.CString(args[0]))
	fmt.Fprintf(os.Stderr, "You've just seen an error in The Matrix") // we shouldn't be here
	os.Exit(3)
}
