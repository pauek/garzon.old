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

var byName = make(map[string]*Language)
var byExt = make(map[string]*Language)

func Register(lang *Language) {
	byName[lang.Name] = lang
	byExt[lang.Extension] = lang
}

func ByName(name string) *Language {
	if L, ok := byName[name]; ok {
		return L
	}
	return nil
}

func ByExtension(ext string) *Language {
	if L, ok := byExt[ext]; ok {
		return L
	}
	return nil
}