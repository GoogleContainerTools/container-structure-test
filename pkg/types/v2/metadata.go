// Copyright 2017 Google Inc. All rights reserved.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v2

import (
	"github.com/sirupsen/logrus"

	"github.com/GoogleCloudPlatform/container-structure-test/pkg/drivers"
	types "github.com/GoogleCloudPlatform/container-structure-test/pkg/types/unversioned"
	"github.com/GoogleCloudPlatform/container-structure-test/pkg/utils"
)

type MetadataTest struct {
	Env          []types.EnvVar `yaml:"env"`
	ExposedPorts []string       `yaml:"exposedPorts"`
	Entrypoint   *[]string      `yaml:"entrypoint"`
	Cmd          *[]string      `yaml:"cmd"`
	Workdir      string         `yaml:"workdir"`
	Volumes      []string       `yaml:"volumes"`
}

func (mt MetadataTest) LogName() string {
	return "Metadata Test"
}

func (mt MetadataTest) Run(driver drivers.Driver) *types.TestResult {
	result := &types.TestResult{
		Name: mt.LogName(),
		Pass: true,
	}
	logrus.Info(mt.LogName())
	imageConfig, err := driver.GetConfig()
	if err != nil {
		result.Errorf("Error retrieving image config: %s", err.Error())
		result.Fail()
		return result
	}
	for _, pair := range mt.Env {
		if imageConfig.Env[pair.Key] == "" {
			result.Errorf("variable %s not found in image env", pair.Key)
			result.Fail()
		} else if imageConfig.Env[pair.Key] != pair.Value {
			result.Errorf("env var %s value does not match expected value: %s", pair.Key, pair.Value)
			result.Fail()
		}
	}

	if mt.Cmd != nil {
		if len(*mt.Cmd) != len(imageConfig.Cmd) {
			result.Errorf("Image Cmd %v does not match expected Cmd: %v", imageConfig.Cmd, *mt.Cmd)
			result.Fail()
		}
		for i := range *mt.Cmd {
			if (*mt.Cmd)[i] != imageConfig.Cmd[i] {
				result.Errorf("Image config Cmd does not match expected value: %s", *mt.Cmd)
				result.Fail()
			}
		}
	}

	if mt.Entrypoint != nil {
		if len(*mt.Entrypoint) != len(imageConfig.Entrypoint) {
			result.Errorf("Image entrypoint %v does not match expected entrypoint: %v", imageConfig.Entrypoint, *mt.Entrypoint)
			result.Fail()
		}
		for i := range *mt.Entrypoint {
			if (*mt.Entrypoint)[i] != imageConfig.Entrypoint[i] {
				result.Errorf("Image config entrypoint does not match expected value: %s", *mt.Entrypoint)
				result.Fail()
			}
		}
	}

	if mt.Workdir != "" && mt.Workdir != imageConfig.Workdir {
		result.Errorf("Image workdir %s does not match config workdir: %s", imageConfig.Workdir, mt.Workdir)
		result.Fail()
	}

	for _, port := range mt.ExposedPorts {
		if !utils.ValueInList(port, imageConfig.ExposedPorts) {
			result.Errorf("Port %s not found in config", port)
			result.Fail()
		}
	}

	for _, volume := range mt.Volumes {
		if !utils.ValueInList(volume, imageConfig.Volumes) {
			result.Errorf("Volume %s not found in config", volume)
			result.Fail()
		}
	}
	return result
}
