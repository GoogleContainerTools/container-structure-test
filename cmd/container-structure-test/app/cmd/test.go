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
	"fmt"
	"io"
	"io/ioutil"

	"os"
	"regexp"
	"strings"

	"github.com/GoogleContainerTools/container-structure-test/cmd/container-structure-test/app/cmd/test"

	"github.com/GoogleContainerTools/container-structure-test/pkg/color"
	"github.com/GoogleContainerTools/container-structure-test/pkg/config"
	"github.com/GoogleContainerTools/container-structure-test/pkg/drivers"
	"github.com/GoogleContainerTools/container-structure-test/pkg/output"
	"github.com/GoogleContainerTools/container-structure-test/pkg/types/unversioned"
	"github.com/GoogleContainerTools/container-structure-test/pkg/utils"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const warnMessage = `WARNING: Using the host driver runs tests directly on the machine
that this binary is being run on, and can potentially corrupt your system.
Be sure you know what you're doing before continuing!

Continue? (y/n)`

var totalTests int
var testReportFile *os.File

var (
	opts = &config.StructureTestOptions{}

	args       *drivers.DriverConfig
	driverImpl func(drivers.DriverConfig) (drivers.Driver, error)
)

func NewCmdTest(out io.Writer) *cobra.Command {
	var testCmd = &cobra.Command{
		Use:   "test",
		Short: "Runs the tests",
		Long:  `Runs the tests`,
		Args: func(cmd *cobra.Command, _ []string) error {
			return test.ValidateArgs(opts)
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			if opts.TestReport != "" {
				// Force JsonOutput
				if opts.Output == unversioned.Text {
					opts.JSON = true
					opts.Output = unversioned.Json

					logrus.Warn("Raw text format unsupported for writing output file, defaulting to JSON")
				}
				testReportFile, err := os.Create(opts.TestReport)
				if err != nil {
					return err
				}
				rootCmd.SetOutput(testReportFile)
				out = testReportFile // override writer
			}

			if opts.Quiet {
				out = ioutil.Discard
			}

			color.NoColor = opts.NoColor

			if opts.JSON {
				opts.Output = unversioned.Json
			}

			return run(out)
		},
	}

	AddTestFlags(testCmd)
	return testCmd
}

func run(out io.Writer) error {
	args = &drivers.DriverConfig{
		Image:    opts.ImagePath,
		Save:     opts.Save,
		Metadata: opts.Metadata,
		Runtime:  opts.Runtime,
	}

	var err error

	if opts.Pull {
		if opts.Driver != drivers.Docker {
			logrus.Fatal("Image pull not supported when not using Docker driver")
		}
		var repository, tag string
		parts := splitImagePath(opts.ImagePath)
		if len(parts) < 2 {
			logrus.Fatal("No tag specified for provided image")
		}
		repository = parts[0]
		tag = parts[1]
		client, err := docker.NewClientFromEnv()
		if err = client.PullImage(docker.PullImageOptions{
			Repository:   repository,
			Tag:          tag,
			OutputStream: out,
		}, docker.AuthConfiguration{}); err != nil {
			logrus.Fatalf("Error pulling remote image %s: %s", opts.ImagePath, err.Error())
		}
	}

	if opts.Driver == drivers.Host && !utils.UserConfirmation(warnMessage, opts.Force) {
		logrus.Fatalf("Aborted by user")
	}

	driverImpl = drivers.InitDriverImpl(opts.Driver)
	if driverImpl == nil {
		logrus.Fatalf("Unsupported driver type: %s", opts.Driver)
	}
	if err != nil {
		logrus.Fatal(err.Error())
	}
	channel := make(chan interface{}, 1)
	go runTests(out, channel, args, driverImpl)
	// TODO(nkubala): put a sync.WaitGroup here
	return test.ProcessResults(out, opts.Output, channel)
}

func runTests(out io.Writer, channel chan interface{}, args *drivers.DriverConfig, driverImpl func(drivers.DriverConfig) (drivers.Driver, error)) {
	for _, file := range opts.ConfigFiles {
		if opts.Output == unversioned.Text {
			output.Banner(out, file)
		}
		tests, err := test.Parse(file, args, driverImpl)
		if err != nil {
			channel <- &unversioned.TestResult{
				Errors: []string{
					fmt.Sprintf("Error parsing config file: %s", err),
				},
			}
			continue // Continue with other config files
		}
		tests.RunAll(channel, file)
	}
	close(channel)
}

func splitImagePath(imagePath string) []string {
	var parts []string
	if strings.Contains(imagePath, "@") {
		parts = strings.Split(imagePath, "@")
	} else {
		switch countColons := strings.Count(imagePath, ":"); countColons {
		case 0:
			parts = []string{imagePath}
		case 1:
			match, _ := regexp.MatchString(`:\d{1,5}\/`, imagePath)
			if match {
				//colon is part of a registry port and no tag available
				parts = []string{imagePath}
			} else {
				//colon separates image name and tag
				parts = strings.Split(imagePath, ":")
			}
		default:
			imageTagRegex, _ := regexp.Compile("(.*):(.*)")
			parts = imageTagRegex.FindStringSubmatch(imagePath)[1:]
		}
	}
	return parts
}

func AddTestFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&opts.ImagePath, "image", "i", "", "path to test image")
	cmd.Flags().StringVarP(&opts.Driver, "driver", "d", "docker", "driver to use when running tests")
	cmd.Flags().StringVar(&opts.Metadata, "metadata", "", "path to image metadata file")
	cmd.Flags().StringVar(&opts.Runtime, "runtime", "", "runtime to use with docker driver")

	cmd.Flags().BoolVar(&opts.Pull, "pull", false, "force a pull of the image before running tests")
	cmd.Flags().BoolVar(&opts.Save, "save", false, "preserve created containers after test run")
	cmd.Flags().BoolVarP(&opts.Quiet, "quiet", "q", false, "flag to suppress output")
	cmd.Flags().BoolVarP(&opts.Force, "force", "f", false, "force run of host driver (without user prompt)")
	cmd.Flags().BoolVarP(&opts.JSON, "json", "j", false, "output test results in json format")
	cmd.Flags().MarkDeprecated("json", "please use --output instead")
	cmd.Flags().VarP(&opts.Output, "output", "o", "output format for the test report (available format: text, json, junit)")
	cmd.Flags().BoolVar(&opts.NoColor, "no-color", false, "no color in the output")

	cmd.Flags().StringArrayVarP(&opts.ConfigFiles, "config", "c", []string{}, "test config files")
	cmd.MarkFlagRequired("config")
	cmd.Flags().StringVar(&opts.TestReport, "test-report", "", "generate test report and write it to specified file (supported format: json, junit; default: json)")
}
