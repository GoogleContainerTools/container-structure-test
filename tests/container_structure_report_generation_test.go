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
package tests

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/GoogleCloudPlatform/runtimes-common/ctc_lib"
	"github.com/GoogleContainerTools/container-structure-test/cmd"
	types "github.com/GoogleContainerTools/container-structure-test/pkg/types/unversioned"
	"github.com/spf13/cobra"
)

var TestImage = "testImage"
var TestConfig = "dummy"
var GenerateSuccessResult = true
var TestChannel chan interface{}

func addResults(b bool) {
	result := &types.TestResult{
		Name:   "TestResult Test",
		Pass:   true,
		Errors: make([]string, 0),
	}
	if !b {
		result.Pass = false
		result.Errors = append(result.Errors, "Test Error")
	}
	TestChannel <- result
	close(TestChannel)
}

func streamOTest(command *cobra.Command, args []string) {
	go addResults(GenerateSuccessResult)
}

func IsErrorEqual(err1 error, err2 error) bool {
	if err1 == nil && err2 == nil {
		return true // Return true Both of them are nil
	} else if err1 == nil || err2 == nil {
		return false // Return false if either of them are nil
	} else if err1.Error() == err2.Error() {
		return true // Return true if Messages are equal
	}
	return false // Return false
}

var testCases = []struct {
	name                  string
	generateSuccessResult bool
	expectedCommandError  error
	expectedOutput        map[string]interface{}
	testfile              string
	checkReport           bool
}{
	{"reportSuccess", true, nil, map[string]interface{}{"Pass": 1.0, "Fail": 0.0, "Total": 1.0}, "", true},
	{"reportFail", false, errors.New("Test(s) FAIL"), map[string]interface{}{"Pass": 0.0, "Fail": 1.0, "Total": 1.0}, "", true},
	{"InvalidReportFile", false, errors.New("open /invalid_dir/file: no such file or directory"), nil, "/invalid_dir/file", false},
}

func TestContainerStructureTestsTestReportGeneration(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			TestChannel = make(chan interface{}, 1)
			if tc.testfile == "" {
				tmpfile, _ := ioutil.TempFile("", "test.json")
				defer os.Remove(tmpfile.Name())
				tc.testfile = tmpfile.Name()
			}
			cmd.RootCmd.SetArgs([]string{"test", "--image", TestImage, "--config", TestConfig,
				"--test-report", tc.testfile})
			cmd.TestCmd.StreamO = streamOTest
			cmd.TestCmd.Stream = TestChannel
			GenerateSuccessResult = tc.generateSuccessResult
			cmd.RootCmd.ResetFlags()
			err := ctc_lib.ExecuteE(cmd.RootCmd)
			if !IsErrorEqual(err, tc.expectedCommandError) {
				t.Fatalf("Expected error %v\nGot %v\n", tc.expectedCommandError, err)
			}
			if tc.checkReport { //Read results if checkReport set to True
				raw, err := ioutil.ReadFile(tc.testfile)
				if err != nil {
					t.Fatalf("Error while reading the test report %v", err)
				}
				c := ctc_lib.ListCommandOutputObject{}
				json.Unmarshal(raw, &c)
				actualMap := c.SummaryObject.(map[string]interface{})
				if !reflect.DeepEqual(actualMap, tc.expectedOutput) {
					t.Fatalf("Expected %v.\nGot %v", tc.expectedOutput, actualMap)
				}
			}
		})
	}
}
