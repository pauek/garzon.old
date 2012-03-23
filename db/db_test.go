
package db

import (
	"fmt"
	//	"encoding/json"
	"testing"
)

type Test1 struct {
	A string
	B int
	C bool
}

func TestA(t *testing.T) {
	m, _ := ToMap(Test1{A: "hi", B: 2, C: true})
	fmt.Println(m)
	var x Test1
	_ = FromMap(m.(Map), &x)
	fmt.Println(x)
}

type Test2 struct {
	D Test1
	E int
	F string
}

func TestB(t *testing.T) {
	m, _ := ToMap(Test2{D: Test1{A: "ho", B: 3, C: false}, E: 4, F: "ha"})
	fmt.Println(m)
}