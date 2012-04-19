
package db

type User struct {
	Login string
	Hpasswd string // hashed password (w/ salt)
}
