/*
Copyright 2018 Google, Inc. All rights reserved.
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
	"github.com/spf13/cobra"
)

type VersionOutput struct {
	Version string
}

var VersionCommand = &ContainerToolCommand{
	ContainerToolCommandBase: &ContainerToolCommandBase{
		Command: &cobra.Command{
			Use:   "version",
			Short: "Print the version",
			Long:  `Print the version`,
			Args:  cobra.ExactArgs(0),
		},
		DefaultTemplate: "{{.Version}}",
	},
	Output: &VersionOutput{},
	RunO: func(command *cobra.Command, args []string) (interface{}, error) {
		var versionOutput = VersionOutput{
			Version: VersionPrefix + Version,
		}
		return versionOutput, nil
	},
}
