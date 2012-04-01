
package main

import (
	"strings"
)

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