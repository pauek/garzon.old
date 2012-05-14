
package programming

import (
	"io"
	"fmt"
	"crypto/sha1"
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
