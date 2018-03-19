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
	"github.com/sirupsen/logrus"
	"path"
	"strings"

	"github.com/GoogleCloudPlatform/container-structure-test/pkg/drivers"
	types "github.com/GoogleCloudPlatform/container-structure-test/pkg/types/unversioned"
	"github.com/GoogleCloudPlatform/container-structure-test/pkg/utils"
)

// Not currently used, but leaving the possibility open
type LicenseTest struct {
	Debian bool     `yaml:"debian"`
	Files  []string `yaml:"files"`
}

var (
	// Whitelist is the list of packages that we want to automatically pass this
	// check even if it would normally fail for one reason or another.
	whitelist = []string{}

	// Blacklist is the set of words that, if contained in a license file, should cause a failure.
	// This will most likely just be names of unsupported licenses.
	blacklist = []string{"AGPL", "WTFPL"}
)

func checkFile(licenseFile string, driver drivers.Driver) error {
	// Read through the copyright file and make sure don't have an unauthorized license
	license, err := driver.ReadFile(licenseFile)
	if err != nil {
		return fmt.Errorf("Error reading license file for %s: %s", licenseFile, err.Error())
	}
	contents := strings.ToUpper(string(license))
	for _, b := range blacklist {
		if strings.Contains(contents, b) {
			return fmt.Errorf("Invalid license for %s, license contains %s", licenseFile, b)
		}
	}
	return nil
}

func (lt LicenseTest) Run(driver drivers.Driver) *types.TestResult {
	result := &types.TestResult{
		Name:   lt.LogName(),
		Pass:   true,
		Errors: make([]string, 0),
	}
	logrus.Info(lt.LogName())
	if lt.Debian {
		root := utils.DebianRoot
		packages, err := driver.ReadDir(root)
		if err != nil {
			result.Errorf("Error reading directory: %s", err)
			result.Fail()
			return result
		}
		for _, p := range packages {
			if !p.IsDir() {
				continue
			}
			logrus.Infof(p.Name())
			// Skip over packages in the whitelist
			whitelisted := false
			for _, w := range whitelist {
				if w == p.Name() {
					whitelisted = true
					break
				}
			}
			if whitelisted {
				continue
			}

			// If package doesn't have copyright file, log an error.
			licenseFile := path.Join(root, p.Name(), utils.LicenseFile)
			_, err := driver.StatFile(licenseFile)
			if err != nil {
				result.Errorf("Error reading license file for %s: %s", p.Name(), err.Error())
				result.Fail()
			}

			if err = checkFile(licenseFile, driver); err != nil {
				result.Error(err.Error())
				result.Fail()
			}
		}
	}

	for _, file := range lt.Files {
		if err := checkFile(file, driver); err != nil {
			result.Error(err.Error())
			result.Fail()
		}
	}
	return result
}

func (lt LicenseTest) LogName() string {
	return "License Test"
}
