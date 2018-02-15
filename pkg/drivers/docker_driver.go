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
	"archive/tar"
	"bufio"
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/GoogleCloudPlatform/container-structure-test/pkg/types/unversioned"
	"github.com/GoogleCloudPlatform/container-structure-test/pkg/utils"
	docker "github.com/fsouza/go-dockerclient"
)

type DockerDriver struct {
	originalImage string
	currentImage  string
	cli           docker.Client
	env           map[string]string
	save          bool
}

func NewDockerDriver(args DriverConfig) (Driver, error) {
	newCli, err := docker.NewClientFromEnv()
	if err != nil {
		return nil, err
	}
	return &DockerDriver{
		originalImage: args.Image,
		currentImage:  args.Image,
		cli:           *newCli,
		env:           nil,
		save:          args.Save,
	}, nil
}

func (d *DockerDriver) Destroy() {
	// since intermediate images are chained, removing the most current
	// image (that isn't the original) removes all previous ones as well.
	if d.currentImage != d.originalImage {
		if err := d.cli.RemoveImage(d.currentImage); err != nil {
			logrus.Warnf("error removing image: %s", err)
		}
	}
}

func (d *DockerDriver) Setup(envVars []unversioned.EnvVar, fullCommands [][]string) error {
	env := d.processEnvVars(envVars)
	for _, cmd := range fullCommands {
		img, err := d.runAndCommit(env, cmd)
		if err != nil {
			return err
		}
		d.currentImage = img
	}
	return nil
}

func (d *DockerDriver) Teardown(envVars []unversioned.EnvVar, fullCommands [][]string) error {
	// since we create a new driver for each test, skip teardown commands
	logrus.Debug("Docker driver does not support teardown commands, since each test gets a new driver. Skipping commands.")
	return nil
}

func (d *DockerDriver) ProcessCommand(envVars []unversioned.EnvVar, fullCommand []string) (string, string, int, error) {
	var env []string
	for _, envVar := range envVars {
		env = append(env, fmt.Sprintf("%s=%s", envVar.Key, envVar.Value))
	}
	stdout, stderr, exitCode, err := d.exec(env, fullCommand)
	if err != nil {
		return "", "", -1, err
	}

	if stdout != "" {
		logrus.Infof("stdout: %s", stdout)
	}
	if stderr != "" {
		logrus.Infof("stderr: %s", stderr)
	}
	return stdout, stderr, exitCode, nil
}

func retrieveEnv(d *DockerDriver) func(string) string {
	return func(envVar string) string {
		var env map[string]string
		if env == nil {
			image, err := d.cli.InspectImage(d.currentImage)
			if err != nil {
				return ""
			}
			// convert env to map for processing
			env = convertEnvToMap(image.Config.Env)
		}
		return env[envVar]
	}
}

func (d *DockerDriver) retrieveEnvVar(envVar string) string {
	// since we're only retrieving these during processing, we can use a closure to cache this
	return retrieveEnv(d)(envVar)
}

func (d *DockerDriver) processEnvVars(vars []unversioned.EnvVar) []string {
	if len(vars) == 0 {
		return nil
	}

	env := []string{}

	for _, envVar := range vars {
		expandedVal := os.Expand(envVar.Value, d.retrieveEnvVar)
		env = append(env, fmt.Sprintf("%s=%s", envVar.Key, expandedVal))
	}
	return env
}

// copies a tar archive starting at the specified path from the image, and returns
// a tar reader which can be used to iterate through its contents and retrieve metadata
func (d *DockerDriver) retrieveTar(path string) (*tar.Reader, error) {
	// this contains a placeholder command which does not get run, since
	// the client doesn't allow creating a container without a command.
	container, err := d.cli.CreateContainer(docker.CreateContainerOptions{
		Config: &docker.Config{
			Image: d.currentImage,
			Cmd:   []string{utils.NoopCommand},
		},
		HostConfig:       nil,
		NetworkingConfig: nil,
	})
	if err != nil {
		return nil, errors.Wrap(err, "Error creating container")
	}
	defer d.removeContainer(container.ID)

	var b bytes.Buffer
	stream := bufio.NewWriter(&b)

	if err = d.cli.DownloadFromContainer(container.ID, docker.DownloadFromContainerOptions{
		OutputStream: stream,
		Path:         path,
	}); err != nil {
		return nil, errors.Wrap(err, "Error retrieving file from container")
	}
	if err = stream.Flush(); err != nil {
		return nil, err
	}
	return tar.NewReader(bytes.NewReader(b.Bytes())), nil
}

func (d *DockerDriver) StatFile(target string) (os.FileInfo, error) {
	reader, err := d.retrieveTar(target)
	if err != nil {
		return nil, err
	}
	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		switch header.Typeflag {
		case tar.TypeDir, tar.TypeReg, tar.TypeLink, tar.TypeSymlink:
			if filepath.Clean(header.Name) == path.Base(target) {
				return header.FileInfo(), nil
			}
		default:
			continue
		}
	}
	return nil, fmt.Errorf("File %s not found in image", target)
}

func (d *DockerDriver) ReadFile(target string) ([]byte, error) {
	reader, err := d.retrieveTar(target)
	if err != nil {
		return nil, err
	}
	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		switch header.Typeflag {
		case tar.TypeDir:
			if filepath.Clean(header.Name) == path.Base(target) {
				return nil, fmt.Errorf("Cannot read specified path: %s is a directory, not a file", target)
			}
		case tar.TypeSymlink:
			return d.ReadFile(header.Linkname)
		case tar.TypeReg, tar.TypeLink:
			if filepath.Clean(header.Name) == path.Base(target) {
				var b bytes.Buffer
				stream := bufio.NewWriter(&b)
				io.Copy(stream, reader)
				return b.Bytes(), nil
			}
		default:
			continue
		}
	}
	return nil, fmt.Errorf("File %s not found in image", target)
}

func (d *DockerDriver) ReadDir(target string) ([]os.FileInfo, error) {
	reader, err := d.retrieveTar(target)
	if err != nil {
		return nil, err
	}
	var infos []os.FileInfo
	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if header.Typeflag == tar.TypeDir {
			// we only want top level dirs here, no recursion. to get these, remove
			// trailing separator and split on separator. there should only be two parts.
			parts := strings.Split(strings.TrimSuffix(header.Name, string(os.PathSeparator)), string(os.PathSeparator))
			if len(parts) == 2 {
				infos = append(infos, header.FileInfo())
			}
		}
	}
	return infos, nil
}

// This method takes a command (in the form of a list of args), and does the following:
// 1) creates a container, based on the "current latest" image, with the command set as
// the command to run when the container starts
// 2) starts the container
// 3) commits the container with its changes to a new image,
// and sets that image as the new "current image"
func (d *DockerDriver) runAndCommit(env []string, command []string) (string, error) {
	container, err := d.cli.CreateContainer(docker.CreateContainerOptions{
		Config: &docker.Config{
			Image:        d.currentImage,
			Env:          env,
			Cmd:          command,
			Entrypoint:   []string{},
			AttachStdout: true,
			AttachStderr: true,
		},
		HostConfig:       nil,
		NetworkingConfig: nil,
	})
	if err != nil {
		return "", errors.Wrap(err, "Error creating container")
	}

	if err = d.cli.StartContainer(container.ID, nil); err != nil {
		return "", errors.Wrap(err, "Error creating container")
	}

	if _, err = d.cli.WaitContainer(container.ID); err != nil {
		return "", errors.Wrap(err, "Error when waiting for container")
	}

	image, err := d.cli.CommitContainer(docker.CommitContainerOptions{
		Container: container.ID,
	})

	if err != nil {
		return "", errors.Wrap(err, "Error committing container")
	}

	if !d.save {
		if err = d.cli.RemoveContainer(docker.RemoveContainerOptions{
			ID: container.ID,
		}); err != nil {
			logrus.Warnf("Error when removing container %s: %s", container.ID, err.Error())
		}
	}

	d.currentImage = image.ID
	return image.ID, nil
}

func (d *DockerDriver) exec(env []string, command []string) (string, string, int, error) {
	// first, start container from the current image
	container, err := d.cli.CreateContainer(docker.CreateContainerOptions{
		Config: &docker.Config{
			Image:        d.currentImage,
			Env:          env,
			Cmd:          command,
			Entrypoint:   []string{},
			AttachStdout: true,
			AttachStderr: true,
		},
		HostConfig:       nil,
		NetworkingConfig: nil,
	})
	if err != nil {
		return "", "", -1, errors.Wrap(err, "Error creating container")
	}
	defer d.removeContainer(container.ID)

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	if err = d.cli.StartContainer(container.ID, nil); err != nil {
		return "", "", -1, errors.Wrap(err, "Error creating container")
	}

	//TODO(nkubala): look into adding timeout
	exitCode, err := d.cli.WaitContainer(container.ID)
	if err != nil {
		return "", "", -1, errors.Wrap(err, "Error when waiting for container")
	}

	if err = d.cli.Logs(docker.LogsOptions{
		Container:    container.ID,
		OutputStream: stdout,
		ErrorStream:  stderr,
		Stdout:       true,
		Stderr:       true,
	}); err != nil {
		return "", "", -1, errors.Wrap(err, "Error retrieving container logs")
	}

	return stdout.String(), stderr.String(), exitCode, nil
}

func (d *DockerDriver) GetConfig() (unversioned.Config, error) {
	img, err := d.cli.InspectImage(d.currentImage)
	if err != nil {
		return unversioned.Config{}, errors.Wrap(err, "Error when inspecting image")
	}

	// docker provides these as maps (since they can be mapped in docker run commands)
	// since this will never be the case when built through a dockerfile, we convert to list of strings
	volumes := []string{}
	for v := range img.Config.Volumes {
		volumes = append(volumes, v)
	}

	ports := []string{}
	for p := range img.Config.ExposedPorts {
		ports = append(ports, p.Port())
	}

	return unversioned.Config{
		Env:          convertEnvToMap(img.Config.Env),
		Entrypoint:   img.Config.Entrypoint,
		Cmd:          img.Config.Cmd,
		Volumes:      volumes,
		Workdir:      img.Config.WorkingDir,
		ExposedPorts: ports,
	}, nil
}

func (d *DockerDriver) removeContainer(containerID string) {
	if d.save {
		return
	}
	if err := d.cli.RemoveContainer(docker.RemoveContainerOptions{
		ID: containerID,
	}); err != nil {
		logrus.Warnf("Error when removing container %s: %s", containerID, err.Error())
	}
}
