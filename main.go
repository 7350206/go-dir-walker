package main

import (
	"flag"
	"fmt"
	"io" // use io.Writer
	"os"
	"path/filepath"
)

// list of params to run will be long,
// provide some of these args packaged in a custom type.
type config struct {
	ext  string // extension to filter out
	size int64  // min file size
	list bool   // list files
}

// descend dir by root and find all diles and sub-dirs on it
// checking to see if the provided error is not nil,
// which means that Walk() was unable to walk to this file or directory.
// The error is exposed this way so you can handle it appropriately.
// In this case, you return the error to the calling function,
// which effectively stops processing any other files.
func run(root string, out io.Writer, cfg config) error {

	return filepath.Walk(root,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// defines whether the current file or directory should be filtered out.
			// If so, the function returns nil, which skips the rest of the function
			// making Walk process the next file or directory.
			if filterOut(path, cfg.ext, cfg.size, info) {
				return nil
			}

			// if -list is set - dont do anything
			if cfg.list {
				return listFile(path, out)
			}

			// executing the action which, FOR NOW, is to list the name of the file
			// onscreen by calling the function listFile()

			// list is a default option, if nothing else was set
			return listFile(path, out)
		},
	)

}

func main() {
	root := flag.String("root", ".", "Root dir to start")
	list := flag.Bool("list", false, "List files only")
	ext := flag.String("ext", "", "File extension to filter out")
	size := flag.Int64("size", 0, "Minimum file size")
	flag.Parse()

	// instance of config
	c := config{
		ext:  *ext,
		size: *size,
		list: *list,
	}

	// check for errors and print error to stderr if any
	if err := run(*root, os.Stdout, c); err != nil {
		fmt.Println("error here")
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
