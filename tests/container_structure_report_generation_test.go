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
	"fmt"
	"io/ioutil"
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

func TestContainerStructureTestsTestReportGenerationWhenCmdSucceds(t *testing.T) {
	TestChannel = make(chan interface{}, 1)
	tmpfile, _ := ioutil.TempFile("", "test.json")
	cmd.RootCmd.SetArgs([]string{"test", "--image", TestImage, "--config", TestConfig,
		"--test-report", tmpfile.Name()})
	cmd.TestCmd.StreamO = streamOTest
	cmd.TestCmd.Stream = TestChannel
	GenerateSuccessResult = true
	err := ctc_lib.ExecuteE(cmd.RootCmd)
	if err != nil {
		fmt.Println("Unexpected error", err)
		t.Fail()
	}
	raw, err := ioutil.ReadFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("Error while reading the test report %v", err)
	}
	var c ctc_lib.ListCommandOutputObject
	json.Unmarshal(raw, &c)
	expectedMap := c.SummaryObject.(map[string]interface{})
	if expectedMap["Fail"].(float64) != 0 {
		t.Fatalf("Expected 0 Fail count. Got %d", expectedMap["Fail"])
	}
}

func TestContainerStructureTestsTestReportGenerationWhenCmdFails(t *testing.T) {
	cmd.RootCmd.ResetFlags()
	TestChannel = make(chan interface{}, 1)
	tmpfile, _ := ioutil.TempFile("", "test.json")
	cmd.RootCmd.SetArgs([]string{"test", "--image", TestImage, "--config", TestConfig,
		"--test-report", tmpfile.Name()})
	cmd.TestCmd.StreamO = streamOTest
	cmd.TestCmd.Stream = TestChannel
	GenerateSuccessResult = false // Generate Failed Test Results
	err := ctc_lib.ExecuteE(cmd.RootCmd)
	if err == nil {
		fmt.Println("Expected Command to fail but it Succeeded")
		t.Fail()
	}
	raw, err := ioutil.ReadFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("Error while reading the test report %v", err)
	}
	var c ctc_lib.ListCommandOutputObject
	json.Unmarshal(raw, &c)
	expectedMap := c.SummaryObject.(map[string]interface{})
	if expectedMap["Pass"].(float64) != 0 {
		t.Fatalf("Expected 0 Pass count. Got %d", expectedMap["Pass"])
	}
}

func TestContainerStructureTestsTestReportGenerationWhenTestReportFileInvalid(t *testing.T) {
	TestChannel = make(chan interface{}, 1)
	ctc_lib.SetExitOnError(false)
	defer ctc_lib.SetExitOnError(true)
	cmd.RootCmd.SetArgs([]string{"test", "--image", TestImage, "--config", TestConfig,
		"--test-report", "/invalid_dir/test.json"})
	cmd.TestCmd.StreamO = streamOTest
	cmd.TestCmd.Stream = TestChannel
	GenerateSuccessResult = true // Generate Failed Test Results
	cmd.RootCmd.ResetFlags()
	err := ctc_lib.ExecuteE(cmd.RootCmd)
	if err == nil {
		t.Fatal("Expected Command to fail but it Succeeded")
	}
}
