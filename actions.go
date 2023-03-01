package main

import (
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

// 1) preserve the relative directory tree so the files are archived
// in the same directories relative to the source root
// 2) compress the data
// destDir: The destination directory where the files will be archived.
// root: The root directory where the search was started
// use this value to determine the relative path of the files to archive
// so can create a similar directory tree in the destination directory
// path: The path of the file to be archived
// returns error, calling function can check its value and
// interrupt processing when issues occur
func archiveFile(destDir, root, path string) error {
	// 1) check if the argument destDir is a directory
	info, err := os.Stat(destDir)
	if err != nil {
		return err
	}

	if !info.IsDir() {
		return fmt.Errorf("%s id not directory", destDir)
	}

	// 2) determine the relative directory of the file to be archived
	// in relation to its source root path using the Rel() from filepath:
	relDir, err := filepath.Rel(root, filepath.Dir(path))
	if err != nil {
		return err
	}

	// 3) Create the new file name by adding the .gz suffix to the original file name
	// which obtained by calling the filepath.Base fn.
	// Define the target path by joining all 3 pieces together:
	// the destination directory, the relative directory, file name, filepath.Join()
	dest := fmt.Sprintf("%s.gz", filepath.Base(path))
	targetPath := filepath.Join(destDir, relDir, dest)

	// 4) create the target directory tree using os.MkdirAll:
	// os.MkdirAll function creates all the required directories at once
	// but will do nothing if the directories already exist,
	// which means you don’t have to write any additional checks
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return err
	}

	// Once have the target path, create the compressed archive.
	// To do this, use the io.Copy function to copy data
	// from the source file to the destination file.
	// But instead of using the destination file directly as an argument,
	// use the type gzip.Writer.
	// The gzip.Writer type implements the io.Writer interface,
	// which allows it to be used as an argument to any functions
	// that expect that interface as input, such as io.Copy,
	// but it writes the data in compressed form.
	out, err := os.OpenFile(targetPath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer out.Close()

	in, err := os.Open(path)
	if err != nil {
		return err
	}
	defer in.Close()

	zw := gzip.NewWriter(out)
	zw.Name = filepath.Base(path)

	if _, err := io.Copy(zw, in); err != nil {
		return err
	}

	if err := zw.Close(); err != nil {
		return err
	}

	return out.Close()

}

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
