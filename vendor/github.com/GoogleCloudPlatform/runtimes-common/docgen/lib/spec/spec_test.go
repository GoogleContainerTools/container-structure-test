package spec_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/GoogleCloudPlatform/runtimes-common/docgen/lib/spec"
	"github.com/ghodss/yaml"
	"github.com/golang/protobuf/jsonpb"
)

// Verifies that the YAML files under testdata can be converted to Document proto.
func TestYamlToProtoConversion(t *testing.T) {
	const testdataPath = "docgen/lib/spec/testdata"
	files, err := ioutil.ReadDir(testdataPath)
	if err != nil {
		t.Fatal(err)
	}

	count := 0
	for _, f := range files {
		if f.IsDir() || !strings.HasSuffix(f.Name(), ".yaml") {
			continue
		}
		count++

		document, err := spec.FromYamlFile(filepath.Join(testdataPath, f.Name()))
		if err != nil {
			t.Fatal(err)
		}

		// Converts proto to YAML for outputting.
		// TODO(huyhuynh): Add golden files.
		marshaler := &jsonpb.Marshaler{}
		var b bytes.Buffer
		marshaler.Marshal(&b, document)
		yamlContent, _ := yaml.JSONToYAML(b.Bytes())
		fmt.Printf("--- %s:\n%s\n", f.Name(), yamlContent)
	}

	if count <= 0 {
		t.Fatal("No .yaml files found in testdata")
	}
}

func TestMain(m *testing.M) {
	// If test is running under Bazel, Chdir to the working directory for data.
	path := os.ExpandEnv("${TEST_SRCDIR}/${TEST_WORKSPACE}")
	if path != "/" {
		os.Chdir(path)
	}
	os.Exit(m.Run())
}
