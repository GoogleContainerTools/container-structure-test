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

	pkgutil "github.com/GoogleContainerTools/container-diff/pkg/util"
	"github.com/GoogleContainerTools/container-structure-test/pkg/types/unversioned"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type TarDriver struct {
	Image pkgutil.Image
	Save  bool
}

func NewTarDriver(args DriverConfig) (Driver, error) {
	if pkgutil.IsTar(args.Image) {
		// tar provided, so don't provide any prefix. container-diff can figure this out.
		image, err := pkgutil.GetImageForName(args.Image)
		if err != nil {
			return nil, errors.Wrap(err, "processing tar image reference")
		}
		return &TarDriver{
			Image: image,
			Save:  args.Save,
		}, nil
	}
	// try the local docker daemon first
	image, err := pkgutil.GetImageForName("daemon://" + args.Image)
	if err == nil {
		logrus.Debugf("image found in local docker daemon")
		return &TarDriver{
			Image: image,
			Save:  args.Save,
		}, nil
	}

	// image not found in local daemon, so try remote.
	logrus.Infof("unable to retrieve image locally: %s", err)
	image, err = pkgutil.GetImageForName("remote://" + args.Image)
	if err != nil {
		return nil, errors.Wrap(err, "retrieving image")
	}
	return &TarDriver{
		Image: image,
		Save:  args.Save,
	}, nil
}

func (d *TarDriver) Destroy() {
	if !d.Save {
		pkgutil.CleanupImage(d.Image)
	}
}

func (d *TarDriver) SetEnv(envVars []unversioned.EnvVar) error {
	configFile, err := d.Image.Image.ConfigFile()
	if err != nil {
		return errors.Wrap(err, "retrieving image config")
	}
	config := configFile.Config
	env := convertSliceToMap(config.Env)
	for _, envVar := range envVars {
		env[envVar.Key] = envVar.Value
	}
	newConfig := v1.Config{
		AttachStderr:    config.AttachStderr,
		AttachStdin:     config.AttachStdin,
		AttachStdout:    config.AttachStdout,
		Cmd:             config.Cmd,
		Domainname:      config.Domainname,
		Entrypoint:      config.Entrypoint,
		Env:             convertMapToSlice(env),
		Hostname:        config.Hostname,
		Image:           config.Image,
		Labels:          config.Labels,
		OnBuild:         config.OnBuild,
		OpenStdin:       config.OpenStdin,
		StdinOnce:       config.StdinOnce,
		Tty:             config.Tty,
		User:            config.User,
		Volumes:         config.Volumes,
		WorkingDir:      config.WorkingDir,
		ExposedPorts:    config.ExposedPorts,
		ArgsEscaped:     config.ArgsEscaped,
		NetworkDisabled: config.NetworkDisabled,
		MacAddress:      config.MacAddress,
		StopSignal:      config.StopSignal,
		Shell:           config.Shell,
	}
	newImg, err := mutate.Config(d.Image.Image, newConfig)
	if err != nil {
		return errors.Wrap(err, "setting new config on image")
	}
	newImage := pkgutil.Image{
		Image:  newImg,
		Source: d.Image.Source,
		FSPath: d.Image.FSPath,
		Digest: d.Image.Digest,
		Layers: d.Image.Layers,
	}
	d.Image = newImage
	return nil
}

func (d *TarDriver) Setup(_ []unversioned.EnvVar, _ [][]string) error {
	// this driver is unable to process commands, inform user and fail.
	return errors.New("Tar driver is unable to process commands, please use a different driver")
}

func (d *TarDriver) Teardown(_ [][]string) error {
	return errors.New("Tar driver is unable to process commands, please use a different driver")
}

func (d *TarDriver) ProcessCommand(_ []unversioned.EnvVar, _ []string) (string, string, int, error) {
	// this driver is unable to process commands, inform user and fail.
	return "", "", -1, errors.New("Tar driver is unable to process commands, please use a different driver")
}

func (d *TarDriver) StatFile(path string) (os.FileInfo, error) {
	return os.Lstat(filepath.Join(d.Image.FSPath, path))
}

func (d *TarDriver) ReadFile(path string) ([]byte, error) {
	return ioutil.ReadFile(filepath.Join(d.Image.FSPath, path))
}

func (d *TarDriver) ReadDir(path string) ([]os.FileInfo, error) {
	return ioutil.ReadDir(filepath.Join(d.Image.FSPath, path))
}

func (d *TarDriver) GetConfig() (unversioned.Config, error) {
	configFile, err := d.Image.Image.ConfigFile()
	if err != nil {
		return unversioned.Config{}, errors.Wrap(err, "retrieving config file")
	}
	config := configFile.Config

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
