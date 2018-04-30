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

	"github.com/GoogleCloudPlatform/runtimes-common/ctc_lib"
	"github.com/GoogleContainerTools/container-structure-test/pkg/drivers"
	"github.com/GoogleContainerTools/container-structure-test/pkg/output"
	"github.com/GoogleContainerTools/container-structure-test/pkg/types"
	"github.com/GoogleContainerTools/container-structure-test/pkg/types/unversioned"
	"github.com/GoogleContainerTools/container-structure-test/pkg/utils"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

const warnMessage = `WARNING: Using the host driver runs tests directly on the machine
that this binary is being run on, and can potentially corrupt your system.
Be sure you know what you're doing before continuing!

Continue? (y/n)`

var totalTests int
var TestReportFile *os.File

var driverImpl func(drivers.DriverConfig) (drivers.Driver, error)
var args *drivers.DriverConfig

var Channel = make(chan interface{}, 1)

var TestCmd = &ctc_lib.ContainerToolListCommand{
	ContainerToolCommandBase: &ctc_lib.ContainerToolCommandBase{
		Command: &cobra.Command{
			Use:   "test",
			Short: "Runs the tests",
			Long:  `Runs the tests`,
			Args: func(cmd *cobra.Command, _ []string) error {
				return validateArgs()
			},
		},
		Phase:           "stable",
		DefaultTemplate: output.StructureTestsTemplate,
		TemplateFuncMap: initTemplateFuncMap(),
	},
	OutputList:      make([]interface{}, 0),
	SummaryTemplate: output.SummaryTemplate,
	SummaryObject:   &unversioned.SummaryObject{},
	StreamO: func(command *cobra.Command, args []string) {
		Run()
	},
	Stream: Channel,
	TotalO: func(list []interface{}) (interface{}, error) {
		totalPass := 0
		totalFail := 0
		errStrings := make([]string, 0)
		var err error
		for _, r := range list {
			value, ok := r.(*unversioned.TestResult)
			if !ok {
				errStrings = append(errStrings, fmt.Sprintf("unexpected value %v in list", value))
				ctc_lib.Log.Errorf("unexpected value %v in list", value)
				continue
			}
			if value.IsPass() {
				totalPass++
			} else {
				totalFail++
			}
		}
		if totalFail > 0 {
			errStrings = append(errStrings, "Test(s) FAIL")
		}
		if len(errStrings) > 0 {
			err = fmt.Errorf(strings.Join(errStrings, "\n"))
		}

		return unversioned.SummaryObject{
			Total: totalFail + totalPass,
			Pass:  totalPass,
			Fail:  totalFail,
		}, err
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

func RunTests() {
	for _, file := range configFiles {
		Channel <- output.Banner(file)
		tests, err := Parse(file)
		if err != nil {
			ctc_lib.Log.Errorf("Error parsing config file: %s", err)
			continue // Continue with other config files
		}
		tests.RunAll(Channel, file)
	}
	close(Channel)
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
	case strings.HasSuffix(fp, ".yml"):
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

func Run() {
	args = &drivers.DriverConfig{
		Image:    imagePath,
		Save:     save,
		Metadata: metadata,
	}

	var err error

	if pull {
		if driver != drivers.Docker {
			ctc_lib.Log.Fatal("Image pull not supported when not using Docker driver")
		}
		var repository, tag string
		parts := strings.Split(imagePath, ":")
		if len(parts) < 2 {
			ctc_lib.Log.Fatal("Please provide specific tag for image")
		}
		repository = parts[0]
		tag = parts[1]
		client, err := docker.NewClientFromEnv()
		if err = client.PullImage(docker.PullImageOptions{
			Repository:   repository,
			Tag:          tag,
			OutputStream: os.Stdout,
		}, docker.AuthConfiguration{}); err != nil {
			ctc_lib.Log.Fatalf("Error pulling remote image %s: %s", imagePath, err.Error())
		}
	}

	if driver == drivers.Host && !utils.UserConfirmation(warnMessage, force) {
		ctc_lib.Log.Fatalf("User Aborted")
	}

	driverImpl = drivers.InitDriverImpl(driver)
	if driverImpl == nil {
		ctc_lib.Log.Fatalf("Unsupported driver type: %s", driver)
	}
	if err != nil {
		ctc_lib.Log.Fatal(err.Error())
	}
	go RunTests()
}

func init() {
	RootCmd.AddCommand(TestCmd)
	TestCmd.Flags().StringVar(&imagePath, "image", "", "path to test image")
	TestCmd.Flags().StringVar(&driver, "driver", "docker", "driver to use when running tests")
	TestCmd.Flags().StringVar(&metadata, "metadata", "", "path to image metadata file")
	TestCmd.Flags().BoolVar(&pull, "pull", false, "force a pull of the image before running tests")
	TestCmd.Flags().BoolVar(&save, "save", false, "preserve created containers after test run")
	TestCmd.Flags().BoolVar(&quiet, "quiet", false, "flag to suppress output")
	TestCmd.Flags().BoolVar(&force, "force", false, "force run of host driver (without user prompt)")

	TestCmd.Flags().StringArrayVar(&configFiles, "config", []string{}, "test config files")
	TestCmd.Flags().StringVar(&testReport, "test-report", "", `Generate Json Test Report and write it to specified filename.
Implies --jsonOutput flag`)
}

func isQuiet() bool {
	return quiet
}
func initTemplateFuncMap() map[string]interface{} {
	var templateFuncMap = map[string]interface{}{
		"isQuiet": isQuiet,
	}
	// Copy over all the Output utilities to TemplateFuncMap
	for k, v := range output.TemplateMap {
		templateFuncMap[k] = v
	}
	return templateFuncMap
}
