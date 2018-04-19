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
	"fmt"
	"path/filepath"
	"time"

	"github.com/GoogleCloudPlatform/runtimes-common/ctc_lib/config"
	"github.com/GoogleCloudPlatform/runtimes-common/ctc_lib/constants"
	"github.com/GoogleCloudPlatform/runtimes-common/ctc_lib/logging"
	"github.com/GoogleCloudPlatform/runtimes-common/ctc_lib/notify"
	"github.com/GoogleCloudPlatform/runtimes-common/ctc_lib/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type CLIInterface interface {
	printO(c *cobra.Command, args []string) error
	setRunE(func(c *cobra.Command, args []string) error)
	getCommand() *cobra.Command
	ValidateCommand() error
	isRunODefined() bool
	toolName() string
	Init()
}

func Execute(ctb CLIInterface) {
	defer errRecover()
	err := ExecuteE(ctb)
	CommandExit(err)
}

func ExecuteE(ctb CLIInterface) (err error) {
	if err := ctb.ValidateCommand(); err != nil {
		return err
	}
	ctb.Init()
	if ctb.isRunODefined() {
		cobraRunE := func(c *cobra.Command, args []string) error {
			err = ctb.printO(c, args)
			return err
		}
		ctb.setRunE(cobraRunE)
	}

	err = ctb.getCommand().Execute()

	//Add empty line as template.Execute does not print an empty line
	ctb.getCommand().Println()

	// Run Update Command to see if Updates are available.

	lastUpdatedCheckFilePath := filepath.Join(
		util.GetToolTempDirOrDefault(viper.GetString(config.TmpDirKey), ctb.toolName()),
		constants.LastUpdatedCheckFileName)
	Log.WithFields(log.Fields{
		"updatecheck":             viper.GetString(config.UpdateCheckConfigKey),
		"last_updated_check_file": lastUpdatedCheckFilePath,
		"update_interval_in_sec":  viper.GetFloat64(config.UpdateCheckIntervalInSecs),
	}).Debug("Checking if Update Check is required")
	if notify.ShouldCheckURLVersion(lastUpdatedCheckFilePath) && ReleaseUrl != "" {
		// Calling UpdateCheckCommand Explicitly. Hence no need to pass args.
		Log.Debug("Running Update Check Command")
		UpdateCheckCommand.Run(ctb.getCommand(), nil)
		notify.WriteTimeToFile(lastUpdatedCheckFilePath, time.Now().UTC())
	}

	if util.IsDebug(Log.Level) {
		logFile, ok := logging.GetCurrentFileName(Log)
		if ok {
			ctb.getCommand().Println("See logs at ", logFile)
		}
	}
	return err
}

// errRecover is the handler that turns panics into returns from the top
// level of Parse.
func errRecover() {
	if e := recover(); e != nil {
		err := fmt.Errorf("%v", e)
		CommandExit(err)
	}
}
