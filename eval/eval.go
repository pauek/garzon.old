
package eval

import (
	"log"
	"reflect"
)	

// Problems //////////////////////////////////////////////////////////

type Problem struct {
	Id, Title, Solution string
	Tests []Tester
}

func (P *Problem) FromJSON(M map[string]interface{}) {
	P.Id       = M["_id"].(string)
	P.Title    = M["Title"].(string)
	P.Solution = M["Solution"].(string)
	
	tests := M["Tests"].([]interface{})
	P.Tests = make([]Tester, len(tests))
	for i, T := range tests {
		P.Tests[i] = *TesterFromJSON(T)
	}
}

func (P *Problem) ToJSON() (M map[string]interface{}) {
	M = make(map[string]interface{})
	M["_id"] = P.Id
	M["Title"] = P.Title
	M["Solution"] = P.Solution
	tests := make([]interface{}, len(P.Tests))
	for i, T := range P.Tests {
		tests[i] = TesterToJSON(&T)
	}
	M["Tests"] = tests
	return
}

// Testers ///////////////////////////////////////////////////////////

type Result struct {
	Veredict string
	Reason   interface{}
}

type Tester interface {
	Veredict() Result
	ToMap() map[string]interface{}       // to JSON
	FromMap(data map[string]interface{}) // from JSON
}


var typeMap map[string]reflect.Type

func init() {
	typeMap = make(map[string]reflect.Type)
}

func typeName(v interface{}) string {
	typ := reflect.TypeOf(v).Elem()
	return typ.PkgPath() + ":" + typ.Name()
}

func RegisterTester(t Tester) {
	name := typeName(t)
	typeMap[name] = reflect.TypeOf(t).Elem()
}

func mustFindType(typname string) reflect.Type {
	typ, ok := typeMap[typname]
	if ! ok {
		log.Fatalf("Tester '%s' not registered!\n", typname)
	}
	return typ
}

func TesterFromJSON(v interface{}) (tester *Tester) {
	M := v.(map[string]interface{})
	typname := M["_type"].(string)
	typ := mustFindType(typname)
	t := reflect.New(typ).Interface().(Tester)
	t.FromMap(M)
	return &t
}

func TesterToJSON(tester *Tester) interface{} {
	typname := typeName(*tester)
	_ = mustFindType(typname)
	M := (*tester).ToMap()
	M["_type"] = typname
	return M
}
