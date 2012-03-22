
package eval

import (
	"fmt"
	"reflect"
	"encoding/json"
	"log"
)	

// Problems //////////////////////////////////////////////////////////

type Problem struct {
	ID, Title, Solution string
	Tests []Tester
}

func (P *Problem) UnmarshalJSON(data []byte) error {
	var p interface{}
	json.Unmarshal(data, &p)
	m, ok := p.(map[string]interface{})
	if ! ok {
		fmt.Errorf("Cannot unmarshal JSON data")
	}
	P.ID       = m["_id"].(string)
	P.Title    = m["Title"].(string)
	P.Solution = m["Solution"].(string)
	
	tests := m["Tests"].([]interface{})
	P.Tests = make([]Tester, len(tests))
	for i, T := range tests {
		P.Tests[i] = *TesterFromJSON(T)
	}
	return nil
}

func (P *Problem) MarshalJSON() ([]byte, error) {
	M := make(map[string]interface{})
	M["_id"] = P.ID
	M["Title"] = P.Title
	M["Solution"] = P.Solution
	tests := make([]interface{}, len(P.Tests))
	for i, T := range P.Tests {
		tests[i] = TesterToJSON(T)
	}
	M["Tests"] = tests
	return json.Marshal(M)
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
	typeMap[name] = reflect.TypeOf(t)
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
	tester = reflect.New(typ).Interface().(*Tester)
	(*tester).FromMap(M)
	return
}

func TesterToJSON(tester Tester) interface{} {
	typname := typeName(tester)
	_ = mustFindType(typname)
	M := tester.ToMap()
	M["_type"] = typname
	return M
}
