package lang

type Language struct {
	Name, Extension string
	Functions       Compiler
}

type Compiler interface {
	Compile(infile, outfile string) error
	Execute(exefile, input string) (string, error)
}

type CompilationError struct {
	Output string
}

func (e *CompilationError) Error() string { return "Compilation Error" }

var Languages map[string]*Language

func init() {
	Languages = make(map[string]*Language)
	Languages["c++"] = &Language{
		Name:      "c++",
		Extension: "cc",
		Functions: new(Cpp),
	}
}

func Get(lang string) *Language {
	if L, ok := Languages[lang]; ok {
		return L
	}
	return nil
}
