
package problems

import (
	"fmt"
	"log"
	"garzon/db"
	"garzon/eval"
)

var D *db.Database

func init() {
	var err error
	D, err = db.Get("localhost:5984", "problems")
	if err != nil {
		D = nil
		log.Printf("db/problems/init: Warning: cannot get database")
	}
}

func Get(id string) (P *eval.Problem, err error) {
	if D == nil {
		return nil, fmt.Errorf("Database not available")
	}
	obj, _, err := D.Get(id)
	if err != nil {
		return nil, err
	}
	var ok bool
	P, ok = obj.(*eval.Problem)
	if ! ok {
		return nil, fmt.Errorf("Object is not a *Problem")
	}
	return P, nil
}