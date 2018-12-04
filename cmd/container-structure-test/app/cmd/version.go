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

	"github.com/GoogleContainerTools/container-structure-test/cmd/container-structure-test/app/flags"
	"github.com/GoogleContainerTools/container-structure-test/pkg/version"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var versionFlag = flags.NewTemplateFlag("{{.Version}}\n", version.Info{})

func NewCmdVersion(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version information",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunVersion(out, cmd)
		},
	}

	cmd.Flags().VarP(versionFlag, "output", "o", versionFlag.Usage())
	return cmd
}

func RunVersion(out io.Writer, cmd *cobra.Command) error {
	if err := versionFlag.Template().Execute(out, version.GetVersion()); err != nil {
		return errors.Wrap(err, "executing template")
	}
	return nil
}
