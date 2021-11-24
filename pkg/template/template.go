//go:generate go run ../../cmd/generator/main.go

// Package template defines template
package template

import (
	"bytes"
	"io/ioutil"
	"text/template"

	"github.com/ghodss/yaml"
)

func NewTemplate(s string) (*Template, error) {
	f, err := TemplateSet.Open(s)
	if err != nil {
		return nil, err
	}
	content, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	t, err := template.New(s).Parse(string(content))
	if err != nil {
		return nil, err
	}
	return &Template{
		t: t,
	}, nil
}

type Template struct {
	t *template.Template
}

func (t *Template) Execute(data interface{}, obj interface{}) error {
	var buf bytes.Buffer
	if err := t.t.Execute(&buf, data); err != nil {
		return err
	}
	if err := yaml.Unmarshal(buf.Bytes(), obj); err != nil {
		return err
	}
	return nil
}
