
/*

 Garzón DB

 This module implements a "Polymorphic Object Database" (I really
 don't know how to call it). It is basically a way to store objects in
 CouchDB along with their type. When requesting an ID, you don't know
 in advance the type of the object you are going to get. This is
 necessary since Garzón has many different types of evaluators and
 tests.

 This requires that: 1) types register in a "Type Map", which
 associates type names with types themselves (since Go doesn't seem to
 provide this); 2) the JSON text that is sent to CouchDB includes
 a "-type" ("_type" is illegal) field with the type name of the stored
 object (along with an "_id" and "_rev" which are characteristic of
 CouchDB). Point number 2 is implemented quite "hackisly" inserting
 these fields in the textual representation that json.Marshal returns.

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

type couchobj struct {
	Id  string `json:"_id"`
	Rev string `json:"_rev,omitempty"`
	Typ string `json:"-type"`
}

func (D *Database) url(path string) string {
	return fmt.Sprintf("http://%s:%s/%s/%s", D.host, D.port, D.db, path)
}

func (D *Database) Rev(id string) (rev string, err error) {
	rev, err = "", nil
	req, err := http.NewRequest("HEAD", D.url(id), nil)
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
		err = nil // not found is not an error
		return
	case resp.StatusCode != 200:
		err = fmt.Errorf("Get: HTTP status = '%s'\n", resp.Status)
		return
	}
	rev = resp.Header.Get("Etag")
	if rev == "" {
		err = fmt.Errorf("Header 'Etag' not found\n")
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
	var obj couchobj
	if err = json.Unmarshal(data, &obj); err != nil {
		err = fmt.Errorf("Get: Head json.Unmarshal error: %s\n", err)
		return 
	}
	rev = obj.Rev
	typ := mustFindType(obj.Typ)
	v = reflect.New(typ).Interface()
	if err = json.Unmarshal(data, v); err != nil {
		v, err = nil, fmt.Errorf("Get: Body json.Unmarshal error: %s\n", err)
		return
	}
	return 
}

func (D *Database) Put(id string, v interface{}) error {
	return D.put(id, couchobj{ id, "", typeName(v) }, v)
}

func (D *Database) Update(id, rev string, v interface{}) error {
	return D.put(id, couchobj{ id, rev, typeName(v) }, v)
}

func (D *Database) put(id string, obj couchobj, v interface{}) error {
	head, err := json.Marshal(obj)
	if err != nil {
		return fmt.Errorf("Put: Head json.Marshal error: %s\n", err)
	}
	body, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("Put: Body json.Marshal error: %s\n", err)
	}
	fmt.Printf("head: '%s'\n", head)
	fmt.Printf("body: '%s'\n", head)
	// Begin Hackish part
	body[0] = ',' // remove first '{'
	head = head[:len(head)-1] // remove last '}'
	fmt.Printf("%s%s\n", head, body)
	var buff bytes.Buffer
	if _, err := buff.Write(head); err != nil {
		return fmt.Errorf("Put: Write head to buffer error: %s\n", err)
	}
	if _, err := buff.Write(body); err != nil {
		return fmt.Errorf("Put: Write body to buffer error: %s\n", err)
	}
	// End Hackish part
	req, err := http.NewRequest("PUT", D.url(id), &buff)
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
