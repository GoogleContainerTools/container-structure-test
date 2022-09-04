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

	"github.com/GoogleContainerTools/container-structure-test/cmd/container-structure-test/app/cmd/test"
	"github.com/GoogleContainerTools/container-structure-test/pkg/color"
	"github.com/GoogleContainerTools/container-structure-test/pkg/config"
	"github.com/GoogleContainerTools/container-structure-test/pkg/drivers"
	"github.com/GoogleContainerTools/container-structure-test/pkg/output"
	"github.com/GoogleContainerTools/container-structure-test/pkg/types/unversioned"
	"github.com/GoogleContainerTools/container-structure-test/pkg/utils"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/daemon"
	"github.com/google/go-containerregistry/pkg/v1/layout"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const warnMessage = `WARNING: Using the host driver runs tests directly on the machine
that this binary is being run on, and can potentially corrupt your system.
Be sure you know what you're doing before continuing!

Continue? (y/n)`

var (
	opts = &config.StructureTestOptions{}

	args       *drivers.DriverConfig
	driverImpl func(drivers.DriverConfig) (drivers.Driver, error)
)

func NewCmdTest(out io.Writer) *cobra.Command {
	testCmd := &cobra.Command{
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

					logrus.Warn("raw text format unsupported for writing output file, defaulting to JSON")
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

	if opts.ImageFromLayout != "" {
		if opts.Driver != drivers.Docker {
			logrus.Fatal("--image-from-oci-layout is not supported when not using Docker driver")
		}
		l, err := layout.ImageIndexFromPath(opts.ImageFromLayout)
		if err != nil {
			logrus.Fatalf("loading %s as OCI layout: %v", opts.ImageFromLayout, err)
		}
		m, err := l.IndexManifest()
		if err != nil {
			logrus.Fatalf("could not read OCI index manifest %s: %v", opts.ImageFromLayout, err)
		}

		if len(m.Manifests) != 1 {
			logrus.Fatalf("OCI layout contains %d entries. expected only one", len(m.Manifests))
		}

		desc := m.Manifests[0]

		if desc.MediaType.IsIndex() {
			logrus.Fatal("multi-arch images are not supported yet.")
		}

		img, err := l.Image(desc.Digest)
		if err != nil {
			logrus.Fatalf("could not get image from %s: %v", opts.ImageFromLayout, err)
		}

		var tag name.Tag

		ref := desc.Annotations[v1.AnnotationRefName]
		if ref != "" {
			tag, err = name.NewTag(ref)
			if err != nil {
				logrus.Fatalf("could not parse ref annotation %s: %v", v1.AnnotationRefName, err)
			}
		} else {
			if opts.DefaultImageTag == "" {
				logrus.Fatalf("index does not contain a reference annotation. --default-image-tag must be provided.")
			}
			tag, err = name.NewTag(opts.DefaultImageTag, name.StrictValidation)
			if err != nil {
				logrus.Fatalf("could parse the default image tag %s: %v", opts.DefaultImageTag, err)
			}
		}
		if _, err = daemon.Write(tag, img); err != nil {
			logrus.Fatalf("error loading oci layout into daemon: %v", err)
		}

		opts.ImagePath = tag.String()
		args.Image = tag.String()
	}

	if opts.Pull {
		if opts.Driver != drivers.Docker {
			logrus.Fatal("image pull not supported when not using Docker driver")
		}
		ref, err := name.ParseReference(opts.ImagePath)
		if err != nil {
			logrus.Fatal(err)
		}
		client, err := docker.NewClientFromEnv()
		if err != nil {
			logrus.Fatalf("error connecting to daemon: %v", err)
		}
		if err = client.PullImage(docker.PullImageOptions{
			Repository:   ref.Context().RepositoryStr(),
			Tag:          ref.Identifier(),
			Registry:     ref.Context().RegistryStr(),
			OutputStream: out,
		}, docker.AuthConfiguration{}); err != nil {
			logrus.Fatalf("error pulling remote image %s: %s", opts.ImagePath, err.Error())
		}
	}

	if opts.Driver == drivers.Host && !utils.UserConfirmation(warnMessage, opts.Force) {
		logrus.Fatalf("aborted by user")
	}

	driverImpl = drivers.InitDriverImpl(opts.Driver)
	if driverImpl == nil {
		logrus.Fatalf("unsupported driver type: %s", opts.Driver)
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
					fmt.Sprintf("error parsing config file: %s", err),
				},
			}
			continue // Continue with other config files
		}
		tests.RunAll(channel, file)
	}
	close(channel)
}

func AddTestFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&opts.ImagePath, "image", "i", "", "path to test image")
	cmd.Flags().StringVar(&opts.ImageFromLayout, "image-from-oci-layout", "", "path to the oci layout to test against")
	cmd.Flags().StringVar(&opts.DefaultImageTag, "default-image-tag", "", "default image tag to used when loading images to the daemon. required when --image-from-oci-layout refers to a oci layout lacking the reference annotation.")
	cmd.MarkFlagsMutuallyExclusive("image", "image-from-oci-layout")
	cmd.Flags().StringVarP(&opts.Driver, "driver", "d", "docker", "driver to use when running tests")
	cmd.Flags().StringVar(&opts.Metadata, "metadata", "", "path to image metadata file")
	cmd.Flags().StringVar(&opts.Runtime, "runtime", "", "runtime to use with docker driver")

	cmd.Flags().BoolVar(&opts.Pull, "pull", false, "force a pull of the image before running tests")
	cmd.MarkFlagsMutuallyExclusive("image-from-oci-layout", "pull")
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
