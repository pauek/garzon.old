package main

import (
	"fmt"
	"os"
	"path/filepath"
	"io/ioutil"
)

func _err(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "grz: "+format+"\n", args...)
}

func _errx(format string, args ...interface{}) {
	_err(format+"\n", args...)
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

func saveAuthToken(tok string) error {
	filename := configFile("auth", true)
	err := ioutil.WriteFile(filename, []byte(tok), 0600)
	if err != nil {
		return fmt.Errorf("Cannot write '%s': %s", err)
	}
	return nil
}

func readAuthToken() (string, error) {
	// TODO: Detect that the file is missing to report "you should login first"
	filename := configFile("auth", false)
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("Cannot read '%s': %s", filename, err)
	}
	return string(data), nil
}

func removeAuthToken() error {
	filename := configFile("auth", false)
	err := os.Remove(filename)
	if err != nil {
		return fmt.Errorf("Cannot remove '%s': %s", filename, err)
	}
	return nil
}