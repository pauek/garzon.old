
package client

import (
	"fmt"
	"testing"
)

func TestA(t *testing.T) {
	id, err := Submit("cpp.intro.SumaEnteros", "sumab.cc")
	if err != nil {
		t.Errorf("Submit: %s\n", err)
	}
	fmt.Printf("ID: %s", id)
}