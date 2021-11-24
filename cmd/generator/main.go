package main

import (
	"log"
	"net/http"

	"github.com/shurcooL/vfsgen"
)

func main() {
	var template = http.Dir("tmpl")

	err := vfsgen.Generate(template, vfsgen.Options{
		PackageName:  "template",
		VariableName: "TemplateSet",
	})
	if err != nil {
		log.Fatalln(err)
	}
}
