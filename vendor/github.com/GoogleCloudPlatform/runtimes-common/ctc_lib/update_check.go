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
	"github.com/GoogleCloudPlatform/runtimes-common/ctc_lib/notify"
	"github.com/blang/semver"
	"github.com/spf13/cobra"
)

type UpdateCheckOutput struct {
	CurrentVersion semver.Version
	LatestVersion  semver.Version
	DownloadUrl    string
}

var UpdateCheckCommand = &ContainerToolCommand{
	ContainerToolCommandBase: &ContainerToolCommandBase{
		Command: &cobra.Command{
			Use:   "updatecheck",
			Short: "Checks if an update is available",
			Long:  `Checks if an update is available.`,
			Args:  cobra.ExactArgs(0),
		},
		DefaultTemplate: `{{if .CurrentVersion.EQ .LatestVersion }}You are at the latest Version.
No updates Available.{{end}}
{{if .CurrentVersion.LT .LatestVersion}}There is a newer version {{.LatestVersion}} of tool available.
Download it here: {{.DownloadUrl}}{{end}}`,
	},
	Output: &UpdateCheckOutput{},
	RunO: func(command *cobra.Command, args []string) (interface{}, error) {
		if ReleaseUrl == "" {
			Log.Panicf("No ReleaseUrl defined. Cannot Check for Updates.")
		}
		latestVersion, err := notify.GetLatestVersionFromURL(ReleaseUrl, VersionPrefix)
		if err != nil {
			Log.Panic(err)
		}
		currentVersion, err := semver.Make(Version)
		if err != nil {
			Log.Panic(err)
		}
		var updateCheck = UpdateCheckOutput{
			CurrentVersion: currentVersion,
			LatestVersion:  latestVersion,
			DownloadUrl:    DownloadUrl,
		}
		return updateCheck, nil
	},
}
