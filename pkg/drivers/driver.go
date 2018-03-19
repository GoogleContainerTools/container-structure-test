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

package drivers

import (
	"os"
	"strings"

	"github.com/GoogleCloudPlatform/container-structure-test/pkg/types/unversioned"
)

const (
	Docker = "docker"
	Tar    = "tar"
	Host   = "host"
)

type DriverConfig struct {
	Image    string // used by Docker/Tar drivers
	Save     bool   // used by Docker/Tar drivers
	Metadata string // used by Host driver
}

type Driver interface {
	Setup(envVars []unversioned.EnvVar, fullCommands [][]string) error

	// Teardown is optional and is only used in the host driver
	Teardown(envVars []unversioned.EnvVar, fullCommands [][]string) error

	// given an array of command parts, construct a full command and execute it against the
	// current environment. a list of environment variables can be passed to be set in the
	// environment before the command is executed. additionally, a boolean flag is passed
	// to specify whether or not we care about the output of the command.
	ProcessCommand(envVars []unversioned.EnvVar, fullCommand []string) (string, string, int, error)

	StatFile(path string) (os.FileInfo, error)

	ReadFile(path string) ([]byte, error)

	ReadDir(path string) ([]os.FileInfo, error)

	GetConfig() (unversioned.Config, error)

	Destroy()
}

func InitDriverImpl(driver string) func(DriverConfig) (Driver, error) {
	switch driver {
	// future drivers will be added here
	case Docker:
		return NewDockerDriver
	case Tar:
		return NewTarDriver
	case Host:
		return NewHostDriver
	default:
		return nil
	}
}

func convertEnvToMap(env []string) map[string]string {
	// convert env to map for processing
	envMap := make(map[string]string)
	for _, varPair := range env {
		pair := strings.Split(varPair, "=")
		envMap[pair[0]] = pair[1]
	}
	return envMap
}
