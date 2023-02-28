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
