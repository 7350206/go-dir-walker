package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

func filterOut(path, ext string, minSize int64, info os.FileInfo) bool {
	if info.IsDir() || info.Size() < minSize {
		return true
	}

	// if the fn received a value for the ext argument
	// extract the extension of the file and compare it to the ext argument.
	if ext != "" && filepath.Ext(path) != ext {
		return true
	}
	return false
}

func listFile(path string, out io.Writer) error {
	_, err := fmt.Fprintln(out, path)
	return err
}

// receives one argument: the file path to be deleted.
// return the potential error from os.Remove directly
// as the return value of function.
// if os.Remove fails to delete the file, its error will bubble up,
// stopping the tool’s execution and showing the error message to the user.
func delFile(path string, delLogger *log.Logger) error {
	// return os.Remove(path)
	if err := os.Remove(path); err != nil {
		return err
	}
	delLogger.Println(path)
	return nil
}
