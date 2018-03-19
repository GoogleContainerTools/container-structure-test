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
)

type EnvVar struct {
	Key   string
	Value string
}

type Config struct {
	Env          map[string]string
	Entrypoint   []string
	Cmd          []string
	Volumes      []string
	Workdir      string
	ExposedPorts []string
}

type FlattenedConfig struct {
	Env          []string            `json:"Env"`
	Entrypoint   []string            `json:"Entrypoint"`
	Cmd          []string            `json:"Cmd"`
	Volumes      map[string]string   `json:"Volumes"`
	Workdir      string              `json:"WorkingDir"`
	ExposedPorts map[string][]string `json:"ExposedPorts"`
}

type FlattenedMetadata struct {
	Config FlattenedConfig `json:"config"`
}

type FullResult struct {
	FileName string
	Results  []*TestResult
}

type TestResult struct {
	Name   string
	Pass   bool
	Stdout string
	Stderr string
	Errors []string
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
