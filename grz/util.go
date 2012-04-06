
package main

import (
	"os"
	"fmt"
	"strings"
)

func _err(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "grz: " + format + "\n", args...)
}

func _errx(format string, args ...interface{}) {
	_err(format, args...)
	os.Exit(2)
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