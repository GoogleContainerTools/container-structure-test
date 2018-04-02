// Copyright 2018 Google Inc. All rights reserved.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/GoogleCloudPlatform/container-structure-test/pkg/drivers"
	"github.com/GoogleCloudPlatform/container-structure-test/pkg/output"
	"github.com/GoogleCloudPlatform/container-structure-test/pkg/types"
	"github.com/GoogleCloudPlatform/container-structure-test/pkg/utils"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

const warnMessage = `WARNING: Using the host driver runs tests directly on the machine 
that this binary is being run on, and can potentially corrupt your system.
Be sure you know what you're doing before continuing!

Continue? (y/n)`

var totalTests int

var driverImpl func(drivers.DriverConfig) (drivers.Driver, error)
var args *drivers.DriverConfig

var TestCmd = &cobra.Command{
	Use:   "test",
	Short: "Runs the tests",
	Long:  `Runs the tests`,
	Args: func(cmd *cobra.Command, _ []string) error {
		return validateArgs()
	},
	Run: func(cmd *cobra.Command, _ []string) {
		out := &output.OutWriter{
			Verbose: verbose,
			Quiet:   quiet,
		}
		if pass := Run(out); pass {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	},
}

func validateArgs() error {
	if driver == drivers.Host {
		if metadata == "" {
			return fmt.Errorf("Please provide path to image metadata file")
		}
		if imagePath != "" {
			return fmt.Errorf("Cannot provide both image path and metadata file")
		}
	} else {
		if imagePath == "" {
			return fmt.Errorf("Please supply path to image or tarball to test against")
		}
		if metadata != "" {
			return fmt.Errorf("Cannot provide both image path and metadata file")
		}
	}
	if len(configFiles) == 0 {
		return fmt.Errorf("Please provide at least one test config file")
	}
	return nil
}

func RunTests(out *output.OutWriter) bool {
	pass := true
	for _, file := range configFiles {
		out.Banner(file)
		tests, err := Parse(file)
		if err != nil {
			logrus.Fatalf("Error parsing config file: %s", err)
		}
		pass = pass && out.FinalResults(tests.RunAll(out))
	}
	return pass
}

func Parse(fp string) (types.StructureTest, error) {
	testContents, err := ioutil.ReadFile(fp)
	if err != nil {
		return nil, err
	}

	// We first have to unmarshal to determine the schema version, then we unmarshal again
	// to do the full parse.
	var unmarshal types.Unmarshaller
	var strictUnmarshal types.Unmarshaller
	var versionHolder types.SchemaVersion

	switch {
	case strings.HasSuffix(fp, ".json"):
		unmarshal = json.Unmarshal
		strictUnmarshal = json.Unmarshal
	case strings.HasSuffix(fp, ".yaml"):
		unmarshal = yaml.Unmarshal
		strictUnmarshal = yaml.UnmarshalStrict
	default:
		return nil, errors.New("Please provide valid JSON or YAML config file")
	}

	if err := unmarshal(testContents, &versionHolder); err != nil {
		return nil, err
	}

	version := versionHolder.SchemaVersion
	if version == "" {
		return nil, errors.New("Please provide JSON schema version")
	}

	var st types.StructureTest
	if schemaVersion, ok := types.SchemaVersions[version]; ok {
		st = schemaVersion()
	} else {
		return nil, errors.New("Unsupported schema version: " + version)
	}

	strictUnmarshal(testContents, st)

	tests, _ := st.(types.StructureTest) //type assertion
	tests.SetDriverImpl(driverImpl, *args)
	return tests, nil
}

func Run(out *output.OutWriter) bool {
	args = &drivers.DriverConfig{
		Image:    imagePath,
		Save:     save,
		Metadata: metadata,
	}

	var err error

	if pull {
		if driver != drivers.Docker {
			logrus.Fatal("Image pull not supported when not using Docker driver")
		}
		var repository, tag string
		parts := strings.Split(imagePath, ":")
		if len(parts) < 2 {
			logrus.Fatal("Please provide specific tag for image")
		}
		repository = parts[0]
		tag = parts[1]
		client, err := docker.NewClientFromEnv()
		if err = client.PullImage(docker.PullImageOptions{
			Repository:   repository,
			Tag:          tag,
			OutputStream: os.Stdout,
		}, docker.AuthConfiguration{}); err != nil {
			logrus.Fatalf("Error pulling remote image %s: %s", imagePath, err.Error())
		}
	}

	if driver == drivers.Host && !utils.UserConfirmation(warnMessage, force) {
		return false
	}

	driverImpl = drivers.InitDriverImpl(driver)
	if driverImpl == nil {
		logrus.Fatalf("Unsupported driver type: %s", driver)
	}
	if err != nil {
		logrus.Fatal(err.Error())
	}
	logrus.Infof("Using driver %s\n", driver)
	return RunTests(out)
}

func init() {
	RootCmd.AddCommand(TestCmd)
	TestCmd.Flags().StringVar(&imagePath, "image", "", "path to test image")
	TestCmd.Flags().StringVar(&driver, "driver", "docker", "driver to use when running tests")
	TestCmd.Flags().StringVar(&metadata, "metadata", "", "path to image metadata file")
	TestCmd.Flags().BoolVar(&pull, "pull", false, "force a pull of the image before running tests")
	TestCmd.Flags().BoolVar(&save, "save", false, "preserve created containers after test run")
	TestCmd.Flags().BoolVar(&force, "force", false, "force run of host driver (without command line input)")
	TestCmd.Flags().BoolVar(&quiet, "quiet", false, "flag to suppress output")
	TestCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "verbose testing output")
	TestCmd.Flags().StringArrayVar(&configFiles, "config", []string{}, "test config files")
}
