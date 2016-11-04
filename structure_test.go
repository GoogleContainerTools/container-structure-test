package main

import (
	"encoding/json"
	"errors"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/ghodss/yaml"
)

func TestAll(t *testing.T) {
	var err error
	var tests StructureTest
	for _, file := range configFiles {
		if tests, err = Parse(file); err != nil {
			log.Fatalf("Error parsing config file: %s", err)
		}
		log.Printf("Running tests for file %s", file)
		tests.RunAll(t)
	}
}

func compileAndRunRegex(regex string, base string, t *testing.T, err string, shouldMatch bool) {
	r, rErr := regexp.Compile(regex)
	if rErr != nil {
		t.Errorf("Error compiling regex %s : %s", regex, rErr.Error())
		return
	}
	if shouldMatch != r.MatchString(base) {
		t.Errorf(err)
	}
}

func Parse(fp string) (StructureTest, error) {
	testContents, err := ioutil.ReadFile(fp)
	if err != nil {
		return nil, err
	}

	var unmarshal Unmarshaller
	var versionHolder SchemaVersion

	switch {
	case strings.HasSuffix(fp, ".json"):
		unmarshal = json.Unmarshal
	case strings.HasSuffix(fp, ".yaml"):
		unmarshal = yaml.Unmarshal
	default:
		return nil, errors.New("Please provide valid JSON or YAML config file.")
	}

	if err := unmarshal(testContents, &versionHolder); err != nil {
		return nil, err
	}

	version := versionHolder.SchemaVersion
	if version == "" {
		return nil, errors.New("Please provide JSON schema version.")
	}
	st := schemaVersions[version]
	if st == nil {
		return nil, errors.New("Unsupported schema version: " + version)
	}
	unmarshal(testContents, st)
	tests, ok := st.(StructureTest) //type assertion
	if !ok {
		return nil, errors.New("Error encountered when type casting Structure Test interface!")
	}
	return tests, nil
}

var configFiles arrayFlags

func TestMain(m *testing.M) {
	flag.Var(&configFiles, "config", "path to the .yaml file containing test definitions.")
	flag.Parse()

	if len(configFiles) == 0 {
		configFiles = append(configFiles, "/workspace/structure_test.json")
	}

	if exit := m.Run(); exit != 0 {
		os.Exit(exit)
	}
}
