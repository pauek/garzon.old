
package lang

type Language struct {
	Name, Extension string
	Functions Compiler
}

type Compiler interface {
	Compile(ID string) (string, error)
	Execute(ID string, input string) (string, error)
}

var Languages map[string]Language

func init() {
	Languages = make(map[string]Language)
	Languages["c++"] = Language{
	   Name: "c++", 
	   Extension: "cc", 
	   Functions: new(Cpp),
   }
}
