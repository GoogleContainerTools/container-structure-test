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
	"github.com/GoogleContainerTools/container-structure-test/pkg/drivers"
	types "github.com/GoogleContainerTools/container-structure-test/pkg/types/unversioned"
	"github.com/GoogleContainerTools/container-structure-test/pkg/utils"
	"github.com/sirupsen/logrus"
)

type MetadataTest struct {
	Env              []types.EnvVar `yaml:"env"`
	UnboundEnv       []types.EnvVar `yaml:"unboundEnv"`
	ExposedPorts     []string       `yaml:"exposedPorts"`
	UnexposedPorts   []string       `yaml:"unexposedPorts"`
	Entrypoint       *[]string      `yaml:"entrypoint"`
	Cmd              *[]string      `yaml:"cmd"`
	Workdir          string         `yaml:"workdir"`
	Volumes          []string       `yaml:"volumes"`
	UnmountedVolumes []string       `yaml:"unmountedVolumes"`
	Labels           []types.Label  `yaml:"labels"`
	User             string         `yaml:"user"`
}

func (mt MetadataTest) IsEmpty() bool {
	return len(mt.Env) == 0 &&
		len(mt.UnboundEnv) == 0 &&
		len(mt.ExposedPorts) == 0 &&
		len(mt.UnexposedPorts) == 0 &&
		mt.Entrypoint == nil &&
		mt.Cmd == nil &&
		mt.Workdir == "" &&
		mt.User == "" &&
		len(mt.Volumes) == 0 &&
		len(mt.UnmountedVolumes) == 0 &&
		len(mt.Labels) == 0
}

func (mt MetadataTest) LogName() string {
	return "Metadata Test"
}

func (mt MetadataTest) Validate(channel chan interface{}) bool {
	res := &types.TestResult{
		Name: mt.LogName(),
	}
	for _, envVar := range mt.Env {
		if envVar.Key == "" {
			res.Error("Environment variable key cannot be empty")
		}
	}
	for _, label := range mt.Labels {
		if label.Key == "" {
			res.Error("Label key cannot be empty")
		}
	}
	for _, port := range mt.ExposedPorts {
		if port == "" {
			res.Error("Port cannot be empty")
		}
	}
	for _, volume := range mt.Volumes {
		if volume == "" {
			res.Error("Volume cannot be empty")
		}
	}
	if len(res.Errors) > 0 {
		channel <- res
		return false
	}
	return true
}

func (mt MetadataTest) Run(driver drivers.Driver) *types.TestResult {
	result := &types.TestResult{
		Name: mt.LogName(),
		Pass: true,
	}
	logrus.Debug(mt.LogName())
	imageConfig, err := driver.GetConfig()
	if err != nil {
		result.Errorf("Error retrieving image config: %s", err.Error())
		result.Fail()
		return result
	}

	for _, pair := range mt.Env {
		if val, ok := imageConfig.Env[pair.Key]; ok {
			var match bool
			if pair.IsRegex {
				match = utils.CompileAndRunRegex(pair.Value, val, true)
			} else {
				match = (pair.Value == val)
			}
			if !match {
				result.Errorf("env var %s value %s does not match expected value: %s", pair.Key, val, pair.Value)
				result.Fail()
			}
		} else {
			result.Errorf("variable %s not found in image env", pair.Key)
			result.Fail()
		}
	}

	for _, pair := range mt.UnboundEnv {
		if _, ok := imageConfig.Env[pair.Key]; ok {
			result.Errorf("env variable %s should not be present in image metadata", pair.Key)
			result.Fail()
		}
	}

	for _, pair := range mt.Labels {
		if val, ok := imageConfig.Labels[pair.Key]; ok {
			var match bool
			if pair.IsRegex {
				match = utils.CompileAndRunRegex(pair.Value, val, true)
			} else {
				match = (pair.Value == val)
			}
			if !match {
				result.Errorf("label %s value %s does not match expected value: %s", pair.Key, val, pair.Value)
				result.Fail()
			}
		} else {
			result.Errorf("label %s not found in image metadata", pair.Key)
			result.Fail()
		}
	}

	if mt.Cmd != nil {
		if len(*mt.Cmd) != len(imageConfig.Cmd) {
			result.Errorf("Image Cmd %v does not match expected Cmd: %v", imageConfig.Cmd, *mt.Cmd)
			result.Fail()
		} else if len(*mt.Cmd) > 0 {
			for i := range *mt.Cmd {
				if (*mt.Cmd)[i] != imageConfig.Cmd[i] {
					result.Errorf("Image config Cmd %v does not match expected value: %s", imageConfig.Cmd, *mt.Cmd)
					result.Fail()
				}
			}
		}
	}

	if mt.Entrypoint != nil {
		if len(*mt.Entrypoint) != len(imageConfig.Entrypoint) {
			result.Errorf("Image entrypoint %v does not match expected entrypoint: %v", imageConfig.Entrypoint, *mt.Entrypoint)
			result.Fail()
		} else {
			for i := range *mt.Entrypoint {
				if (*mt.Entrypoint)[i] != imageConfig.Entrypoint[i] {
					result.Errorf("Image config entrypoint %v does not match expected value: %s", imageConfig.Entrypoint, *mt.Entrypoint)
					result.Fail()
				}
			}
		}
	}

	if mt.Workdir != "" && mt.Workdir != imageConfig.Workdir {
		result.Errorf("Image workdir %s does not match config workdir: %s", imageConfig.Workdir, mt.Workdir)
		result.Fail()
	}

	if mt.User != "" && mt.User != imageConfig.User {
		result.Errorf("Image user %s does not match config user: %s", imageConfig.User, mt.User)
		result.Fail()
	}

	for _, port := range mt.ExposedPorts {
		if !utils.ValueInList(port, imageConfig.ExposedPorts) {
			result.Errorf("Port %s not found in config", port)
			result.Fail()
		}
	}

	for _, port := range mt.UnexposedPorts {
		if utils.ValueInList(port, imageConfig.ExposedPorts) {
			result.Errorf("Port %s should not be exposed", port)
			result.Fail()
		}
	}

	for _, volume := range mt.Volumes {
		if !utils.ValueInList(volume, imageConfig.Volumes) {
			result.Errorf("Volume %s not found in config", volume)
			result.Fail()
		}
	}

	for _, volume := range mt.UnmountedVolumes {
		if utils.ValueInList(volume, imageConfig.Volumes) {
			result.Errorf("Volume %s should not be mounted", volume)
			result.Fail()
		}
	}
	return result
}
