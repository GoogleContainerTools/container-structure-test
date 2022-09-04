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

package v1

import (
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
	// Wait till the file is Processed so we can display the results per file.
	fileProcessed := make(chan bool, 1)
	go st.runAll(channel, fileProcessed)
	<-fileProcessed
}

func (st *StructureTest) runAll(channel chan interface{}, fileProcessed chan bool) {
	st.RunCommandTests(channel)
	st.RunFileContentTests(channel)
	st.RunFileExistenceTests(channel)
	st.RunLicenseTests(channel)
	fileProcessed <- true
}

func (st *StructureTest) RunCommandTests(channel chan interface{}) {
	for _, test := range st.CommandTests {
		if err := test.Validate(); err != nil {
			logrus.Error(err.Error())
			continue
		}
		driver, err := st.NewDriver()
		if err != nil {
			logrus.Fatal(err.Error())
		}
		vars := append(st.GlobalEnvVars, test.EnvVars...)
		if err = driver.Setup(vars, test.Setup); err != nil {
			logrus.Error(err.Error())
			driver.Destroy()
			continue
		}
		defer func() {
			if err := driver.Teardown(test.Teardown); err != nil {
				logrus.Error(err.Error())
			}
			driver.Destroy()
		}()
		channel <- test.Run(driver)
	}
}

func (st *StructureTest) RunFileExistenceTests(channel chan interface{}) {
	for _, test := range st.FileExistenceTests {
		if err := test.Validate(); err != nil {
			logrus.Error(err.Error())
			continue
		}
		driver, err := st.NewDriver()
		if err != nil {
			logrus.Fatal(err.Error())
		}
		channel <- test.Run(driver)
		driver.Destroy()
	}
}

func (st *StructureTest) RunFileContentTests(channel chan interface{}) {
	for _, test := range st.FileContentTests {
		if err := test.Validate(); err != nil {
			logrus.Error(err.Error())
			continue
		}
		driver, err := st.NewDriver()
		if err != nil {
			logrus.Fatal(err.Error())
		}
		channel <- test.Run(driver)
		driver.Destroy()
	}
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
