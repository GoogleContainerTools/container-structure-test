// Copyright 2018 Google Inc. All rights reserved.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package unversioned

import (
	"fmt"
	"strings"
)

type EnvVar struct {
	Key     string
	Value   string
	IsRegex bool `yaml:"isRegex"`
}

type Label struct {
	Key     string
	Value   string
	IsRegex bool `yaml:"isRegex"`
}

type Config struct {
	Env          map[string]string
	Entrypoint   []string
	Cmd          []string
	Volumes      []string
	Workdir      string
	ExposedPorts []string
	Labels       map[string]string
}

type TestResult struct {
	Name   string
	Pass   bool
	Stdout string   `json:",omitempty"`
	Stderr string   `json:",omitempty"`
	Errors []string `json:",omitempty"`
}

func (t *TestResult) String() string {
	strRepr := fmt.Sprintf("\nTest Name:%s", t.Name)
	testStatus := "Fail"
	if t.IsPass() {
		testStatus = "Pass"
	}
	strRepr += fmt.Sprintf("\nTest Status:%s", testStatus)
	if t.Stdout != "" {
		strRepr += fmt.Sprintf("\nStdout:%s", t.Stdout)
	}
	if t.Stderr != "" {
		strRepr += fmt.Sprintf("\nStderr:%s", t.Stderr)
	}
	strRepr += fmt.Sprintf("\nErrors:%s\n", strings.Join(t.Errors, ","))
	return strRepr
}

func (t *TestResult) Error(s string) {
	t.Errors = append(t.Errors, s)
}

func (t *TestResult) Errorf(s string, args ...interface{}) {
	t.Errors = append(t.Errors, fmt.Sprintf(s, args...))
}

func (t *TestResult) Fail() {
	t.Pass = false
}

func (t *TestResult) IsPass() bool {
	return t.Pass
}

type SummaryObject struct {
	Pass    int
	Fail    int
	Total   int
	Results []*TestResult `json:",omitempty"`
}
