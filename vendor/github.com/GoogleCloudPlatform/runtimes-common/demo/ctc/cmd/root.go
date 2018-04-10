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

package cmd

import (
	"github.com/GoogleCloudPlatform/runtimes-common/ctc_lib"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type RootCommandOutput struct {
	Message string
}

var Message string

var RootCommand = &ctc_lib.ContainerToolCommand{
	ContainerToolCommandBase: &ctc_lib.ContainerToolCommandBase{
		Command: &cobra.Command{
			Use:   "echo",
			Short: "Echo's Message",
		},
		Phase:           "test",
		DefaultTemplate: "{{.Message}}",
	},
	Output: &RootCommandOutput{},
	RunO: func(command *cobra.Command, args []string) (interface{}, error) {
		// An Example of Logging.
		ctc_lib.Log.WithFields(log.Fields{
			"message": viper.GetString("message"),
		}).Info("You are running echo command with following values ")
		return RootCommandOutput{
			Message: viper.GetString("message"),
		}, nil
	},
}

func Execute() {
	ctc_lib.Version = "1.0.1"
	ctc_lib.ConfigFile = "demo/ctc/testConfig.json"
	ctc_lib.ReleaseUrl = "https://raw.githubusercontent.com/tejal29/runtimes-common/add_update_check_logic/demo/ctc/releases.json"
	ctc_lib.Execute(RootCommand)
}

func init() {
	RootCommand.Flags().StringVarP(&Message, "message", "m", "YOUR TEXT TO ECHO", "Message to Echo")
	//Add Subcommand using AddCommand.
	RootCommand.AddCommand(PanicCommand)
	viper.BindPFlag("message", RootCommand.Flags().Lookup("message"))
}
