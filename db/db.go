
package db

import (
	"net/http"
	"bytes"
	"fmt"
	"log"
	"reflect"
	"io/ioutil"
	"encoding/json"
	"math/rand"
)

var client *http.Client

func init() {
	client = &http.Client{}
}

// UUIDs

const hex = "0123456789abcdef"

func NewUUID() string {
	var uuid [32]byte
	for i := 0; i < 32; i++ {
		uuid[i] = hex[rand.Intn(16)]
	}
	return fmt.Sprintf("%s", uuid)
}

// Type Map

var typMap map[string]reflect.Type

func init() {
	typMap = make(map[string]reflect.Type)
}

func typeName(v interface{}) string {
	typ := reflect.TypeOf(v)
	return typ.PkgPath() + ":" + typ.Name() 
}

func mustFindType(typname string) reflect.Type {
	typ, ok := typMap[typname]
	if ! ok {
		log.Fatalf("Tester '%s' not registered!\n", typname)
	}
	return typ
}

func Register(v interface{}) {
	typ := reflect.TypeOf(v)
	if typ.Kind() == reflect.Ptr {
		panic("Registering a pointer")
	}
	typMap[typeName(v)] = typ
}

// Mapper

type Map map[string]interface{}

func FromMap(M Map, v interface{}) error {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() == reflect.Struct {
		t := val.Type()
		z := reflect.Zero(t)
		n := z.NumField()
		for i := 0; i < n; i++ {
			f := val.Field(i)
			key := t.Field(i).Name
			if f.CanSet() {
				if fv, ok := M[key]; ok {
					f.Set(reflect.ValueOf(fv))
				}
			}
		}
		return nil
	}
	return fmt.Errorf("FromMap: Unsupported Kind '%s'", val.Kind())
}

func ToMap(in interface{}) (out interface{}, err error) {
	v := reflect.ValueOf(in)
	if v.Kind() == reflect.Struct {
		t := v.Type()
		z := reflect.Zero(t)
		n := z.NumField()
		var M Map = make(map[string]interface{})
		for i := 0; i < n; i++ {
			f := t.Field(i)
			if f.PkgPath != "" {
				continue // ignore lowercase fields
			}
			M[f.Name] = v.Field(i).Interface()
		}
		return M, nil
	}
	return nil, fmt.Errorf("ToMap: input not struct or map")
}

// Database

type Database struct {
	host, port, db string
}

func (D *Database) url(path string) string {
	return fmt.Sprintf("http://%s:%s/%s/%s", D.host, D.port, D.db, path)
}

func (D *Database) Get(id string) (v interface{}, rev string, err error) {
	v = nil
	rev = ""
	req, err := http.NewRequest("GET", D.url(id), nil)
	if err != nil {
		err = fmt.Errorf("Get: cannot create request: %s\n", err)
		return
	}
	resp, err := client.Do(req)
	switch {
	case err != nil:
		err = fmt.Errorf("Get: http.client error: %s\n", err)
		return
	case resp.StatusCode != 200:
		err = fmt.Errorf("Get: HTTP status = '%s'\n", resp.Status)
		return
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("Get: cannot read response body: %s\n", err)
		return 
	}
	var M Map
	if err = json.Unmarshal(data, &M); err != nil {
		err = fmt.Errorf("Get: json.Unmarshal error: %s\n", err)
		return 
	}
	typ := mustFindType(M[".type"].(string))
	v = reflect.New(typ).Interface()
	if err = FromMap(M, v); err != nil {
		err = fmt.Errorf("Get: FromMap error: %s\n", err)
		return
	}
	rev = M["_rev"].(string)
	return 
}

func (D *Database) Put(id string, v interface{}) error {
	out, err := ToMap(v)
	if err != nil {
		return fmt.Errorf("Put: ToMap error: %s\n", err)
	}
	M := out.(Map)
	M["_id"]   = id
	M[".type"] = typeName(v)
	return D.put(id, M)
}

func (D *Database) Update(id, rev string, v interface{}) error {
	out, err := ToMap(v)
	M := out.(map[string]interface{})
	if err != nil {
		return fmt.Errorf("Update: ToMap error: %s\n", err)
	}
	M["_id"]   = id
	M["_rev"]  = rev
	M[".type"] = typeName(v)
	return D.put(id, M)
}

func (D *Database) put(id string, M Map) error {
	data, err := json.Marshal(M)
	if err != nil {
		return fmt.Errorf("Put: json.Marshal error: %s\n", err)
	}
	buff := bytes.NewBuffer(data)
	req, err := http.NewRequest("PUT", D.url(id), buff)
	if err != nil {
		return fmt.Errorf("Put: cannot create request: %s\n", err)
	}
	resp, err := client.Do(req)
	switch {
	case err != nil:
		return fmt.Errorf("Put: http.client error: %s\n", err)
	case resp.StatusCode != 201:
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Printf("%s\n", body)
		return fmt.Errorf("Put: HTTP status = '%s'\n", resp.Status)
	}
	return nil
}

func (D *Database) Delete(id, rev string) error {
	req, err := http.NewRequest("DELETE", D.url(id), nil)
	req.Header.Set("If-Match", rev)
	if err != nil {
		return fmt.Errorf("Delete: cannot create request: %s\n", err)
	}
	resp, err := client.Do(req)
	switch {
	case err != nil:
		return fmt.Errorf("Delete: http.client error: %s\n", err)
	case resp.StatusCode == 404:
		return nil
	case resp.StatusCode != 200:
		return fmt.Errorf("Delete: HTTP status = '%s'\n", resp.Status)
	}
	return nil
}
