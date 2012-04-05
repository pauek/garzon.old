
package db

var problems *Database

func getdb(id string) *Database {
	D, err := Get("localhost:5984", id)
	if err != nil { D = nil }
	return D
}

func Problems() *Database    { return getdb("problems") }
func Submissions() *Database { return getdb("submissions") }

