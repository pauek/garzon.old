package main

import (
	"code.google.com/p/go.crypto/bcrypt"
	"crypto/rand"
	"encoding/base64"
	"github.com/pauek/garzon/db"
	"io"
	"net/http"
)

var (
	tokens = make(map[string]string)
	logins = make(map[string]string)
)

const alphabet = "./ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

var bcEncoding = base64.NewEncoding(alphabet)

func base64Encode(src []byte) []byte {
	n := bcEncoding.EncodedLen(len(src))
	dst := make([]byte, n)
	bcEncoding.Encode(dst, src)
	for dst[n-1] == '=' {
		n--
	}
	return dst[:n]
}

func LoginCorrect(login, passwd string) bool {
	// find user in DB
	var user db.User
	_, err := Users.Get(login, &user)
	if err != nil {
		return false
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Hpasswd), []byte(passwd))
	if err != nil {
		return false
	}
	return true
}

func CreateToken(login string) (string, error) {
	unencodedTok := make([]byte, 32)
	_, err := io.ReadFull(rand.Reader, unencodedTok)
	if err != nil {
		return "", err
	}
	tok := string(base64Encode(unencodedTok))
	tokens[tok] = login
	logins[login] = tok
	return tok, nil
}

func TokenExists(tok string) (bool, string) {
	login, ok := tokens[tok]
	return ok, login
}

func DeleteToken(login string) {
	tok, ok := logins[login]
	if !ok {
		panic("Token not found!")
	}
	delete(tokens, tok)
	delete(logins, login)
}

func IsAuthorized(req *http.Request) (bool, string) {
	if Mode["open"] {
		return true, "[anonymous]"
	}
	var cookie *http.Cookie
	for _, c := range req.Cookies() {
		if c.Name == "Auth" {
			cookie = c
			break
		}
	}
	if cookie == nil {
		return false, ""
	}
	return TokenExists(cookie.Value)
}
