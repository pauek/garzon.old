package programming

import (
	"crypto/sha1"
	"fmt"
	"io"
	"strings"
)

func hash(s string) string {
	hash := sha1.New()
	io.WriteString(hash, s)
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func prefix(s string, length int) string {
	max := length
	suffix := "..."
	if len(s) < length {
		max = len(s)
		suffix = ""
	}
	return strings.Replace(s[:max], "\n", `\n`, -1) + suffix
}

func seeSpace(a string) (s string) {
	replacements := []struct{ from, to string }{
		{" ", "\u2423"},
		{"\n", "\u21B2\n"},
	}
	s = a
	for _, r := range replacements {
		s = strings.Replace(s, r.from, r.to, -1)
	}
	return
}

func parsePerformance(s string) (perf Performance) {
	lines := strings.Split(s, "\n")
	if len(lines) > 1 {
		fmt.Sscanf(lines[1], "%f", &perf.Seconds)
	}
	if len(lines) > 2 {
		fmt.Sscanf(lines[2], "%f", &perf.Megabytes)
	}
	return
}
