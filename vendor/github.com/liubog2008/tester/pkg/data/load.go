package data

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/ghodss/yaml"
)

// testCase implements TestCase
type testCase struct {
	format TestCaseFileFormat
	data   TestCaseData
}

// Description implements data.TestCase
func (c *testCase) Description() string {
	return c.data.Description
}

// Unmarshal implements data.TestCase
func (c *testCase) Unmarshal(obj interface{}) error {
	switch c.format {
	case JSONFormat:
		return json.Unmarshal([]byte(c.data.Data), obj)
	case YAMLFormat:
		return yaml.Unmarshal([]byte(c.data.Data), obj)
	}
	return fmt.Errorf("unrecognized format: %v, only support json and yaml now", c.format)
}

// Match implements data.TestCase
func (c *testCase) Match(labels map[string]string) bool {
	return contains(c.data.Labels, labels)
}

type testCaseList struct {
	items []TestCase
}

// NewTestCaseList parses multiple files and returns list of test case
func NewTestCaseList(files ...string) (TestCaseList, error) {
	var parser TestCaseParser
	cl := &testCaseList{}
	for _, file := range files {
		ext := TestCaseFileFormat(filepath.Ext(file))
		switch ext {
		case JSONFormat:
			parser = JSONParser()
		case YAMLFormat:
			parser = YAMLParser()
		default:
			return nil, fmt.Errorf("can't find parser for %v", ext)
		}
		body, err := ioutil.ReadFile(filepath.Clean(file))
		if err != nil {
			return nil, err
		}
		cs, err := parser.Parse(body)
		if err != nil {
			return nil, err
		}
		for _, c := range cs {
			cl.items = append(cl.items, &testCase{
				format: ext,
				data:   c,
			})
		}
	}

	return cl, nil
}

func (cl *testCaseList) Select(labels map[string]string) []TestCase {
	cs := []TestCase{}
	for _, item := range cl.items {
		if item.Match(labels) {
			cs = append(cs, item)
		}
	}
	return cs
}

// if all KVs in b are also in a, return true
func contains(a, b map[string]string) bool {
	for k, v := range b {
		if a == nil {
			return false
		}
		av, ok := a[k]
		if !ok {
			return false
		}
		if av != v {
			return false
		}
	}
	return true
}
