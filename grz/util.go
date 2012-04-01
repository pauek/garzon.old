
package main

import (
	"os"
	"fmt"
	"strings"
)

func _err(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format + "\n", args...)
	os.Exit(1)
}

func splitPath(pathstr string) (path []string) {
	if pathstr == "" {
		return
	}
	for _, p := range strings.Split(pathstr, ":") {
		if p[len(p)-1] == '/' {
			p = p[:len(p)-1]
		}
		path = append(path, p)
	}
	return path
}