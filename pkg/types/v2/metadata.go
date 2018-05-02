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
	"fmt"

	"github.com/GoogleCloudPlatform/runtimes-common/ctc_lib"
	"github.com/GoogleContainerTools/container-structure-test/pkg/drivers"
	types "github.com/GoogleContainerTools/container-structure-test/pkg/types/unversioned"
	"github.com/GoogleContainerTools/container-structure-test/pkg/utils"
)

type MetadataTest struct {
	Env          []types.EnvVar `yaml:"env"`
	ExposedPorts []string       `yaml:"exposedPorts"`
	Entrypoint   *[]string      `yaml:"entrypoint"`
	Cmd          *[]string      `yaml:"cmd"`
	Workdir      string         `yaml:"workdir"`
	Volumes      []string       `yaml:"volumes"`
	Labels       []types.Label  `yaml:"labels"`
}

func (mt MetadataTest) IsEmpty() bool {
	return len(mt.Env) == 0 &&
		len(mt.ExposedPorts) == 0 &&
		mt.Entrypoint == nil &&
		mt.Cmd == nil &&
		mt.Workdir == "" &&
		len(mt.Volumes) == 0 &&
		len(mt.Labels) == 0
}

func (mt MetadataTest) LogName() string {
	return "Metadata Test"
}

func (mt MetadataTest) Validate() error {
	for _, envVar := range mt.Env {
		if envVar.Key == "" {
			return fmt.Errorf("Environment variable key cannot be empty")
		}
	}
	for _, label := range mt.Labels {
		if label.Key == "" {
			return fmt.Errorf("Label key cannot be empty")
		}
	}
	for _, port := range mt.ExposedPorts {
		if port == "" {
			return fmt.Errorf("Port cannot be empty")
		}
	}
	for _, volume := range mt.Volumes {
		if volume == "" {
			return fmt.Errorf("Volume cannot be empty")
		}
	}
	return nil
}

func (mt MetadataTest) Run(driver drivers.Driver) *types.TestResult {
	result := &types.TestResult{
		Name: mt.LogName(),
		Pass: true,
	}
	ctc_lib.Log.Debug(mt.LogName())
	imageConfig, err := driver.GetConfig()
	if err != nil {
		result.Errorf("Error retrieving image config: %s", err.Error())
		result.Fail()
		return result
	}

	for _, pair := range mt.Env {
		if act_val, has_key := imageConfig.Env[pair.Key]; has_key {
			if !utils.CompileAndRunRegex(pair.Value, act_val, true) {
				result.Errorf("env var %s value does not match expected value: %s", pair.Key, pair.Value)
				result.Fail()
			}
		} else {
			result.Errorf("variable %s not found in image env", pair.Key)
			result.Fail()
		}
	}

	for _, pair := range mt.Labels {
		if act_val, has_key := imageConfig.Labels[pair.Key]; has_key {
			if !utils.CompileAndRunRegex(pair.Value, act_val, true) {
				result.Errorf("label %s value does not match expected value: %s", pair.Key, pair.Value)
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
			fmt.Printf("config length: %d\n", len(*mt.Cmd))
			fmt.Printf("image command length: %d\n", len(imageConfig.Cmd))
			fmt.Printf("single config command entry: %s\n", (*mt.Cmd)[0])
			fmt.Printf("single image command entry: %s\n", imageConfig.Cmd[0])
			for i := range *mt.Cmd {
				fmt.Println(i)
				if (*mt.Cmd)[i] != imageConfig.Cmd[i] {
					result.Errorf("Image config Cmd does not match expected value: %s", *mt.Cmd)
					result.Fail()
				}
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
