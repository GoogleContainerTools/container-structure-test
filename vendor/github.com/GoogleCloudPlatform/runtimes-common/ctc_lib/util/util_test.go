/*
Copyright 2018 Google Inc. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package util

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	log "github.com/sirupsen/logrus"
)

func TestIsDebugLevel(t *testing.T) {
	// Explicitly comparing boolean to make tests more readable.
	if IsDebug(log.DebugLevel) != true {
		t.Errorf("Expected True but returned False")
	}

	if IsDebug(log.FatalLevel) != false {
		t.Errorf("Expected False but returned True")
	}
}

func TestGetToolTempDirOrDefault(t *testing.T) {
	// Check when TempDir is passed.
	var tmpDir string
	defer os.Remove(tmpDir)
	tmpDir, _ = ioutil.TempDir("", "test")
	expectedDir := filepath.Join(tmpDir, "testTool")
	actualValue := GetToolTempDirOrDefault(tmpDir, "testTool")
	if expectedDir != actualValue {
		t.Errorf("Expected Error: \n %q \nGot:\n %q\n", expectedDir, actualValue)
	}
}

func TestGetToolTempDirOrDefaultWhenDefault(t *testing.T) {
	// When TempDir is not passed, system Os.TmpDir should be picked up.
	tmpDir := os.TempDir()
	expectedDir := filepath.Join(tmpDir, "testTool")
	defer os.Remove(expectedDir)
	actualValue := GetToolTempDirOrDefault("", "testTool")
	if expectedDir != actualValue {
		t.Errorf("Expected Error: \n %q \nGot:\n %q\n", expectedDir, actualValue)
	}
}
