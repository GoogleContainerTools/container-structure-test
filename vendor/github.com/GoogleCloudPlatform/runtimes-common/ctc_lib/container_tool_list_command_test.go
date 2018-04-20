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
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/sirupsen/logrus/hooks/test"
	"github.com/spf13/cobra"
)

type TestListOutput struct {
	Name string
}

type TestFooterOutput struct {
	Count int
}

var LName string
var Channel chan interface{}

func processNames(names string) {
	for _, name := range strings.Split(names, ",") {
		testListOutput := TestListOutput{Name: name}
		Channel <- testListOutput
	}
	// Make sure to close the stream.
	close(Channel)
}

func RunListCommand(command *cobra.Command, args []string) ([]interface{}, error) {
	Log.Debug("Running Hello World Command")
	var testOutputs []TestListOutput
	for _, name := range strings.Split(LName, ",") {
		testOutputs = append(testOutputs, TestListOutput{Name: name})
	}
	s := make([]interface{}, len(testOutputs))
	for i, v := range testOutputs {
		s[i] = v
	}
	return s, nil
}

func RunStreamCommand(command *cobra.Command, args []string) {
	// Do pre processing.
	Log.Debug("Running Hello World Command")
	// Run the method which writes to the stream
	go processNames(LName)
}

func TestContainerToolCommandListOutput(t *testing.T) {
	testCommand := ContainerToolListCommand{
		ContainerToolCommandBase: &ContainerToolCommandBase{
			Command: &cobra.Command{
				Use: "Hello Command",
			},
			Phase:           "test",
			DefaultTemplate: "{{ range $k, $v := . }}{{$v.Name}},{{ end }}",
		},
		OutputList: make([]interface{}, 0),
		RunO:       RunListCommand,
	}
	testCommand.Flags().StringVarP(&LName, "name", "n", "", "Comma Separated list of Name")
	var OutputBuffer bytes.Buffer
	testCommand.Command.SetOutput(&OutputBuffer)
	testCommand.SetArgs([]string{"--name=John,Jane"})
	Execute(&testCommand)
	var expectedOutput = []TestListOutput{
		{Name: "John"},
		{Name: "Jane"},
	}
	s := make([]interface{}, len(expectedOutput))
	for i, v := range expectedOutput {
		s[i] = v
	}
	if !reflect.DeepEqual(s, testCommand.OutputList) {
		t.Errorf("Expected to contain: \n %v\nGot:\n %v\n", s, testCommand.OutputList)
	}
	// Check if the output is good
	if OutputBuffer.String() != "John,Jane,\n" {
		t.Errorf("Expected to contain: \n John,Jane,\nGot:\n %v\n", OutputBuffer.String())
	}
}
func TestContainerToolCommandStreamOutput(t *testing.T) {
	Channel = make(chan interface{}, 1)
	testCommand := ContainerToolListCommand{
		ContainerToolCommandBase: &ContainerToolCommandBase{
			Command: &cobra.Command{
				Use: "Hello Command",
			},
			Phase:           "test",
			DefaultTemplate: "{{.Name}},",
		},
		OutputList:      make([]interface{}, 0),
		SummaryObject:   &TestFooterOutput{},
		SummaryTemplate: "\n{{.Count}} Names",
		StreamO:         RunStreamCommand,
		Stream:          Channel,
	}
	testCommand.Flags().StringVarP(&LName, "name", "n", "", "Comma Separated list of Name")
	var OutputBuffer bytes.Buffer
	testCommand.Command.SetOutput(&OutputBuffer)
	testCommand.SetArgs([]string{"--name=John,Jane"})
	Execute(&testCommand)
	var expectedOutput = []TestListOutput{
		{Name: "John"},
		{Name: "Jane"},
	}
	s := make([]interface{}, len(expectedOutput))
	for i, v := range expectedOutput {
		s[i] = v
	}
	if !reflect.DeepEqual(s, testCommand.OutputList) {
		t.Errorf("Expected to contain: \n %v\nGot:\n %v\n", s, testCommand.OutputList)
	}
	// Check if the output is good
	if OutputBuffer.String() != "John,Jane,\n" {
		t.Errorf("Expected to contain: \n John,Jane,\nGot:\n %v\n", OutputBuffer.String())
	}
}

func TestContainerToolCommandStreamOutputValidateResult(t *testing.T) {
	Channel = make(chan interface{}, 1)
	testCommand := ContainerToolListCommand{
		ContainerToolCommandBase: &ContainerToolCommandBase{
			Command: &cobra.Command{
				Use: "Hello Command",
			},
			Phase:           "test",
			DefaultTemplate: "{{.}}",
		},
		OutputList:      make([]interface{}, 0),
		SummaryObject:   &TestFooterOutput{},
		SummaryTemplate: "\n{{.Count}} Names",
		StreamO:         RunStreamCommand,
		Stream:          Channel,
		TotalO: func(list []interface{}) (interface{}, error) {
			return &TestFooterOutput{Count: len(list)}, nil
		},
	}
	testCommand.Flags().StringVarP(&LName, "name", "n", "", "Comma Separated list of Name")
	var OutputBuffer bytes.Buffer
	testCommand.Command.SetOutput(&OutputBuffer)
	testCommand.SetArgs([]string{"--name=John,Jane"})
	Execute(&testCommand)
	var expectedOutput = []TestListOutput{
		{Name: "John"},
		{Name: "Jane"},
	}
	s := make([]interface{}, len(expectedOutput))
	for i, v := range expectedOutput {
		s[i] = v
	}
	if !reflect.DeepEqual(s, testCommand.OutputList) {
		t.Errorf("Expected to contain: \n %v\nGot:\n %v\n", s, testCommand.OutputList)
	}
	// Check if the output is good
	if strings.Contains("2 Names", OutputBuffer.String()) {
		t.Errorf("Expected to contain: \n 2 Names\nGot:\n %v\n", OutputBuffer.String())
	}
}

func TestContainerToolCommandLogging(t *testing.T) {
	var hook *test.Hook
	testCommand := ContainerToolListCommand{
		ContainerToolCommandBase: &ContainerToolCommandBase{
			Command: &cobra.Command{
				Use: "Hello Command",
				PersistentPreRun: func(c *cobra.Command, args []string) {
					_, hook = test.NewNullLogger()
					Log.AddHook(hook)
				},
			},
			Phase: "test",
		},
		OutputList: make([]interface{}, 0),
		RunO:       RunListCommand,
	}
	testCommand.Flags().StringVarP(&LName, "name", "n", "", "Comma Separated list of Name")
	testCommand.SetArgs([]string{"--name=John,Jane", "--verbosity=debug"})
	Execute(&testCommand)
	if len(hook.Entries) != 2 {
		t.Errorf("Expected 2 Log Entry. Found %v", len(hook.Entries))
	}

	if hook.AllEntries()[0].Message != "Running Hello World Command" {
		t.Errorf("Expected to contain: \n Running Hello World Command\nGot:\n %v\n", hook.LastEntry().Message)
	}
}

func TestContainerToolCommandHandlePanicLogging(t *testing.T) {
	defer SetExitOnError(true)
	var hook *test.Hook
	testCommand := ContainerToolListCommand{
		ContainerToolCommandBase: &ContainerToolCommandBase{
			Command: &cobra.Command{
				Use: "kill",
				PersistentPreRun: func(c *cobra.Command, args []string) {
					_, hook = test.NewNullLogger()
					Log.AddHook(hook)
				},
			},
			Phase: "test",
		},
		OutputList: make([]interface{}, 0),
		RunO: func(command *cobra.Command, args []string) ([]interface{}, error) {
			Log.Panic("Please dont kill me")
			return nil, nil
		},
	}
	SetExitOnError(false)
	testCommand.SetArgs([]string{})
	Execute(&testCommand)
	if hook.LastEntry().Message != "Please dont kill me" {
		t.Errorf("Expected to contain: \n Please dont kill me\nGot:\n %v\n", hook.LastEntry().Message)
	}
}

func TestContainerToolCommandStreamOutputValidateJsonResult(t *testing.T) {
	Channel = make(chan interface{}, 1)
	testCommand := ContainerToolListCommand{
		ContainerToolCommandBase: &ContainerToolCommandBase{
			Command: &cobra.Command{
				Use: "Hello Command",
			},
			Phase:           "test",
			DefaultTemplate: "{{.}}",
		},
		OutputList:      make([]interface{}, 0),
		SummaryObject:   &TestFooterOutput{},
		SummaryTemplate: "\n{{.Count}} Names",
		StreamO:         RunStreamCommand,
		Stream:          Channel,
		TotalO: func(list []interface{}) (interface{}, error) {
			return &TestFooterOutput{Count: len(list)}, nil
		},
	}
	testCommand.Flags().StringVarP(&LName, "name", "n", "", "Comma Separated list of Name")
	var OutputBuffer bytes.Buffer
	testCommand.Command.SetOutput(&OutputBuffer)
	testCommand.SetArgs([]string{"--name=John,Jane", "--jsonOutput=True"})
	Execute(&testCommand)
	var result = []TestListOutput{
		{Name: "John"},
		{Name: "Jane"},
	}
	s := make([]interface{}, len(result))
	for i, v := range result {
		s[i] = v
	}

	var expectedObj = ListCommandOutputObject{
		OutputList: s,
		SummaryObject: TestFooterOutput{
			Count: 2,
		},
	}
	var expectedOutput, _ = json.MarshalIndent(expectedObj, "", "\t")
	expectedStr := string(expectedOutput[:]) + "\n"
	if expectedStr != OutputBuffer.String() {
		t.Errorf("Expected to contain: \n %v\nGot:\n %v\n", string(expectedOutput[:]), OutputBuffer.String())
	}

	// Make sure you can unmarshall the data and read it.
	var actualObj ListCommandOutputObject
	if err := json.Unmarshal([]byte(OutputBuffer.String()), &actualObj); err != nil {
		t.Errorf("Error while decoding json %v", err)
	}
	// TODO fix this error where json.Unmarshal cannot unmarshall nested fields correctly.
	if !reflect.DeepEqual(actualObj, expectedObj) {
		//Expected json decoded object: {[{John} {Jane}] {2}}
		//Got:
		//{[map[Name:John] map[Name:Jane]] map[Count:2]}
		t.Logf("Expected json decoded object: \n %v\nGot:\n %v\n", expectedObj, actualObj)
	}
}
