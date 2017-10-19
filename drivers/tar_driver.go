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
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	pkgutil "github.com/GoogleCloudPlatform/container-diff/pkg/util"
	"github.com/GoogleCloudPlatform/runtimes-common/structure_tests/types/unversioned"
)

type TarDriver struct {
	Image pkgutil.Image
}

func NewTarDriver(imageName string) (Driver, error) {
	// if the image is in the local daemon, we should be using the docker driver anyway.
	// only try remote.

	var prepper pkgutil.Prepper
	if pkgutil.IsTar(imageName) {
		prepper = pkgutil.TarPrepper{
			Source: imageName,
		}
	} else {
		prepper = pkgutil.CloudPrepper{
			Source: imageName,
		}
	}

	image, err := prepper.GetImage()
	if err != nil {
		// didn't find image anywhere; exit
		return nil, err
	}
	return &TarDriver{
		Image: image,
	}, nil
}

func (d *TarDriver) Destroy() {
	pkgutil.CleanupImage(d.Image)
}

func (d *TarDriver) Setup(t *testing.T, envVars []unversioned.EnvVar, fullCommand []unversioned.Command) {
	// this driver is unable to process commands, inform user and fail.
	t.Fatal("Tar driver is unable to process commands, please use a different driver")
}

func (d *TarDriver) ProcessCommand(t *testing.T, envVars []unversioned.EnvVar, fullCommand []string) (string, string, int) {
	// this driver is unable to process commands, inform user and fail.
	t.Fatal("Tar driver is unable to process commands, please use a different driver")
	return "", "", 0
}

func (d *TarDriver) StatFile(t *testing.T, path string) (os.FileInfo, error) {
	return os.Stat(filepath.Join(d.Image.FSPath, path))
}

func (d *TarDriver) ReadFile(t *testing.T, path string) ([]byte, error) {
	return ioutil.ReadFile(filepath.Join(d.Image.FSPath, path))
}

func (d *TarDriver) ReadDir(t *testing.T, path string) ([]os.FileInfo, error) {
	return ioutil.ReadDir(filepath.Join(d.Image.FSPath, path))
}

func (d *TarDriver) GetConfig(t *testing.T) (unversioned.Config, error) {
	// docker provides these as maps (since they can be mapped in docker run commands)
	// since this will never be the case when built through a dockerfile, we convert to list of strings
	volumes := []string{}
	for v := range d.Image.Config.Config.Volumes {
		volumes = append(volumes, v)
	}

	ports := []string{}
	for p := range d.Image.Config.Config.ExposedPorts {
		// docker always appends the protocol to the port, so this is safe
		ports = append(ports, strings.Split(p, "/")[0])
	}

	return unversioned.Config{
		Env:          convertEnvToMap(d.Image.Config.Config.Env),
		Entrypoint:   d.Image.Config.Config.Entrypoint,
		Cmd:          d.Image.Config.Config.Cmd,
		Volumes:      volumes,
		Workdir:      d.Image.Config.Config.Workdir,
		ExposedPorts: ports,
	}, nil
}
