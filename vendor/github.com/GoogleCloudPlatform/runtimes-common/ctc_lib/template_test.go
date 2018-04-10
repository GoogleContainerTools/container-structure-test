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
	"strings"
	"testing"
	"text/template"

	"github.com/spf13/cobra"
)

type TemplateTestResult struct {
	Greeting string
	Name     string
}

func RunTemplateCommand(command *cobra.Command, args []string) (interface{}, error) {
	testOutput := TestInterface{
		Greeting: "hello",
		Name:     "world",
	}
	return testOutput, nil
}

func TestContainerToolIncorrectTemplate(t *testing.T) {
	defer SetExitOnError(true)
	testCommand := ContainerToolCommand{
		ContainerToolCommandBase: &ContainerToolCommandBase{
			Command: &cobra.Command{
				Use: "Hello Command",
			},
			Phase:           "test",
			DefaultTemplate: "{{.greet}}",
		},
		Output: &TestInterface{},
		RunO:   RunTemplateCommand,
	}
	SetExitOnError(false)
	err := ExecuteE(&testCommand)
	expectedError := ("can't evaluate field greet in type ctc_lib.TestInterface")
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected Error to contain: \n %q \nGot:\n %q\n", expectedError, err.Error())
	}
}

func TestContainerToolTestOutputWithFuncTemplate(t *testing.T) {
	testCommand := ContainerToolCommand{
		ContainerToolCommandBase: &ContainerToolCommandBase{
			Command: &cobra.Command{
				Use: "Hello Command",
			},
			Phase:           "test",
			DefaultTemplate: "{{toUpper .Greeting}} {{toUpper .Name}}!",
			TemplateFuncMap: template.FuncMap{
				"toUpper": strings.ToUpper,
			},
		},
		Output: &TestInterface{},
		RunO:   RunTemplateCommand,
	}
	var OutputBuffer bytes.Buffer
	testCommand.Command.SetOutput(&OutputBuffer)
	Execute(&testCommand)
	if OutputBuffer.String() != "HELLO WORLD!\n" {
		t.Errorf("Expected to contain: \n HELLO WORLD! \nGot:\n %v\n", OutputBuffer.String())
	}
}

func TestContainerToolTestOutputWithNilFuncMap(t *testing.T) {
	testCommand := ContainerToolCommand{
		ContainerToolCommandBase: &ContainerToolCommandBase{
			Command: &cobra.Command{
				Use: "Hello Command",
			},
			Phase:           "test",
			DefaultTemplate: "{{.Greeting}} {{.Name}}!",
			TemplateFuncMap: nil,
		},
		Output: &TestInterface{},
		RunO:   RunTemplateCommand,
	}
	var OutputBuffer bytes.Buffer
	testCommand.Command.SetOutput(&OutputBuffer)
	Execute(&testCommand)
	if OutputBuffer.String() != "hello world!\n" {
		t.Errorf("Expected to contain: \n hello world! \nGot:\n %v\n", OutputBuffer.String())
	}
}
