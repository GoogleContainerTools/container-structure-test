package spec

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/GoogleCloudPlatform/runtimes-common/docgen/lib/proto"
	"github.com/ghodss/yaml"
	"github.com/golang/protobuf/jsonpb"
)

// TODO: Add validations for proto fields.

// FromYamlFile parses a YAML file into Document pb.
func FromYamlFile(filename string) (*proto.Document, error) {
	bytesData, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return FromYamlBytes(bytesData)
}

// FromYamlBytes parses a YAML content into Document pb.
func FromYamlBytes(bytesData []byte) (doc *proto.Document, err error) {
	doc = &proto.Document{}

	jsonData, err := yaml.YAMLToJSON(bytesData)
	if err != nil {
		return
	}
	jsonData = cleanUpJson(jsonData)

	// Converts JSON to proto.
	unmarshaler := &jsonpb.Unmarshaler{}
	err = unmarshaler.Unmarshal(bytes.NewBuffer(jsonData), doc)

	return doc, err
}

// cleanUpJson removes top level fields with names starting with underscores.
// These fields are intended as helpers in the YAML file.
// We have to do this because the unmarshaling to pb is strict on
// unknown fields.
func cleanUpJson(jsonData []byte) []byte {
	var doc interface{}
	json.Unmarshal(jsonData, &doc)

	switch d := doc.(type) {
	case map[string]interface{}:
		result := map[string]interface{}{}
		for key, value := range d {
			if !strings.HasPrefix(key, "_") {
				result[key] = value
			}
		}
		doc = result
	}

	r, err := json.Marshal(&doc)
	if err != nil {
		panic(fmt.Sprintf("Unexpected failure when marshaling back to JSON: %v", err))
	}
	return r
}
