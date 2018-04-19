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
	"github.com/GoogleCloudPlatform/runtimes-common/ctc_lib/logging"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type ConfigOutput struct {
	Config map[string]interface{}
}

var ConfigCommand = &ContainerToolCommand{
	ContainerToolCommandBase: &ContainerToolCommandBase{
		Command: &cobra.Command{
			Use: "config",
			Long: `Prints all Config Keys.
This command evalues all the keys from Default Config and Tool Config and prints
the final value`,
			Short: "Prints all Config Keys",
			Args:  cobra.ExactArgs(0),
		},
		DefaultTemplate: "{{ range $k, $v := .Config }}{{$k}} : {{$v}}\n{{ end }}",
	},
	Output: &ConfigOutput{},
	RunO: func(command *cobra.Command, args []string) (interface{}, error) {
		return &ConfigOutput{
			Config: viper.AllSettings(),
		}, nil
	},
}

var SetConfigCommand = &cobra.Command{
	Use:   "set",
	Long:  `Sets the config Key and makes the change in the Tool Config File.`,
	Short: "Sets the config Key in the Tool Config",
	Args:  cobra.ExactArgs(2),
	RunE: func(command *cobra.Command, args []string) error {
		_, exists := viper.AllSettings()[args[0]]
		if !exists {
			Log.Panicf("Config Key %s does not exist.", args[0])
		}
		// We do not want to add the default config keys in the config file.
		// Hence we create a new instance of viper and read the config file again.
		v := viper.New()
		v.SetConfigFile(ConfigFile)
		v.ReadInConfig()
		v.Set(args[0], args[1])
		v.WriteConfig()
		logging.Out.Infof("Config key Changed and written to file %s", ConfigFile)
		return nil
	},
}
