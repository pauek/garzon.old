
/*

 Garzón DB

 This module implements a "Polymorphic Object Database" (I really
 don't know how to call it). It is basically a way to store objects in
 CouchDB along with their type. When requesting an ID, you don't know
 in advance the type of the object you are going to get. This is
 necessary since Garzón has many different types of evaluators and
 tests.

 On top of this, objects may contain heterogeneous arrays of other
 objects inside. For example, problem objects contain an array of
 tests.

 The way this is handled is the following:

 - Every object that needs to be stored polymorphically has to be
   "decorated" (enclosed) in an object of type db.Obj.

 - The type db.Obj has special MarshalJSON and UnmarshalJSON methods
   that take care of the "Obj" object. These methods write a field
   in the JSON data with the type of the object ("-type").

 - Every type that the database needs to care about has to be
   registered previously in the type map.

*/

package db

import (
	"fmt"
	"log"
	"bytes"
	"strings"
	"reflect"
	"io/ioutil"
	"net/http"
	"encoding/json"
	"encoding/gob"
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

// Database Object

type Obj struct {
	Obj interface{}
}

func init() {
	gob.Register(Obj{})
}

func marshal(v interface{}, preamble map[string]string) ([]byte, error) {
	var b bytes.Buffer
	fmt.Fprintf(&b, "{")
	for key, value := range preamble {
		if value != "" {
			fmt.Fprintf(&b, `"%s":"%s",`, key, value)
		}
	}
	json, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(&b, "%s", json[1:]) // includes '}'
	return b.Bytes(), nil
}

func (obj *Obj) MarshalJSON() ([]byte, error) {
	return marshal(obj.Obj, map[string]string{ "-type": findAlias(obj.Obj) })
}

func (obj *Obj) UnmarshalJSON(data []byte) (err error) {
	var t struct {
		Alias string `json:"-type"`
	}
	if err = json.Unmarshal(data, &t); err != nil {
		err = fmt.Errorf("Cannot json.Unmarshal id & rev: %s\n", err)
		return 
	}
	typ := findType(t.Alias)
	obj.Obj = reflect.New(typ).Interface()
	if err = json.Unmarshal(data, obj.Obj); err != nil {
		obj.Obj = nil
		err = fmt.Errorf("Obj json.Unmarshal error: %s\n", err)
	}
	return
}

// Type Map

type TypeInfo struct {
	Typ reflect.Type
	Alias string
}

var typeMap  map[string]TypeInfo // typename -> TypeInfo
var aliasMap map[string]string   // alias -> typename

func init() {
	typeMap  = make(map[string]TypeInfo)
	aliasMap = make(map[string]string)
}

func typeName(v interface{}) string {
	typ := reflect.TypeOf(v)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	return typ.PkgPath() + ":" + typ.Name() 
}

func mustFindTypeInfo(typename string) TypeInfo {
	t, ok := typeMap[typename]
	if ! ok {
		log.Fatalf("Typename '%s' not found!", typename)
	}
	return t
}

func findAlias(v interface{}) string {
	typ := mustFindTypeInfo(typeName(v))
	return typ.Alias
}

func findType(alias string) reflect.Type {
	typename, ok := aliasMap[alias]
	if ! ok {
		log.Fatalf("Alias '%s' not found!", alias)
	}
	return mustFindTypeInfo(typename).Typ
}

// Create an object from a registered type by alias
func ObjFromType(alias string) interface{} {
	typename, ok := aliasMap[alias]
	if ! ok { return nil }
	typ, ok := typeMap[typename]
	if ! ok { return nil }
	return reflect.New(typ.Typ).Interface()
}

// Register a database object by alias
func Register(alias string, v interface{}) {
	typ := reflect.TypeOf(v)
	if typ.Kind() == reflect.Ptr {
		panic("Registering a pointer!")
	}
	typename := typeName(v)
	typeMap[typename] = TypeInfo{Typ: typ, Alias: alias}
	_, ok := aliasMap[alias]
	if ok {
		panic(fmt.Sprintf("Alias '%s' already registered!", alias))
	}
	aliasMap[alias] = typename
}

// Database

type Database struct {
	server, dbname string
}

func (D *Database) url(path string) string {
	return fmt.Sprintf("http://%s/%s/%s", D.server, D.dbname, path)
}

func (D *Database) Rev(id string) (rev string, err error) {
	rev, err = "", nil
	req, err := http.NewRequest("HEAD", D.url(id), nil)
	if err != nil {
		err = fmt.Errorf("Rev: cannot create request: %s\n", err)
		return
	}
	resp, err := client.Do(req)
	switch {
	case err != nil:
		err = fmt.Errorf("Rev: http.client error: %s\n", err)
		return
	case resp.StatusCode == 404:
		err = nil // not found is not an error
		return
	case resp.StatusCode != 200:
		err = fmt.Errorf("Rev: HTTP status = '%s'\n", resp.Status)
		return
	}
	rev = resp.Header.Get("Etag")
	if rev == "" {
		err = fmt.Errorf("Rev: Header 'Etag' not found\n")
	}
	rev = strings.Replace(rev, `"`, ``, -1)
	return
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
	var obj Obj
	if err = json.Unmarshal(data, &obj); err != nil {
		err = fmt.Errorf("Get: json.Unmarshal error: %s\n", err)
		return 
	}
	v = obj.Obj
	rev = resp.Header.Get("Etag")
	return 
}

func (D *Database) Put(id string, v interface{}) error {
	return D.put(id, "", &Obj{v})
}

func (D *Database) Update(id, rev string, v interface{}) error {
	return D.put(id, rev, &Obj{v})
}

func (D *Database) PutOrUpdate(id string, v interface{}) error {
	rev, err := D.Rev(id)
	if err != nil {
		return fmt.Errorf("PutOrUpdate: %s\n", err)
	}
	if rev == "" {
		return D.Put(id, v) 
	} 
	return D.Update(id, rev, v)
}

func (D *Database) put(id, rev string, v interface{}) error {
	// TODO: Detect that 'v' really is db.Obj
	preamble := map[string]string {
		"_id": id, 
		"_rev": rev,
	}
	json, err := marshal(v, preamble)
	if err != nil {
		return fmt.Errorf("Put: json.Marshal error: %s\n", err)
	}
	b := bytes.NewBuffer(json)
	req, err := http.NewRequest("PUT", D.url(id), b)
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

// Database Functions

func Get(server, dbname string) (db *Database, err error) {
	db = nil
	url := fmt.Sprintf("http://%s/%s/", server, dbname)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		err = fmt.Errorf("Get: cannot create request: %s\n", err)
		return
	}
	resp, err := client.Do(req)
	switch {
	case err != nil:
		err = fmt.Errorf("Get: http.client error: %s\n", err)
		return
	case resp.StatusCode == 404:
		err = fmt.Errorf("Get: database '%s' doesn't exist\n", dbname)
		return
	case resp.StatusCode != 200:
		err = fmt.Errorf("Get: HTTP status = '%s'\n", resp.Status)
		return
	}
	return &Database{server, dbname}, nil
}

func Create(server, dbname string) (db *Database, err error) {
	db = nil
	url := fmt.Sprintf("http://%s/%s/", server, dbname)
	req, err := http.NewRequest("PUT", url, nil)
	if err != nil {
		err = fmt.Errorf("Get: cannot create request: %s\n", err)
		return
	}
	resp, err := client.Do(req)
	switch {
	case err != nil:
		err = fmt.Errorf("Create: http.client error: %s\n", err)
		return
	case resp.StatusCode != 201:
		err = fmt.Errorf("Create: HTTP status = '%s'\n", resp.Status)
		return
	}
	return &Database{server, dbname}, nil
}

func Delete(db *Database) (err error) {
	req, err := http.NewRequest("DELETE", db.url(""), nil)
	if err != nil {
		err = fmt.Errorf("Get: cannot create request: %s\n", err)
		return
	}
	resp, err := client.Do(req)
	switch {
	case err != nil:
		err = fmt.Errorf("Create: http.client error: %s\n", err)
		return
	case resp.StatusCode == 404:
		return
	case resp.StatusCode != 200:
		err = fmt.Errorf("Create: HTTP status = '%s'\n", resp.Status)
		return
	}
	return
}

func GetOrCreate(server, dbname string) (db *Database, err error) {
	db, err = Get(server, dbname)
	if db == nil {
		db, err = Create(server, dbname)
	}
	return
}

