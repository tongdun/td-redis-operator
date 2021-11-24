package tester

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/liubog2008/tester/pkg/data"
)

const defaultTestDataDir = "testdata"

// TestFunc defines function of main testing logic
type TestFunc func(t *testing.T, c data.TestCase)

// Test parses testdata and runs test case one by one
func Test(t *testing.T, f TestFunc) {
	internalTest(t, f, nil, defaultTestDataDir)
}

// TestWithSelector select matched testdata and runs them
func TestWithSelector(t *testing.T, f TestFunc, selector map[string]string) {
	internalTest(t, f, selector, defaultTestDataDir)
}

func internalTest(t *testing.T, f TestFunc, selector map[string]string, dataDir string) {
	pc, file, _, ok := runtime.Caller(2)
	if !ok {
		t.Fatalf("can't find caller of tester")
	}
	dir := filepath.Dir(file)
	base := filepath.Base(file)
	ext := filepath.Ext(base)
	// 1. trim extension suffix
	// 2. trim _test suffix
	base = strings.TrimSuffix(strings.TrimSuffix(base, ext), "_test")
	caller := runtime.FuncForPC(pc)
	ss := strings.Split(caller.Name(), ".")
	callerName := ss[len(ss)-1]

	if !filepath.IsAbs(dataDir) {
		dataDir = filepath.Join(dir, dataDir)
	}

	// TODO(liubog2008): make testdata can be configured
	files, err := findFiles(dataDir, base, callerName)
	if err != nil {
		t.Fatalf("can't find test data: %v", err)
	}

	tcl, err := data.NewTestCaseList(files...)
	if err != nil {
		t.Fatalf("can't parse test data from files: %v", err)
	}

	tcs := tcl.Select(selector)
	for _, tc := range tcs {
		f(t, tc)
	}
}

// findFile finds testdata from pkg dir
// e.g.
//      dataDir: /go/github.com/src/xxx/yyy/testdata
//     fileName: echo
//   callerName: TestEcho
// There are two optional way to find testdata files
//   1. First find file with json or yaml extension, it will be
//      /go/github.com/src/xxx/yyy/testdata/echo/TestEcho.[yaml|json]
//   2. if not find, all files in dir
//      /go/github.com/src/xxx/yyy/testdata/echo/TestEcho
//      will be returned
func findFiles(dataDir, fileName, callerName string) ([]string, error) {
	filePrefix := filepath.Join(dataDir, fileName, callerName)
	matchedFiles, err := filepath.Glob(filePrefix + ".*")
	if err != nil {
		return nil, err
	}
	if len(matchedFiles) > 1 {
		return nil, fmt.Errorf("find more than one matched file: %v", matchedFiles[0])
	}
	if len(matchedFiles) == 1 {
		return matchedFiles, nil
	}
	all := filepath.Join(filePrefix, "*")
	filesInDir, err := filepath.Glob(all)
	if err != nil {
		return nil, err
	}
	if len(filesInDir) == 0 {
		return nil, fmt.Errorf("can't find any files in %v", all)
	}
	return filesInDir, nil
}
