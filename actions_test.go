package main

import (
	"os"
	"testing"
)

func TestFilterOut(t *testing.T) {
	// add anon slice of struct with the definition of the test cases.
	// struct fields will represent the values used for each test
	// such as the test’s name, file to read, extension to filter,
	// minimum file size, and the expected test result:

	testCases := []struct {
		name     string
		file     string
		ext      string
		minSize  int64
		expected bool
	}{
		//each element -- test case
		// name is “FilterNoExtension”.
		// This uses the file testdata/dir.log,
		// the extension to filter is blank,
		// the minimum size is zero, and we expect this test to return false
		{"FilterNoExtension", "testdata/dir.log", "", 0, false},
		{"FilterExtensionMatch", "testdata/dir.log", ".log", 0, false},
		{"FilterExtensionNoMatch", "testdata/dir.log", ".sh", 0, true},
		{"FilterExtensionSizeMatch", "testdata/dir.log", ".log", 10, false},
		{"FilterExtensionSizeNoMatch", "testdata/dir.log", ".log", 20, true},
	}

	// iterare thr cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// retrieve file attributes
			info, err := os.Stat(tc.file)
			if err != nil {
				t.Fatal(err)
			}

			// provide that attrs as info
			f := filterOut(tc.file, tc.ext, tc.minSize, info)

			// compare results
			if f != tc.expected {
				t.Errorf("expected '%t', got '%t' instead", tc.expected, f)
			}
		})
	}

}
