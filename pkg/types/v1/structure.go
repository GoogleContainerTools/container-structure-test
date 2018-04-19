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
	"github.com/GoogleCloudPlatform/runtimes-common/ctc_lib"

	"github.com/GoogleContainerTools/container-structure-test/pkg/drivers"
	types "github.com/GoogleContainerTools/container-structure-test/pkg/types/unversioned"
)

type StructureTest struct {
	DriverImpl         func(drivers.DriverConfig) (drivers.Driver, error)
	DriverArgs         drivers.DriverConfig
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
			ctc_lib.Log.Error(err.Error())
			continue
		}
		driver, err := st.NewDriver()
		if err != nil {
			ctc_lib.Log.Error(err.Error())
			continue
		}
		defer driver.Destroy()
		vars := append(st.GlobalEnvVars, test.EnvVars...)
		if err = driver.Setup(vars, test.Setup); err != nil {
			ctc_lib.Log.Error(err.Error())
			continue
		}
		defer func() {
			if err := driver.Teardown(vars, test.Teardown); err != nil {
				ctc_lib.Log.Error(err.Error())
			}
		}()
		channel <- test.Run(driver)
	}

}

func (st *StructureTest) RunFileExistenceTests(channel chan interface{}) {
	for _, test := range st.FileExistenceTests {
		if err := test.Validate(); err != nil {
			ctc_lib.Log.Error(err.Error())
			continue
		}
		driver, err := st.NewDriver()
		if err != nil {
			ctc_lib.Log.Fatalf(err.Error())
		}
		defer driver.Destroy()
		channel <- test.Run(driver)
	}

}
func (st *StructureTest) RunFileContentTests(channel chan interface{}) {
	for _, test := range st.FileContentTests {
		if err := test.Validate(); err != nil {
			ctc_lib.Log.Error(err.Error())
			continue
		}
		driver, err := st.NewDriver()
		if err != nil {
			ctc_lib.Log.Error(err)
			ctc_lib.Log.Error(err.Error())
		}
		defer driver.Destroy()
		channel <- test.Run(driver)
	}
}

func (st *StructureTest) RunLicenseTests(channel chan interface{}) {
	for _, test := range st.LicenseTests {
		driver, err := st.NewDriver()
		if err != nil {
			ctc_lib.Log.Error(err.Error())
			continue
		}
		defer driver.Destroy()
		channel <- test.Run(driver)
	}
}
