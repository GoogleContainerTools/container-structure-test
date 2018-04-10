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
package ctc_lib

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/spf13/cobra"
)

func TestConfigCommandGet(t *testing.T) {
	var testConfigCommand = ContainerToolCommand{
		ContainerToolCommandBase: &ContainerToolCommandBase{
			Command: &cobra.Command{
				Use: "Hello Command",
			},
			Phase: "test",
		},
		Output: nil,
		RunO: func(command *cobra.Command, args []string) (interface{}, error) {
			return nil, nil
		},
	}

	Version = "1.0.1"
	ConfigFile = "testdata/testConfig.json"
	testConfigCommand.SetArgs([]string{"config"})
	Execute(&testConfigCommand)
	expectedConfig := &ConfigOutput{
		Config: map[string]interface{}{
			"message":      "echo", // Make sure user defined config are also returned
			"update_check": "true", // inhertited from the Default Config
			"logdir":       "/tmp", // This overrides the default Config
		},
	}
	if reflect.DeepEqual(ConfigCommand.Output, expectedConfig) {
		t.Errorf("Expected to contain: \n %v\nGot:\n %v\n", expectedConfig, ConfigCommand.Output)
	}
}

func TestConfigCommandSet(t *testing.T) {
	var testConfigCommand = ContainerToolCommand{
		ContainerToolCommandBase: &ContainerToolCommandBase{
			Command: &cobra.Command{
				Use: "Hello Command",
			},
			Phase: "test",
		},
		Output: nil,
		RunO: func(command *cobra.Command, args []string) (interface{}, error) {
			return nil, nil
		},
	}

	tmpDir, _ := ioutil.TempDir("", "tests")
	defer os.RemoveAll(tmpDir)

	Version = "1.0.1"
	ConfigFile = filepath.Join(tmpDir, "testConfig.json")
	jsonConfigData, _ := json.Marshal(map[string]interface{}{
		"logdir":  "/tmp",
		"message": "echo",
	})
	ioutil.WriteFile(ConfigFile, jsonConfigData, 0644)

	testConfigCommand.SetArgs([]string{"config", "set", "message", "hi"})
	Execute(&testConfigCommand)

	expectedConfig := map[string]interface{}{
		"logdir":  "/tmp", // This overrides the default Config
		"message": "hi",   // Make sure user defined config are also returned
	}
	//Actual Config
	var actualConfig map[string]interface{}
	raw, err := ioutil.ReadFile(ConfigFile)
	if err != nil {
		t.Errorf("Error Reading Test Config File %s", ConfigFile)
	}
	json.Unmarshal(raw, &actualConfig)
	if !reflect.DeepEqual(actualConfig, expectedConfig) {
		t.Errorf("Expected to contain: \n %v\nGot:\n %v\n", expectedConfig, actualConfig)
	}
}
