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
	"io"

	"github.com/GoogleContainerTools/container-structure-test/pkg/version"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var v string

var rootCmd = &cobra.Command{
	Use:   "container-structure-test",
	Short: "container-structure-test provides a framework to test the structure of a container image",
	Long: `container-structure-test provides a powerful framework to validate
the structure of a container image.
These tests can be used to check the output of commands in an image,
as well as verify metadata and contents of the filesystem.`,
}

func NewRootCommand(out, err io.Writer) *cobra.Command {
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, _ []string) error {
		if err := SetUpLogs(err, v); err != nil {
			return err
		}

		rootCmd.SilenceUsage = true
		logrus.Infof("container-structure-test %+v", version.GetVersion())
		return nil
	}

	rootCmd.SilenceErrors = true
	rootCmd.AddCommand(NewCmdVersion(out))
	rootCmd.AddCommand(NewCmdTest(out))

	rootCmd.PersistentFlags().StringVarP(&v, "verbosity", "v", logrus.WarnLevel.String(), "Log level (debug, info, warn, error, fatal, panic)")

	return rootCmd
}

func SetUpLogs(out io.Writer, level string) error {
	logrus.SetOutput(out)
	lvl, err := logrus.ParseLevel(v)
	if err != nil {
		return errors.Wrap(err, "parsing log level")
	}
	logrus.SetLevel(lvl)
	return nil
}
