package data

import (
	"encoding/json"

	"github.com/ghodss/yaml"
)

type jsonParseFunc func(body []byte) ([]TestCaseData, error)

// Parse implements data.TestCaseParser
func (p jsonParseFunc) Parse(body []byte) ([]TestCaseData, error) {
	return p(body)
}

// jsonParse implements data.TestCaseParser
func jsonParse(body []byte) ([]TestCaseData, error) {
	cs := []TestCaseData{}
	if err := json.Unmarshal(body, &cs); err != nil {
		return nil, err
	}
	return cs, nil
}

// JSONParser returns a parser to parse bytes in json format
func JSONParser() TestCaseParser {
	return jsonParseFunc(jsonParse)
}

type yamlParseFunc func(body []byte) ([]TestCaseData, error)

// Parse implements data.TestCaseParser
func (p yamlParseFunc) Parse(body []byte) ([]TestCaseData, error) {
	return p(body)
}

func yamlParse(body []byte) ([]TestCaseData, error) {
	cs := []TestCaseData{}
	if err := yaml.Unmarshal(body, &cs); err != nil {
		return nil, err
	}
	return cs, nil
}

// YAMLParser returns a parser to parse bytes in yaml format
func YAMLParser() TestCaseParser {
	return yamlParseFunc(yamlParse)
}
