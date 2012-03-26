
package eval

type Problem struct {
	Title, Solution string
	Tests []Tester
}

type Tester interface {
	Veredict() Result
}

type Result struct {
	Veredict string
	Reason   interface{}
}
