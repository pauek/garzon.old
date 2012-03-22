
package eval

type Result struct {
	Veredict string
	Reason   interface{}
}

type Tester interface {
	Veredict() Result
}