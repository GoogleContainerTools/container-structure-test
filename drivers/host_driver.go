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
	"strings"
	"testing"

	"github.com/GoogleCloudPlatform/container-structure-test/types/unversioned"
)

type HostDriver struct {
	// Root       string // defaults to /
	ConfigPath string // path to image metadata config on host fs
}

func NewHostDriver(args DriverConfig) (Driver, error) {
	return &HostDriver{
		// Root: root,
		ConfigPath: args.Metadata,
	}, nil
}

func (d *HostDriver) Destroy() {
	// since we're running on the host, don't do anything
}

func (d *HostDriver) Setup(t *testing.T, envVars []unversioned.EnvVar, fullCommand []unversioned.Command) {
	// TODO(nkubala): implement
}

func (d *HostDriver) ProcessCommand(t *testing.T, envVars []unversioned.EnvVar, fullCommand []string) (string, string, int) {
	// TODO(nkubala): implement
	return "", "", 0
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
