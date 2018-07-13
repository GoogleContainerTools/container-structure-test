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
	"text/template"

	"github.com/GoogleCloudPlatform/runtimes-common/ctc_lib/config"
	"github.com/GoogleCloudPlatform/runtimes-common/ctc_lib/constants"
	"github.com/GoogleCloudPlatform/runtimes-common/ctc_lib/flags"
	"github.com/GoogleCloudPlatform/runtimes-common/ctc_lib/logging"
	"github.com/GoogleCloudPlatform/runtimes-common/ctc_lib/types"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type ContainerToolCommandBase struct {
	*cobra.Command
	Phase           string
	DefaultTemplate string //TODO: Validate Default Config.
	TemplateFuncMap template.FuncMap
}

func (ctb *ContainerToolCommandBase) getCommand() *cobra.Command {
	return ctb.Command
}

func (ctb *ContainerToolCommandBase) setRunE(cobraRunE func(c *cobra.Command, args []string) error) {
	if ctb.Run == nil && ctb.RunE == nil {
		ctb.RunE = cobraRunE
	}
}

func (ctb *ContainerToolCommandBase) toolName() string {
	return ctb.Name()
}

func (ctb *ContainerToolCommandBase) Init() {
	// Init Logging with info level with colors disabled since initLogging gets called
	// only after arguments are parsed correctly.
	Log = logging.NewLogger(
		viper.GetString(config.LogDirConfigKey),
		ctb.Name(),
		log.InfoLevel,
		false,
	)
	cobra.OnInitialize(initConfig, ctb.initLogging, ctb.SetSilenceUsage)
	ctb.AddFlags()
	ctb.AddSubCommands()
}

func (ctb *ContainerToolCommandBase) SetSilenceUsage() {
	// Do not display usage when using RunE after args are parsed.
	// See https://github.com/spf13/cobra/issues/340 for more information.
	ctb.SilenceUsage = true
}

func (ctb *ContainerToolCommandBase) initLogging() {
	Log = logging.NewLogger(
		viper.GetString(config.LogDirConfigKey),
		ctb.Name(),
		flags.Verbosity.Level,
		flags.EnableColors,
	)
	Log.SetLevel(flags.Verbosity.Level)
	Log.AddHook(logging.NewFatalHook(exitOnError))
	logging.InitStdOutLogger(flags.EnableColors, flags.Verbosity.Level)
}

func (ctb *ContainerToolCommandBase) AddSubCommands() {
	// Add version subcommand
	ctb.AddCommand(VersionCommand)
	ConfigCommand.Command.AddCommand(SetConfigCommand)
	ctb.AddCommand(ConfigCommand)
	ctb.AddCommand(UpdateCheckCommand)

	// Set up Root Command
	ctb.Command.SetHelpTemplate(HelpTemplate)
}

func (ctb *ContainerToolCommandBase) AddCommand(command CLIInterface) {
	cobraRunE := func(c *cobra.Command, args []string) error {
		return command.printO(c, args)
	}
	command.setRunE(cobraRunE)
	ctb.Command.AddCommand(command.getCommand())
}

func (ctb *ContainerToolCommandBase) AddFlags() {
	// Add template Flag
	ctb.PersistentFlags().StringVarP(&flags.TemplateString, "template", "t", constants.EmptyTemplate, "Output format")
	ctb.PersistentFlags().VarP(types.NewLogLevel(constants.DefaultLogLevel, &flags.Verbosity), "verbosity", "v",
		`verbosity. Logs to File when verbosity is set to Debug. For all other levels Logs to StdOut.`)
	ctb.PersistentFlags().BoolVarP(&flags.UpdateCheck, "updateCheck", "u", true, "Run Update Check")
	viper.BindPFlag("updateCheck", ctb.PersistentFlags().Lookup("updateCheck"))

	ctb.PersistentFlags().BoolVar(&flags.EnableColors, "enableColors", true, `Enable Colors when displaying logs to Std Out.`)
	ctb.PersistentFlags().StringVar(&flags.LogDir, "logDir", "", "LogDir")
	viper.BindPFlag("logDir", ctb.PersistentFlags().Lookup("logDir"))

	ctb.PersistentFlags().BoolVar(&flags.JsonOutput, "jsonOutput", false, "Output Json format")
}

func (ctb *ContainerToolCommandBase) ReadTemplateFromFlagOrCmdDefault() string {
	if flags.TemplateString == constants.EmptyTemplate && ctb.DefaultTemplate != "" {
		return ctb.DefaultTemplate
	}
	return flags.TemplateString
}
