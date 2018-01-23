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

package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"testing"

	"gopkg.in/yaml.v2"

	"github.com/GoogleCloudPlatform/container-structure-test/drivers"
	"github.com/GoogleCloudPlatform/container-structure-test/utils"
	docker "github.com/fsouza/go-dockerclient"
)

var totalTests int

func TestAll(t *testing.T) {
	for _, file := range configFiles {
		tests, err := Parse(t, file)
		if err != nil {
			log.Fatalf("Error parsing config file: %s", err)
		}
		log.Printf("Running tests for file %s", file)
		totalTests += tests.RunAll(t)
	}
	if totalTests == 0 {
		t.Fatalf("No tests run! Check config file format.")
	} else {
		t.Logf("Total tests run: %d", totalTests)
	}
}

func Parse(t *testing.T, fp string) (StructureTest, error) {
	testContents, err := ioutil.ReadFile(fp)
	if err != nil {
		return nil, err
	}

	// We first have to unmarshal to determine the schema version, then we unmarshal again
	// to do the full parse.
	var unmarshal Unmarshaller
	var strictUnmarshal Unmarshaller
	var versionHolder SchemaVersion

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

	var st StructureTest
	if schemaVersion, ok := schemaVersions[version]; ok {
		st = schemaVersion()
	} else {
		return nil, errors.New("Unsupported schema version: " + version)
	}

	strictUnmarshal(testContents, st)

	tests, ok := st.(StructureTest) //type assertion
	if !ok {
		return nil, errors.New("Error encountered when type casting Structure Test interface")
	}
	tests.SetDriverImpl(driverImpl, *args)
	return tests, nil
}

var configFiles arrayFlags

var imagePath, driver, metadata string
var save, pull, force bool
var driverImpl func(drivers.DriverConfig) (drivers.Driver, error)
var args *drivers.DriverConfig

func TestMain(m *testing.M) {
	flag.StringVar(&imagePath, "image", "", "path to test image")
	flag.StringVar(&driver, "driver", "docker", "driver to use when running tests")
	flag.StringVar(&metadata, "metadata", "", "path to image metadata file")
	flag.BoolVar(&pull, "pull", false, "force a pull of the image before running tests")
	flag.BoolVar(&save, "save", false, "preserve created containers after test run")
	flag.BoolVar(&force, "force", false, "force run of host driver (without command line input)")

	flag.Parse()
	configFiles = flag.Args()

	if driver == drivers.Host {
		if metadata == "" {
			fmt.Println("Please provide path to image metadata file")
			os.Exit(1)
		}
		if imagePath != "" {
			fmt.Println("Cannot provide both image path and metadata file")
			os.Exit(1)
		}
	} else {
		if imagePath == "" {
			fmt.Println("Please supply path to image or tarball to test against")
			os.Exit(1)
		}
		if metadata != "" {
			fmt.Println("Cannot provide both image path and metadata file")
			os.Exit(1)
		}
	}
	args = &drivers.DriverConfig{
		Image:    imagePath,
		Save:     save,
		Metadata: metadata,
	}

	if len(configFiles) == 0 {
		fmt.Println("Please provide at least one test config file")
		os.Exit(1)
	}

	var err error

	if pull {
		if driver != drivers.Docker {
			fmt.Println("Image pull not supported when not using Docker driver")
			os.Exit(1)
		}
		var repository, tag string
		parts := strings.Split(imagePath, ":")
		repository = parts[0]
		if len(parts) < 2 {
			fmt.Println("Please provide specific tag for image")
			os.Exit(1)
		}
		tag = parts[1]
		client, err := docker.NewClientFromEnv()
		if err = client.PullImage(docker.PullImageOptions{
			Repository:   repository,
			Tag:          tag,
			OutputStream: os.Stdout,
		}, docker.AuthConfiguration{}); err != nil {
			fmt.Printf("Error pulling remote image %s: %s", imagePath, err.Error())
			os.Exit(1)
		}
	}

	warnMessage := `WARNING: Using the host driver runs tests directly on the machine 
that this binary is being run on, and can potentially corrupt your system.
Be sure you know what you're doing before continuing!

Continue? (y/n)`

	if driver == drivers.Host && !utils.UserConfirmation(warnMessage, force) {
		os.Exit(1)
	}

	driverImpl = drivers.InitDriverImpl(driver)
	if driverImpl == nil {
		fmt.Printf("Unsupported driver type: %s", driver)
		os.Exit(1)
	}
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	fmt.Printf("Using driver %s\n", driver)

	if exit := m.Run(); exit != 0 {
		os.Exit(exit)
	}
}
