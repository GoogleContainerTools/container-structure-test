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

	"github.com/GoogleCloudPlatform/container-structure-test/pkg/drivers"
	types "github.com/GoogleCloudPlatform/container-structure-test/pkg/types/unversioned"
)

type StructureTest struct {
	DriverImpl         func(drivers.DriverConfig) (drivers.Driver, error)
	DriverArgs         drivers.DriverConfig
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

func (st *StructureTest) RunAll() []*types.TestResult {
	results := make([]*types.TestResult, 0)
	results = append(results, st.RunCommandTests()...)
	results = append(results, st.RunFileExistenceTests()...)
	results = append(results, st.RunFileContentTests()...)
	results = append(results, st.RunLicenseTests()...)
	return results
}

func (st *StructureTest) RunCommandTests() []*types.TestResult {
	results := make([]*types.TestResult, 0)
	for _, test := range st.CommandTests {
		if err := test.Validate(); err != nil {
			logrus.Error(err.Error())
			continue
		}
		driver, err := st.NewDriver()
		if err != nil {
			logrus.Error(err.Error())
			continue
		}
		defer driver.Destroy()
		vars := append(st.GlobalEnvVars, test.EnvVars...)
		if err = driver.Setup(vars, test.Setup); err != nil {
			logrus.Error(err.Error())
			continue
		}
		defer func() {
			if err := driver.Teardown(vars, test.Teardown); err != nil {
				logrus.Error(err.Error())
			}
		}()
		results = append(results, test.Run(driver))
	}
	return results
}

func (st *StructureTest) RunFileExistenceTests() []*types.TestResult {
	results := make([]*types.TestResult, 0)
	for _, test := range st.FileExistenceTests {
		if err := test.Validate(); err != nil {
			logrus.Error(err.Error())
			continue
		}
		driver, err := st.NewDriver()
		if err != nil {
			logrus.Fatalf(err.Error())
		}
		defer driver.Destroy()
		results = append(results, test.Run(driver))
	}
	return results
}

func (st *StructureTest) RunFileContentTests() []*types.TestResult {
	results := make([]*types.TestResult, 0)
	for _, test := range st.FileContentTests {
		if err := test.Validate(); err != nil {
			logrus.Error(err.Error())
			continue
		}
		driver, err := st.NewDriver()
		if err != nil {
			logrus.Error(err)
			logrus.Error(err.Error())
		}
		defer driver.Destroy()
		results = append(results, test.Run(driver))
	}
	return results
}

func (st *StructureTest) RunMetadataTests() *types.TestResult {
	driver, err := st.NewDriver()
	if err != nil {
		logrus.Error(err.Error())
	}
	defer driver.Destroy()
	return st.MetadataTest.Run(driver)
}

func (st *StructureTest) RunLicenseTests() []*types.TestResult {
	results := make([]*types.TestResult, 0)
	for _, test := range st.LicenseTests {
		driver, err := st.NewDriver()
		if err != nil {
			logrus.Error(err.Error())
			continue
		}
		defer driver.Destroy()
		results = append(results, test.Run(driver))
	}
	return results
}
