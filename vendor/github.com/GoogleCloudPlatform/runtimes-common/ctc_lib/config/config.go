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

package config

import (
	"fmt"

	"github.com/GoogleCloudPlatform/runtimes-common/ctc_lib/constants"
)

var DefaultConfig = []byte(fmt.Sprintf(`{
	"%s": true,
	"%s": %d
	}`, UpdateCheckConfigKey, UpdateCheckIntervalInSecs, constants.DayInSeconds))

var DefaultConfigType = "json"

// All Keys are Lower case since Viper converts them to lower case after reading.
const (
	UpdateCheckConfigKey      = "updatecheck"
	LogDirConfigKey           = "logdir"
	UpdateCheckIntervalInSecs = "update_check_interval_in_secs"
	TmpDirKey                 = "tmpdir"
)
