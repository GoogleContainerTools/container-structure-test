package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/GoogleCloudPlatform/runtimes-common/docgen/lib/render"
	"github.com/GoogleCloudPlatform/runtimes-common/docgen/lib/spec"
)

var spec_file = flag.String("spec_file", "", "Path to the file containing the spec")

func main() {
	flag.Parse()

	if *spec_file == "" {
		log.Fatal("--spec_file must be specified")
	}

	doc, err := spec.FromYamlFile(*spec_file)
	check(err)

	out, err := render.Render(doc)
	check(err)
	fmt.Println(string(out))
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
