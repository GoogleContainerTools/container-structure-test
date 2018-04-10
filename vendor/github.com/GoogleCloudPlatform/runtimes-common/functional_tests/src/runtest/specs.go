// Copyright 2017 Google Inc. All rights reserved.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

type Suite struct {
	Setup    []Setup
	Teardown []Teardown
	Target   string
	Tests    []Test
}

type Setup struct {
	Command []string
}

type Teardown struct {
	Command []string
}

type Test struct {
	Name    string
	Command []string
	Expect  Expect
}

type Expect struct {
	Stdout StringAssert
	Stderr StringAssert
}

type StringAssert struct {
	Exactly     string
	Equals      string
	Matches     string
	MustBeEmpty bool `yaml:"mustBeEmpty"`
}

func LoadSuite(path string) Suite {
	data, err := ioutil.ReadFile(path)
	check(err)

	if strings.HasSuffix(path, ".json") {
		return loadJsonSuite(data)
	} else if strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") {
		return loadYamlSuite(data)
	}
	log.Fatalf("Unrecognized test suite file type: %v", path)
	return Suite{}
}

func loadJsonSuite(data []byte) Suite {
	suite := Suite{}
	err := json.Unmarshal(data, &suite)
	check(err)
	return suite
}

func loadYamlSuite(data []byte) Suite {
	suite := Suite{}
	err := yaml.Unmarshal(data, &suite)
	check(err)
	return suite
}

func check(err interface{}) {
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
}
