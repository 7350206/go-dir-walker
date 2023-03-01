// test variations of integrate tests
// go executes all test cases for each test function,
// using the test name configured to present the results.
// This makes it easier to reference each test and troubleshoot them
// in case a test doesn’t pass.
package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	testCases := []struct {
		name     string
		root     string
		cfg      config
		expected string
	}{
		{name: "NoFilter", root: "testdata", cfg: config{ext: "", size: 0, list: true},
			expected: "testdata/dir.log\ntestdata/dir2/script.sh\n"},
		{name: "FilterExtensionMatch", root: "testdata",
			cfg:      config{ext: ".log", size: 0, list: true},
			expected: "testdata/dir.log\n"},
		{name: "FilterExtensionSizeMatch", root: "testdata",
			cfg:      config{ext: ".log", size: 10, list: true},
			expected: "testdata/dir.log\n"},
		{name: "FilterExtensionSizeNoMatch", root: "testdata",
			cfg:      config{ext: ".log", size: 20, list: true},
			expected: ""},
		{name: "FilterExtensionNoMatch", root: "testdata",
			cfg:      config{ext: ".gz", size: 0, list: true},
			expected: ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buffer bytes.Buffer

			if err := run(tc.root, &buffer, tc.cfg); err != nil {
				t.Fatal(err)
			}

			res := buffer.String()

			if tc.expected != res {
				t.Errorf("expected %q, got %q instead\n", tc.expected, res)
			}

		})
	}

}

func TestRunArchive(t *testing.T) {
	testCases := []struct {
		name         string
		cfg          config
		extNoArchive string
		nArchive     int
		nNoArchive   int
	}{
		{name: "ArchiveExtensionNoMatch",
			cfg:          config{ext: ".log"},
			extNoArchive: ".gz", nArchive: 0, nNoArchive: 10},
		{name: "ArchiveExtensionMatch",
			cfg:          config{ext: ".log"},
			extNoArchive: "", nArchive: 10, nNoArchive: 0},
		{name: "ArchiveExtensionMixed",
			cfg:          config{ext: ".log"},
			extNoArchive: ".gz", nArchive: 5, nNoArchive: 5},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// define a buffer variable to capture the output of the tool
			var buffer bytes.Buffer

			// in this case, use createTempDir to create both the origin directory
			// and the archiving directory.
			tempDir, cleanup := createTempDir(t, map[string]int{
				tc.cfg.ext:      tc.nArchive,
				tc.extNoArchive: tc.nNoArchive,
			})
			defer cleanup()

			// create the temporary archive directory using the helper function,
			// provide a value of nil as the file map input
			// since we don’t need any files in this directory.
			archiveDir, cleanupArchive := createTempDir(t, nil)
			defer cleanupArchive()

			// Assign the archiveDir variable containing the name of the archive dir
			// to the field tc.cfg.archive to be used as input for the function run.
			tc.cfg.archive = archiveDir

			// - If the run function returns an error, we fail the test
			// using t.Fata() from the testing type.
			// - Assuming the func completes successfully,
			// we validate the output content and the number of files archived.
			if err := run(tempDir, &buffer, tc.cfg); err != nil {
				t.Fatal(err)
			}

			// - The archiving feature outputs the name of each archived file,
			// so need a list of files that expect to be archived
			// to compare with the actual results.
			// Since the test creates the directory and files dynamically for each test,
			// you don’t have the name of the files beforehand.
			// - Create this list dynamically by reading the data from the temp dir.
			// Use the Glob() from the filepath package to find all file names
			// from the tempDir that match the archiving extension.
			// Use Join() from the filepath package to concatenate the pattern
			// with the temporary directory path:
			pattern := filepath.Join(tempDir, fmt.Sprintf("*%s", tc.cfg.ext))
			expFiles, err := filepath.Glob(pattern)
			if err != nil {
				t.Fatal(err)
			}

			// create the final list as a string to compare with the output,
			// use the strings.Join() to join each file path in the expFiles slice
			// with the newline character:
			expOut := strings.Join(expFiles, "\n")

			// Before comparing the two values,
			// remove the last new line from the output by using the strings.TrimSpace()
			// on the output variable buffer.
			// use the String() from the bytes.Buffer type to extract the content
			// of the buffer as a string.
			res := strings.TrimSpace(buffer.String())

			// compare the expected output expOut with the actual output res,
			// failing the test if they don’t match:
			if expOut != res {
				t.Errorf("expected %q, got %q instead\n", expOut, res)
			}

			// validate the number of files archived.
			// Start by reading the content of the temporary archive dir archiveDir,
			//  using the ReadDir() again.
			// Store the results into the slice filesArchived:
			filesArchived, err := ioutil.ReadDir(archiveDir)
			if err != nil {
				t.Fatal(err)
			}

			// compare the number of files archived with the expected number of files
			// that should be archived, tc.nArchive,
			// failing the test if they don’t match.
			// Use the len() to obtain the number of files in the filesArchived slice:
			if len(filesArchived) != tc.nArchive {
				t.Errorf("expected %d files archived, got %d files instead\n", filesArchived, tc.nArchive)
			}

		})
	}

}

func TestRunDelExtension(t *testing.T) {
	testCases := []struct {
		name        string
		cfg         config
		extNoDelete string
		nDelete     int
		nNoDelete   int
		expected    string
	}{
		{name: "DeleteExtensionNoMatch",
			cfg:         config{ext: ".log", del: true},
			extNoDelete: ".gz", nDelete: 0, nNoDelete: 10,
			expected: ""},
		{name: "DeleteExtensionMatch",
			cfg:         config{ext: ".log", del: true},
			extNoDelete: "", nDelete: 10, nNoDelete: 0,
			expected: ""},
		{name: "DeleteExtensionMixed",
			cfg:         config{ext: ".log", del: true},
			extNoDelete: ".gz", nDelete: 5, nNoDelete: 5,
			expected: ""},
	}

	// iterate over test cases
	// The main difference is that in this case, will call the helper function
	// to create the temporary directory and files

	// exec runDel test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var (
				buffer    bytes.Buffer
				logBuffer bytes.Buffer
			)

			tc.cfg.wLog = &logBuffer

			tempDir, cleanup := createTempDir(t, map[string]int{
				tc.cfg.ext:     tc.nDelete,
				tc.extNoDelete: tc.nNoDelete,
			})
			// This ensures that it gets executed at the end of the test,
			// cleaning up the temporary directory.
			defer cleanup()

			if err := run(tempDir, &buffer, tc.cfg); err != nil {
				t.Fatal(err)
			}

			res := buffer.String()

			if tc.expected != res {
				t.Errorf("expected %q, got %q instead\n", tc.expected, res)
			}

			// Finally, read the files that were left in the directory
			// after the delete by using the ioutil.ReadDir func on the temp test dir.
			// Compare the number of files left with the expected number,
			// failing the test if they don’t match.
			filesLeft, err := ioutil.ReadDir(tempDir)
			if err != nil {
				t.Error(err)
			}
			if len(filesLeft) != tc.nNoDelete {
				t.Errorf("expected %d filesleft, got %d instead\n", tc.nNoDelete, len(filesLeft))
			}

			// verify the log output.
			// since fn adds a log line for each deleted file,
			// can count the number of lines in the log output
			// and compare it to the number of deleted files
			// plus one for the final new line added to the end.
			// If they don’t match, the test fails.
			// To count the lines, use the bytes.Split fn passing \n as an argument.
			// This fn outputs a slice so use the built-in len function to get length.
			expLogLines := tc.nDelete + 1
			lines := bytes.Split(logBuffer.Bytes(), []byte("\n"))
			if len(lines) != expLogLines {
				t.Errorf("expected %d log lines, got %d instead\n", expLogLines, len(lines))
			}

		})
	}

}

// helper takes:
// 1) pointer of type *testing.T
// 2) files of type map[string]int for defining the number of files
// this function will create for each extension.
// returns two values:
// the directory name of the created directory, so you can use it during testing,
// the cleanup function cleanup of type func().
func createTempDir(t *testing.T, files map[string]int) (dirname string, cleanup func()) {

	// Mark this function as a test helper by calling the t.Helper method:
	t.Helper()

	// create the temporary directory using the ioutil.TempDir function,
	// with the prefix walktest
	tempDir, err := ioutil.TempDir("", "walktest")
	if err != nil {
		t.Fatal(err)
	}
	// Iterate over the files map, creating the specified number of dummy files
	// for each provided extension:
	for k, n := range files {
		for j := 1; j <= n; j++ {
			fname := fmt.Sprintf("file%d%s", j, k)
			fpath := filepath.Join(tempDir, fname)
			if err := ioutil.WriteFile(fpath, []byte("dummy"), 644); err != nil {
				t.Fatal(err)
			}
		}
	}
	// return the temporary directory name tempDir and an anonymous function
	// which when called, executes os.RemoveAll to remove the temp directory.
	return tempDir, func() { os.RemoveAll(tempDir) }

}
