
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
   that take care of the "Inner" object. These methods write a field
   in the JSON data with the type of the object ("-type").

 - Every type that the database needs to care about has to be
   registered previously in the type map.

*/

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

// Database Object

type Obj struct {
	Inner interface{}
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
	return marshal(obj.Inner, map[string]string{ "-type": typeName(obj.Inner) })
}

func (obj *Obj) UnmarshalJSON(data []byte) (err error) {
	var t struct {
		Typ string `json:"-type"`
	}
	if err = json.Unmarshal(data, &t); err != nil {
		err = fmt.Errorf("Cannot json.Unmarshal id & rev: %s\n", err)
		return 
	}
	typ := mustFindType(t.Typ)
	obj.Inner = reflect.New(typ).Interface()
	if err = json.Unmarshal(data, obj.Inner); err != nil {
		obj.Inner = nil
		err = fmt.Errorf("Inner json.Unmarshal error: %s\n", err)
	}
	return
}

// Type Map

var typMap map[string]reflect.Type

func init() {
	typMap = make(map[string]reflect.Type)
}

func typeName(v interface{}) string {
	typ := reflect.TypeOf(v)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
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

// Database

type Database struct {
	host, port, db string
}

func (D *Database) url(path string) string {
	return fmt.Sprintf("http://%s:%s/%s/%s", D.host, D.port, D.db, path)
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
	v = obj.Inner
	return 
}

func (D *Database) Put(id string, v interface{}) error {
	return D.put(id, "", &Obj{v})
}

func (D *Database) Update(id, rev string, v interface{}) error {
	return D.put(id, rev, &Obj{v})
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
