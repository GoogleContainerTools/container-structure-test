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

	"github.com/GoogleCloudPlatform/container-structure-test/drivers"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/ghodss/yaml"
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

	var unmarshal Unmarshaller
	var versionHolder SchemaVersion

	switch {
	case strings.HasSuffix(fp, ".json"):
		unmarshal = json.Unmarshal
	case strings.HasSuffix(fp, ".yaml"):
		unmarshal = yaml.Unmarshal
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

	unmarshal(testContents, st)

	tests, ok := st.(StructureTest) //type assertion
	if !ok {
		return nil, errors.New("Error encountered when type casting Structure Test interface")
	}
	tests.SetDriverImpl(driverImpl, args)
	return tests, nil
}

var configFiles arrayFlags

var imagePath, driver string
var save, pull bool
var driverImpl func([]interface{}) (drivers.Driver, error)
var args []interface{}

func TestMain(m *testing.M) {
	flag.StringVar(&imagePath, "image", "", "path to test image")
	flag.StringVar(&driver, "driver", "docker", "driver to use when running tests")
	flag.BoolVar(&pull, "pull", false, "force a pull of the image before running tests")
	flag.BoolVar(&save, "save", false, "preserve created containers after test run")

	flag.Parse()
	configFiles = flag.Args()

	if imagePath == "" {
		fmt.Println("Please supply path to image or tarball to test against")
		os.Exit(1)
	}
	// These args MUST be passed in this order; the docker/tar drivers expect them there,
	// and will not be instantiated correctly if they are changed
	args = make([]interface{}, 2)
	args[0] = imagePath
	args[1] = save

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
