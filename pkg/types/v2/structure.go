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
	"github.com/sirupsen/logrus"

	"github.com/GoogleContainerTools/container-structure-test/pkg/drivers"
	types "github.com/GoogleContainerTools/container-structure-test/pkg/types/unversioned"
)

type StructureTest struct {
	DriverImpl          func(drivers.DriverConfig) (drivers.Driver, error)
	DriverArgs          drivers.DriverConfig
	SchemaVersion       string                    `yaml:"schemaVersion"`
	GlobalEnvVars       []types.EnvVar            `yaml:"globalEnvVars"`
	CommandTests        []CommandTest             `yaml:"commandTests"`
	FileExistenceTests  []FileExistenceTest       `yaml:"fileExistenceTests"`
	FileContentTests    []FileContentTest         `yaml:"fileContentTests"`
	MetadataTest        MetadataTest              `yaml:"metadataTest"`
	LicenseTests        []LicenseTest             `yaml:"licenseTests"`
	ContainerRunOptions types.ContainerRunOptions `yaml:"containerRunOptions"`
}

func (st *StructureTest) NewDriver() (drivers.Driver, error) {
	args := st.DriverArgs
	if st.ContainerRunOptions.IsSet() {
		args.RunOpts = st.ContainerRunOptions
	}
	return st.DriverImpl(args)
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
	st.runReadOnlyTests(channel)
	st.RunLicenseTests(channel)
	fileProcessed <- true
}

// runReadOnlyTests answers every test that only reads the image against ONE driver. Under
// the tar driver a driver is an extracted copy of the image, so building one per test
// unpacks the same image once per test.
func (st *StructureTest) runReadOnlyTests(channel chan interface{}) {
	driver, err := st.readOnlyDriver(channel, "")
	if err != nil {
		return
	}
	defer driver.Destroy()

	st.runFileContentTests(channel, driver)
	st.runFileExistenceTests(channel, driver)
	st.runMetadataTests(channel, driver)
}

// readOnlyDriver builds the driver the read-only tests share and applies the run's global
// environment to it once, reporting either failure under name.
func (st *StructureTest) readOnlyDriver(channel chan interface{}, name string) (drivers.Driver, error) {
	driver, err := st.NewDriver()
	if err != nil {
		channel <- driverError(name, "error creating driver: %s", err.Error())
		return nil, err
	}
	if err = driver.SetEnv(st.GlobalEnvVars); err != nil {
		driver.Destroy()
		channel <- driverError(name, "error setting env vars: %s", err.Error())
		return nil, err
	}
	return driver, nil
}

func driverError(name, format string, args ...interface{}) *types.TestResult {
	res := &types.TestResult{Name: name, Pass: false}
	res.Errorf(format, args...)
	return res
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
	if len(st.FileExistenceTests) == 0 {
		return
	}
	driver, err := st.readOnlyDriver(channel, st.FileExistenceTests[0].Name)
	if err != nil {
		return
	}
	defer driver.Destroy()
	st.runFileExistenceTests(channel, driver)
}

func (st *StructureTest) runFileExistenceTests(channel chan interface{}, driver drivers.Driver) {
	for _, test := range st.FileExistenceTests {
		if !test.Validate(channel) {
			continue
		}
		channel <- test.Run(driver)
	}
}

func (st *StructureTest) RunFileContentTests(channel chan interface{}) {
	if len(st.FileContentTests) == 0 {
		return
	}
	driver, err := st.readOnlyDriver(channel, st.FileContentTests[0].Name)
	if err != nil {
		return
	}
	defer driver.Destroy()
	st.runFileContentTests(channel, driver)
}

func (st *StructureTest) runFileContentTests(channel chan interface{}, driver drivers.Driver) {
	for _, test := range st.FileContentTests {
		if !test.Validate(channel) {
			continue
		}
		channel <- test.Run(driver)
	}
}

func (st *StructureTest) RunMetadataTests(channel chan interface{}) {
	if st.MetadataTest.IsEmpty() {
		logrus.Debug("Skipping empty metadata test")
		return
	}
	driver, err := st.readOnlyDriver(channel, st.MetadataTest.LogName())
	if err != nil {
		return
	}
	defer driver.Destroy()
	st.runMetadataTests(channel, driver)
}

func (st *StructureTest) runMetadataTests(channel chan interface{}, driver drivers.Driver) {
	if st.MetadataTest.IsEmpty() {
		logrus.Debug("Skipping empty metadata test")
		return
	}
	if !st.MetadataTest.Validate(channel) {
		return
	}
	channel <- st.MetadataTest.Run(driver)
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
