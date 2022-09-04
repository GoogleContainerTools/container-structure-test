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
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/GoogleContainerTools/container-structure-test/pkg/types/unversioned"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type HostDriver struct {
	ConfigPath string // path to image metadata config on host fs
	GlobalVars []unversioned.EnvVar
}

func NewHostDriver(args DriverConfig) (Driver, error) {
	return &HostDriver{
		ConfigPath: args.Metadata,
	}, nil
}

func (d *HostDriver) Destroy() {
	// since we're running on the host, don't do anything
}

func (d *HostDriver) Setup(envVars []unversioned.EnvVar, fullCommands [][]string) error {
	// since we're running on the host, we'll provide an optional teardown field for
	// each test that will allow users to undo the setup they did.
	// keep track of the original env vars so we can reset later.
	d.GlobalVars = SetEnvVars(envVars)
	for _, cmd := range fullCommands {
		_, _, _, err := d.ProcessCommand(nil, cmd)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *HostDriver) Teardown(fullCommands [][]string) error {
	// since we're running on the host, we'll provide an optional teardown field for each test that
	// will allow users to undo the setup they did.
	ResetEnvVars(d.GlobalVars)
	for _, cmd := range fullCommands {
		_, _, _, err := d.ProcessCommand(nil, cmd)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *HostDriver) SetEnv(envVars []unversioned.EnvVar) error {
	for _, envVar := range envVars {
		if err := os.Setenv(envVar.Key, os.ExpandEnv(envVar.Value)); err != nil {
			return err
		}
	}
	return nil
}

// given a list of environment variable key/value pairs, set these in the current environment.
// also, keep track of the previous values of these vars to reset after test execution.
func SetEnvVars(envVars []unversioned.EnvVar) []unversioned.EnvVar {
	var originalVars []unversioned.EnvVar
	for _, envVar := range envVars {
		originalVars = append(originalVars, unversioned.EnvVar{Key: envVar.Key, Value: os.Getenv(envVar.Key), IsRegex: envVar.IsRegex})
		if err := os.Setenv(envVar.Key, os.ExpandEnv(envVar.Value)); err != nil {
			logrus.Errorf("Error setting env var: %s", err)
		}
	}
	return originalVars
}

func ResetEnvVars(envVars []unversioned.EnvVar) {
	for _, envVar := range envVars {
		var err error
		if envVar.Value == "" {
			// if the previous value was empty string, the variable did not
			// exist in the environment; unset it
			err = os.Unsetenv(envVar.Key)
		} else {
			// otherwise, set it back to its previous value
			err = os.Setenv(envVar.Key, envVar.Value)
		}
		if err != nil {
			logrus.Errorf("error resetting env var: %s", err)
		}
	}
}

func (d *HostDriver) ProcessCommand(envVars []unversioned.EnvVar, fullCommand []string) (string, string, int, error) {
	originalVars := SetEnvVars(envVars)
	defer ResetEnvVars(originalVars)
	cmd := exec.Command(fullCommand[0], fullCommand[1:]...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	exitCode := 0

	if err := cmd.Start(); err != nil {
		logrus.Fatalf("error starting command: %v", err)
	}

	if err := cmd.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				exitCode = status.ExitStatus()
			}
		} else {
			return "", "", -1, errors.Wrap(err, "Error when retrieving exit code")
		}
	}
	logrus.Debugf("command output: %s", stdout.String())
	return stdout.String(), stderr.String(), exitCode, nil
}

func (d *HostDriver) StatFile(path string) (os.FileInfo, error) {
	return os.Lstat(path)
}

func (d *HostDriver) ReadFile(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}

func (d *HostDriver) ReadDir(path string) ([]os.FileInfo, error) {
	return ioutil.ReadDir(path)
}

func (d *HostDriver) GetConfig() (unversioned.Config, error) {
	file, err := ioutil.ReadFile(d.ConfigPath)
	if err != nil {
		return unversioned.Config{}, errors.Wrap(err, "Error retrieving config")
	}

	var metadata v1.ConfigFile

	json.Unmarshal(file, &metadata)
	config := metadata.Config

	// docker provides these as maps (since they can be mapped in docker run commands)
	// since this will never be the case when built through a dockerfile, we convert to list of strings
	volumes := []string{}
	for v := range config.Volumes {
		volumes = append(volumes, v)
	}

	ports := []string{}
	for p := range config.ExposedPorts {
		// docker always appends the protocol to the port, so this is safe
		ports = append(ports, strings.Split(p, "/")[0])
	}

	return unversioned.Config{
		Env:          convertSliceToMap(config.Env),
		Entrypoint:   config.Entrypoint,
		Cmd:          config.Cmd,
		Volumes:      volumes,
		Workdir:      config.WorkingDir,
		ExposedPorts: ports,
		Labels:       config.Labels,
		User:         config.User,
	}, nil
}
