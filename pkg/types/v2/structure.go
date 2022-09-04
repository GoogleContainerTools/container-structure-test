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

package v2

import (
	"fmt"

	"github.com/GoogleContainerTools/container-structure-test/pkg/drivers"
	types "github.com/GoogleContainerTools/container-structure-test/pkg/types/unversioned"
	"github.com/sirupsen/logrus"
)

type StructureTest struct {
	DriverImpl         func(drivers.DriverConfig) (drivers.Driver, error)
	DriverArgs         drivers.DriverConfig
	SchemaVersion      string              `yaml:"schemaVersion"`
	GlobalEnvVars      []types.EnvVar      `yaml:"globalEnvVars"`
	CommandTests       []CommandTest       `yaml:"commandTests"`
	FileExistenceTests []FileExistenceTest `yaml:"fileExistenceTests"`
	FileContentTests   []FileContentTest   `yaml:"fileContentTests"`
	MetadataTest       MetadataTest        `yaml:"metadataTest"`
	LicenseTests       []LicenseTest       `yaml:"licenseTests"`
}

func (st *StructureTest) NewDriver() (drivers.Driver, error) {
	return st.DriverImpl(st.DriverArgs)
}

func (st *StructureTest) SetDriverImpl(f func(drivers.DriverConfig) (drivers.Driver, error), args drivers.DriverConfig) {
	st.DriverImpl = f
	st.DriverArgs = args
}

func (st *StructureTest) RunAll(channel chan interface{}, file string) {
	fileProcessed := make(chan bool, 1)
	go st.runAll(channel, fileProcessed)
	<-fileProcessed
}

func (st *StructureTest) runAll(channel chan interface{}, fileProcessed chan bool) {
	st.RunCommandTests(channel)
	st.RunFileContentTests(channel)
	st.RunFileExistenceTests(channel)
	st.RunLicenseTests(channel)
	st.RunMetadataTests(channel)
	fileProcessed <- true
}

func (st *StructureTest) RunCommandTests(channel chan interface{}) {
	for _, test := range st.CommandTests {
		if !test.Validate(channel) {
			continue
		}
		res := &types.TestResult{
			Name: test.Name,
			Pass: false,
		}
		driver, err := st.NewDriver()
		if err != nil {
			res.Errorf("error creating driver: %s", err.Error())
			channel <- res
			continue
		}
		defer driver.Destroy()
		if err = driver.SetEnv(st.GlobalEnvVars); err != nil {
			res.Errorf("error setting env vars: %s", err.Error())
			channel <- res
			continue
		}
		if err = driver.Setup(test.EnvVars, test.Setup); err != nil {
			res.Errorf("error in setup: %s", err.Error())
			channel <- res
			continue
		}
		defer func() {
			if err := driver.Teardown(test.Teardown); err != nil {
				logrus.Error(err.Error())
			}
		}()
		channel <- test.Run(driver)
	}
}

func (st *StructureTest) RunFileExistenceTests(channel chan interface{}) {
	for _, test := range st.FileExistenceTests {
		if !test.Validate(channel) {
			continue
		}
		res := &types.TestResult{
			Name: test.Name,
			Pass: false,
		}
		driver, err := st.NewDriver()
		if err != nil {
			res.Errorf("error creating driver: %s", err.Error())
			channel <- res
			continue
		}
		if err = driver.SetEnv(st.GlobalEnvVars); err != nil {
			res.Errorf("error setting env vars: %s", err.Error())
			channel <- res
			continue
		}
		channel <- test.Run(driver)
		driver.Destroy()
	}
}

func (st *StructureTest) RunFileContentTests(channel chan interface{}) {
	for _, test := range st.FileContentTests {
		if !test.Validate(channel) {
			continue
		}
		res := &types.TestResult{
			Name: test.Name,
			Pass: false,
		}
		driver, err := st.NewDriver()
		if err != nil {
			res.Errorf("error creating driver: %s", err.Error())
			channel <- res
			continue
		}
		if err = driver.SetEnv(st.GlobalEnvVars); err != nil {
			res.Errorf("error setting env vars: %s", err.Error())
			channel <- res
			continue
		}
		channel <- test.Run(driver)
		driver.Destroy()
	}
}

func (st *StructureTest) RunMetadataTests(channel chan interface{}) {
	if st.MetadataTest.IsEmpty() {
		logrus.Debug("Skipping empty metadata test")
		return
	}
	if !st.MetadataTest.Validate(channel) {
		return
	}
	driver, err := st.NewDriver()
	if err != nil {
		channel <- &types.TestResult{
			Name: st.MetadataTest.LogName(),
			Errors: []string{
				fmt.Sprintf("error creating driver: %s", err.Error()),
			},
		}
		return
	}
	channel <- st.MetadataTest.Run(driver)
	driver.Destroy()
}

func (st *StructureTest) RunLicenseTests(channel chan interface{}) {
	for _, test := range st.LicenseTests {
		driver, err := st.NewDriver()
		if err != nil {
			logrus.Fatal(err.Error())
		}
		channel <- test.Run(driver)
		driver.Destroy()
	}
}
