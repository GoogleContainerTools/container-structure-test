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
	"errors"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"testing"

	"github.com/spf13/cobra"
)

type TestInterface struct {
	Greeting string
	Name     string
}

type TestSubcommandOutput struct {
	Breed string
	Size  string
}

var Greeting string
var Name string

func RunCommand(command *cobra.Command, args []string) (interface{}, error) {
	fmt.Println("Running Hello World Command")
	if Name == "" {
		return (*TestInterface)(nil), errors.New("Please supply Name Argument")
	}
	testOutput := TestInterface{
		Greeting: Greeting,
		Name:     Name,
	}
	return testOutput, nil
}

func TestContainerToolCommandTemplate(t *testing.T) {
	testCommand := ContainerToolCommand{
		ContainerToolCommandBase: &ContainerToolCommandBase{
			Command: &cobra.Command{
				Use: "Hello Command",
			},
			Phase: "test",
		},
		Output: &TestInterface{},
		RunO:   RunCommand,
	}
	var OutputBuffer bytes.Buffer
	testCommand.Flags().StringVarP(&Greeting, "greeting", "g", "Hello", "Greeting")
	testCommand.Flags().StringVarP(&Name, "name", "n", "", "Name")
	testCommand.Command.SetOutput(&OutputBuffer)
	testCommand.SetArgs([]string{"version"})
	Execute(&testCommand)
	// check template applies to the output
	if OutputBuffer.String() != "1.0.1\n" {
		t.Errorf("Expected to contain: \n %v\nGot:\n %v\n", "1.0.1", OutputBuffer.String())
	}
}

func TestContainerToolCommandOutput(t *testing.T) {
	testCommand := ContainerToolCommand{
		ContainerToolCommandBase: &ContainerToolCommandBase{
			Command: &cobra.Command{
				Use: "Hello Command",
			},
			Phase: "test",
		},
		Output: &TestInterface{},
		RunO:   RunCommand,
	}
	testCommand.Flags().StringVarP(&Greeting, "greeting", "g", "Hello", "Greeting")
	testCommand.Flags().StringVarP(&Name, "name", "n", "", "Name")
	var OutputBuffer bytes.Buffer
	testCommand.Command.SetOutput(&OutputBuffer)
	testCommand.SetArgs([]string{"--name=Sparks"})
	Execute(&testCommand)
	var expectedOutput = TestInterface{
		Greeting: "Hello",
		Name:     "Sparks",
	}

	if expectedOutput != testCommand.Output {
		t.Errorf("Expected to contain: \n %v\nGot:\n %v\n", expectedOutput, testCommand.Output)
	}
}

func TestContainerToolCommandSubCommandOutput(t *testing.T) {
	testCommand := &ContainerToolCommand{
		ContainerToolCommandBase: &ContainerToolCommandBase{
			Command: &cobra.Command{
				Use: "Hello Command",
			},
			Phase: "test",
		},
		Output: &TestInterface{},
		RunO:   RunCommand,
	}
	testCommand.SetArgs([]string{"details", "--template", "{{.Breed}}"})
	var OutputBuffer bytes.Buffer
	testCommand.Command.SetOutput(&OutputBuffer)
	testSubCommand := &ContainerToolCommand{
		ContainerToolCommandBase: &ContainerToolCommandBase{
			Command: &cobra.Command{
				Use:   "details",
				Short: "More Info",
			},
		},
		Output: &TestSubcommandOutput{},
		RunO: func(command *cobra.Command, args []string) (interface{}, error) {
			return TestSubcommandOutput{
				Breed: "Chihuhua Mix",
				Size:  "Small",
			}, nil
		},
	}
	testCommand.AddCommand(testSubCommand)
	Execute(testCommand)
	var expectedOutput = TestSubcommandOutput{
		Breed: "Chihuhua Mix",
		Size:  "Small",
	}

	if testSubCommand.Output != expectedOutput {
		t.Errorf("Expected to contain: \n %v\nGot:\n %v\n", expectedOutput, testSubCommand.Output)
	}

	// check template applies to the output
	if OutputBuffer.String() != "Chihuhua Mix\n" {
		t.Errorf("Expected to contain: \n Chihuhua Mix \nGot:\n %v\n", OutputBuffer.String())
	}

}

func TestContainerToolCommandPanic(t *testing.T) {
	testCommand := ContainerToolCommand{
		ContainerToolCommandBase: &ContainerToolCommandBase{
			Command: &cobra.Command{
				Use: "Hello Command",
			},
			Phase: "test",
		},
		Output: &TestInterface{},
		RunO:   RunCommand,
	}
	testCommand.Flags().String("foo", "", "")
	testCommand.MarkFlagRequired("foo")
	if os.Getenv("TEST_EXIT_CODE") == "1" {
		Execute(&testCommand)
		return
	}
	// Run the go test again with environment variable set to run the command.
	cmd := exec.Command(os.Args[0], "-test.run=TestContainerToolCommandPanic")
	cmd.Env = append(os.Environ(), "TEST_EXIT_CODE=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatalf("process ran with err %v, want non zero exit status", err)
}

func TestContainerToolCommandPanicWithNoExit(t *testing.T) {
	defer SetExitOnError(true)
	testCommand := ContainerToolCommand{
		ContainerToolCommandBase: &ContainerToolCommandBase{
			Command: &cobra.Command{
				Use: "Hello Command",
			},
			Phase: "test",
		},
		Output: &TestInterface{},
		RunO:   RunCommand,
	}
	SetExitOnError(false)

	testCommand.Flags().String("foo", "", "")
	testCommand.MarkFlagRequired("foo")
	err := ExecuteE(&testCommand)

	expected := fmt.Sprintf("Required flag(s) %q have/has not been set", "foo")
	if err.Error() != expected {
		t.Errorf("Expected Error: \n %q \nGot:\n %q\n", expected, err.Error())
	}
}

func TestContainerToolCommandRunDefined(t *testing.T) {
	defer SetExitOnError(true)
	runDefined := ContainerToolCommand{
		ContainerToolCommandBase: &ContainerToolCommandBase{
			Command: &cobra.Command{
				Use:   "run",
				Short: "Invalid command Description",
				Run: func(command *cobra.Command, args []string) {
					fmt.Println("Run is defined")
				},
			},
		},
		Output: "",
		RunO: func(command *cobra.Command, args []string) (interface{}, error) {
			return nil, nil
		},
	}
	SetExitOnError(false)
	err := runDefined.ValidateCommand()
	expectedError := ("Cannot provide both Command.Run and RunO implementation." +
		"\nEither implement Command.Run implementation or RunO implemetation")
	if err.Error() != expectedError {
		t.Errorf("Expected Error: \n %q \nGot:\n %q\n", expectedError, err.Error())
	}
}

func TestContainerToolCommandOutputInJson(t *testing.T) {
	testCommand := ContainerToolCommand{
		ContainerToolCommandBase: &ContainerToolCommandBase{
			Command: &cobra.Command{
				Use: "Hello Command",
			},
			Phase: "test",
		},
		Output: &TestInterface{},
		RunO:   RunCommand,
	}
	testCommand.Flags().StringVarP(&Greeting, "greeting", "g", "Hello", "Greeting")
	testCommand.Flags().StringVarP(&Name, "name", "n", "", "Name")
	var OutputBuffer bytes.Buffer
	testCommand.Command.SetOutput(&OutputBuffer)
	testCommand.SetArgs([]string{"--name=Sparks", "--jsonOutput=True"})
	Execute(&testCommand)
	var expectedObj = TestInterface{
		Greeting: "Hello",
		Name:     "Sparks",
	}
	var expectedOutput, _ = json.MarshalIndent(expectedObj, "", "\t")
	expectedStr := string(expectedOutput[:]) + "\n"
	if expectedStr != OutputBuffer.String() {
		t.Errorf("Expected to contain: \n %v\nGot:\n %v\n", string(expectedOutput[:]), OutputBuffer.String())
	}

	// Make sure you can unmarshall the data and read it.
	var actualObj TestInterface
	if err := json.Unmarshal([]byte(OutputBuffer.String()), &actualObj); err != nil {
		t.Errorf("Error while decoding json %v", err)
	}
	if !reflect.DeepEqual(actualObj, expectedObj) {
		t.Errorf("Expected to contain: \n %v\nGot:\n %v\n", expectedObj, actualObj)
	}
}
