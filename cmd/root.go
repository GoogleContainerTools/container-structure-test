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
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var logLevel string
var imagePath, driver, metadata string
var save, pull, force, quiet, verbose bool

var configFiles []string

var RootCmd = &cobra.Command{
	Use:   "container-structure-test",
	Short: "container-structure-test provides a framework to test the structure of a container image",
	Long: `container-structure-test provides a powerful framework to validate
the structure of a container image.
These tests can be used to check the output of commands in an image, 
as well as verify metadata and contents of the filesystem.`,
	PersistentPreRun: func(c *cobra.Command, s []string) {
		ll, err := logrus.ParseLevel(logLevel)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		logrus.SetLevel(ll)
	},
}

func init() {
	RootCmd.PersistentFlags().StringVar(&logLevel, "logLevel", "warning", "This flag controls the verbosity of container-structure-test.")
}
