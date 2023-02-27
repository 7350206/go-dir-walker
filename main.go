package main

import (
	"flag"
	"fmt"
	"io" // use io.Writer
	"log"
	"os"
	"path/filepath"
)

var (
	f   = os.Stdout
	err error
)

// list of params to run will be long,
// provide some of these args packaged in a custom type.
type config struct {
	ext  string // extension to filter out
	size int64  // min file size
	list bool   // list files
	del  bool
	// make code flexible, accepting a file in the main program
	// or a buffer that can use while testing the tool.
	wLog io.Writer // log destination
}

// descend dir by root and find all diles and sub-dirs on it
// checking to see if the provided error is not nil,
// which means that Walk() was unable to walk to this file or directory.
// The error is exposed this way so you can handle it appropriately.
// In this case, you return the error to the calling function,
// which effectively stops processing any other files.
func run(root string, out io.Writer, cfg config) error {

	// new instance of log.Logger by using log.New() from the log package
	// log.Logger instance to log deleted files to the provided io.Writer
	// interface instance cfg.wLog.
	delLogger := log.New(cfg.wLog, "DELETED FILE:", log.LstdFlags)

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

			// check if the variable cfg.del is set, and, if so,
			// call the delFile function to delete the file.
			if cfg.del {
				return delFile(path, delLogger)
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
	del := flag.Bool("del", false, "Delete files")

	// default is "", so if not specify - send output to STDOUT
	logFile := flag.String("log", "", "Log deleted to that file")

	flag.Parse()

	// instance of config
	c := config{
		ext:  *ext,
		size: *size,
		list: *list,
		del:  *del, // mapping the field del to the flag value so that it’s passed to run()
		wLog: f,
	}

	// check for errors and print error to stderr if any
	if err := run(*root, os.Stdout, c); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// we aren’t able to test it
	// But it allows us to have the run() receive an io.Writer interface
	// which makes it easier to test the logging functionality.
	// This is a good trade-off since this block of code is opening a file
	// using the std lib which has already been tested by the Go team.
	if *logFile != "" {
		// os.OpenFile function returns a value f of type os.File
		// that implements the io.Writer interface,
		// which means you can use it as the value for the wLog field
		// in the config struct.
		f, err := os.OpenFile(*logFile, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		defer f.Close()
	}

}
