package main

import (
	"fmt"
	"garzon/grz-judge/client"
	"io/ioutil"
	"os"
	"path/filepath"
)

func _err(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "grz: "+format+"\n", args...)
}

func _errx(format string, args ...interface{}) {
	_err(format, args...)
	os.Exit(2)
}

func checkNArgs(n int, cmd string, iargs []string) (oargs []string) {
	if len(iargs) != n {
		_err("Wrong number of arguments")
		usageCmd(cmd, 2)
	}
	return iargs
}

func checkOneArg(cmd string, args []string) string {
	return checkNArgs(1, cmd, args)[0]
}

func checkTwoArgs(cmd string, iargs []string) (a, b string) {
	oargs := checkNArgs(2, cmd, iargs)
	return oargs[0], oargs[1]
}

func maybeCreateDir(dir string) error {
	info, err := os.Stat(dir)
	if err == nil {
		if !info.IsDir() {
			return fmt.Errorf("'%s' exists and is not a directory", dir)
		}
	} else {
		err := os.Mkdir(dir, 0700)
		if err != nil {
			return fmt.Errorf("Cannot create directory '%s'", dir)
		}
	}
	return nil
}

func configFile(name string, createParents bool) string {
	configDir := filepath.Join(os.Getenv("HOME"), ".config")
	if createParents {
		maybeCreateDir(configDir)
	}
	garzonDir := filepath.Join(configDir, "garzon")
	if createParents {
		maybeCreateDir(garzonDir)
	}
	return filepath.Join(garzonDir, "auth")
}

func saveAuthToken() error {
	tok := client.AuthToken
	filename := configFile("auth", true)
	err := ioutil.WriteFile(filename, []byte(tok), 0600)
	if err != nil {
		_errx("Cannot write auth token to '%s': %s", err)
	}
	return nil
}

func readAuthToken() {
	filename := configFile("auth", false)
	_, err := os.Stat(filename)
	if err != nil {
		if os.IsNotExist(err.(*os.PathError).Err) {
			_errx("You should login first")
		} else {
			_errx("error: %s", err)
		}
	}
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		_errx("Cannot read '%s': %s", filename, err)
	}
	client.AuthToken = string(data)
}

func removeAuthToken() error {
	filename := configFile("auth", false)
	err := os.Remove(filename)
	if err != nil {
		return fmt.Errorf("Cannot remove '%s': %s", filename, err)
	}
	return nil
}
