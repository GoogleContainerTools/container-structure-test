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
	"errors"

	"github.com/GoogleCloudPlatform/runtimes-common/ctc_lib/util"
	"github.com/spf13/cobra"
)

type ContainerToolCommand struct {
	*ContainerToolCommandBase
	Output interface{}
	// RunO Executes cobra.Command.Run and returns an Output
	RunO func(command *cobra.Command, args []string) (interface{}, error)
}

func (ctc *ContainerToolCommand) isRunODefined() bool {
	return ctc.RunO != nil
}

func (ctc *ContainerToolCommand) ValidateCommand() error {
	if (ctc.Run != nil || ctc.RunE != nil) && ctc.isRunODefined() {
		return errors.New(`Cannot provide both Command.Run and RunO implementation.
Either implement Command.Run implementation or RunO implemetation`)
	}
	return nil
}

func (ctc *ContainerToolCommand) printO(c *cobra.Command, args []string) error {
	obj, err := ctc.RunO(c, args)
	ctc.Output = obj
	display_err := util.ExecuteTemplate(ctc.ReadTemplateFromFlagOrCmdDefault(),
		ctc.Output, ctc.TemplateFuncMap, ctc.OutOrStdout())
	if err != nil {
		return err
	}
	return display_err
}
