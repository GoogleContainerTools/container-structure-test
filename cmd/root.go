// Copyright 2018 Google Inc. All rights reserved.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"os"

	"github.com/GoogleCloudPlatform/runtimes-common/ctc_lib"
	"github.com/GoogleCloudPlatform/runtimes-common/ctc_lib/flags"
	"github.com/spf13/cobra"
)

var logLevel string
var imagePath, driver, metadata, testReport string
var save, pull, force, quiet bool

var configFiles []string

var RootCmd = &ctc_lib.ContainerToolListCommand{
	ContainerToolCommandBase: &ctc_lib.ContainerToolCommandBase{
		Command: &cobra.Command{
			Use:   "container-structure-test",
			Short: "container-structure-test provides a framework to test the structure of a container image",
			Long: `container-structure-test provides a powerful framework to validate
the structure of a container image.
These tests can be used to check the output of commands in an image,
as well as verify metadata and contents of the filesystem.`,
			SilenceErrors: true,
			PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
				if testReport != "" {
					// Force JsonOutput
					flags.JsonOutput = true
					// Create TestReport File with permissions 0666 Truncate previous if Exists already.
					TestReportFile, err := os.Create(testReport)
					if err != nil {
						return err
					}
					TestCmd.SetOutput(TestReportFile)
				}
				return nil
			},
		},
		Phase:           "stable",
		DefaultTemplate: "{{.}}",
	},
}
