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
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"testing"

	"bytes"

	"github.com/GoogleCloudPlatform/container-structure-test/types/unversioned"
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

func (d *HostDriver) Setup(t *testing.T, envVars []unversioned.EnvVar, fullCommands [][]string) {
	// since we're running on the host, we'll provide an optional teardown field for`
	// each test that will allow users to undo the setup they did.
	// keep track of the original env vars so we can reset later.
	d.GlobalVars = SetEnvVars(t, envVars)
	for _, cmd := range fullCommands {
		d.ProcessCommand(t, nil, cmd)
	}
}

func (d *HostDriver) Teardown(t *testing.T, envVars []unversioned.EnvVar, fullCommands [][]string) {
	// since we're running on the host, we'll provide an optional teardown field for each test that
	// will allow users to undo the setup they did.
	ResetEnvVars(t, d.GlobalVars)
	for _, cmd := range fullCommands {
		d.ProcessCommand(t, nil, cmd)
	}
}

// given a list of environment variable key/value pairs, set these in the current environment.
// also, keep track of the previous values of these vars to reset after test execution.
func SetEnvVars(t *testing.T, vars []unversioned.EnvVar) []unversioned.EnvVar {
	if vars == nil {
		return nil
	}
	var originalVars []unversioned.EnvVar
	for _, env_var := range vars {
		originalVars = append(originalVars, unversioned.EnvVar{env_var.Key, os.Getenv(env_var.Key)})
		if err := os.Setenv(env_var.Key, os.ExpandEnv(env_var.Value)); err != nil {
			t.Fatalf("error setting env var: %s", err)
		}
	}
	return originalVars
}

func ResetEnvVars(t *testing.T, vars []unversioned.EnvVar) {
	if vars == nil {
		return
	}
	for _, env_var := range vars {
		var err error
		if env_var.Value == "" {
			// if the previous value was empty string, the variable did not
			// exist in the environment; unset it
			err = os.Unsetenv(env_var.Key)
		} else {
			// otherwise, set it back to its previous value
			err = os.Setenv(env_var.Key, env_var.Value)
		}
		if err != nil {
			t.Fatalf("error resetting env var: %s", err)
		}
	}
}

func (d *HostDriver) ProcessCommand(t *testing.T, envVars []unversioned.EnvVar, fullCommand []string) (string, string, int) {
	originalVars := SetEnvVars(t, envVars)
	defer ResetEnvVars(t, originalVars)
	cmd := exec.Command(fullCommand[0], fullCommand[1:]...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	exitCode := 0

	if err := cmd.Start(); err != nil {
		t.Errorf("error starting command: %v", err)
	}

	if err := cmd.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				exitCode = status.ExitStatus()
			}
		} else {
			t.Errorf("Error when retrieving exit code: %v", err)
		}
	}
	t.Logf("command output: %s", stdout.String())
	return stdout.String(), stderr.String(), exitCode
}

func (d *HostDriver) StatFile(t *testing.T, path string) (os.FileInfo, error) {
	return os.Stat(path)
}

func (d *HostDriver) ReadFile(t *testing.T, path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}

func (d *HostDriver) ReadDir(t *testing.T, path string) ([]os.FileInfo, error) {
	return ioutil.ReadDir(path)
}

func (d *HostDriver) GetConfig(t *testing.T) (unversioned.Config, error) {
	file, err := ioutil.ReadFile(d.ConfigPath)
	if err != nil {
		t.Errorf("error retrieving config")
		return unversioned.Config{}, err
	}

	var metadata unversioned.FlattenedMetadata

	json.Unmarshal(file, &metadata)
	config := metadata.Config

	// // docker provides these as maps (since they can be mapped in docker run commands)
	// // since this will never be the case when built through a dockerfile, we convert to list of strings
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
		Env:          convertEnvToMap(config.Env),
		Entrypoint:   config.Entrypoint,
		Cmd:          config.Cmd,
		Volumes:      volumes,
		Workdir:      config.Workdir,
		ExposedPorts: ports,
	}, nil
}
