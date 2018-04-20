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
	"errors"
	"fmt"

	"github.com/GoogleCloudPlatform/runtimes-common/ctc_lib/flags"

	"github.com/GoogleCloudPlatform/runtimes-common/ctc_lib/constants"

	"github.com/GoogleCloudPlatform/runtimes-common/ctc_lib/util"
	"github.com/spf13/cobra"
)

type ListCommandOutputObject struct {
	OutputList    []interface{}
	SummaryObject interface{}
}

type ContainerToolListCommand struct {
	*ContainerToolCommandBase
	OutputList []interface{}
	// RunO Executes cobra.Command.Run and returns an List[Output]
	RunO func(command *cobra.Command, args []string) ([]interface{}, error)
	// When defined, StreamO Executes cobra.Command.Run and streams each item in the List as its added.
	// This will ignore the RunO function.
	StreamO func(command *cobra.Command, args []string)
	// This function will execute over the output list and return a Summary Object which can be printed.
	// The SummaryTemplate Field can be used to print the Object.
	TotalO          func(list []interface{}) (interface{}, error)
	SummaryObject   interface{}
	SummaryTemplate string
	// Defines the stream to write objects to when using StreamO
	Stream chan interface{}
}

func (commandList ContainerToolListCommand) ReadFromStream(streamOutput bool) ([]interface{}, error) {
	results := make([]interface{}, 0)
	for obj := range commandList.Stream {
		if _, ok := obj.(string); ok {
			// Display any Arbitary strings written to the channel as is.
			// These could be headers or any text.
			// TODO: Provide a callback for users to overwrite this default behavior.
			if streamOutput {
				util.ExecuteTemplate(constants.EmptyTemplate,
					obj, commandList.TemplateFuncMap, commandList.OutOrStdout())
			}
			continue
		}
		results = append(results, obj)
		if streamOutput {
			err := util.ExecuteTemplate(commandList.ReadTemplateFromFlagOrCmdDefault(),
				obj, commandList.TemplateFuncMap, commandList.OutOrStdout())
			if err != nil {
				return nil, err
			}
		}
	}
	return results, nil
}

func (commandList ContainerToolListCommand) isRunODefined() bool {
	return commandList.RunO != nil || commandList.StreamO != nil
}

func (ctc *ContainerToolListCommand) ValidateCommand() error {
	if (ctc.Run != nil || ctc.RunE != nil) && ctc.isRunODefined() {
		return errors.New(`Cannot provide both Command.Run and RunO implementation.
Either implement Command.Run implementation or RunO implemetation`)
	}
	return nil
}

func (ctc *ContainerToolListCommand) printO(c *cobra.Command, args []string) error {
	var commandError, totalFuncError, totalDisplayError error
	var totalDisplayString string
	if ctc.StreamO != nil {
		// Stream Objects only when JsonOutput = False
		ctc.StreamO(c, args)
		ctc.OutputList, commandError = ctc.ReadFromStream(!flags.JsonOutput)
	} else {
		// Run RunO function.
		ctc.OutputList, commandError = ctc.RunO(c, args)
		LogIfErr(commandError, Log)
	}
	totalDisplayString, totalFuncError, totalDisplayError = ctc.runTotalIfDefined()
	LogIfErr(totalFuncError, Log)
	ctc.printResult(totalDisplayString)
	LogIfErr(totalDisplayError, Log)

	if commandError == nil && totalDisplayError == nil && totalFuncError == nil {
		return nil
	} else if commandError != nil {
		// Return Command Error if occurred
		return commandError
	} else if totalFuncError != nil {
		// Return TotalFuncError if occurred
		return totalFuncError
	} else {
		return totalDisplayError
	}
}

func (ctc *ContainerToolListCommand) runTotalIfDefined() (string, error, error) {
	var totalCommandError, totalError error
	var OutputBuffer bytes.Buffer
	// If TotalO function defined & Summary Template provided, get the summary text
	if ctc.TotalO != nil && ctc.SummaryTemplate != "" {
		ctc.SummaryObject, totalCommandError = ctc.TotalO(ctc.OutputList)
		totalError = util.ExecuteTemplate(ctc.SummaryTemplate,
			ctc.SummaryObject, ctc.TemplateFuncMap, &OutputBuffer)
	}
	return OutputBuffer.String(), totalCommandError, totalError
}

func (ctc *ContainerToolListCommand) printResult(totalDisplayString string) {
	if flags.JsonOutput {
		data := ListCommandOutputObject{
			OutputList:    ctc.OutputList,
			SummaryObject: ctc.SummaryObject,
		}
		util.PrintJson(data, ctc.OutOrStdout())
	} else {
		if ctc.StreamO == nil {
			err := util.ExecuteTemplate(ctc.ReadTemplateFromFlagOrCmdDefault(),
				ctc.OutputList, ctc.TemplateFuncMap, ctc.OutOrStdout())
			LogIfErr(err, Log)
		}
		// Display total if defined.
		if totalDisplayString != "" {
			fmt.Fprintln(ctc.OutOrStdout(), totalDisplayString)
		}
	}
}
