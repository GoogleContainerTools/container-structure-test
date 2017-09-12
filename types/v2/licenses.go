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
	"path"
	"strings"
	"testing"

	"github.com/GoogleCloudPlatform/runtimes-common/structure_tests/drivers"
)

// Not currently used, but leaving the possibility open
type LicenseTest struct {
	Debian bool
	Files  []string
}

var (
	// Whitelist is the list of packages that we want to automatically pass this
	// check even if it would normally fail for one reason or another.
	whitelist = []string{}

	// Blacklist is the set of words that, if contained in a license file, should cause a failure.
	// This will most likely just be names of unsupported licenses.
	blacklist = []string{"AGPL", "WTFPL"}
)

func checkFile(t *testing.T, licenseFile string, driver drivers.Driver) {
	// Read through the copyright file and make sure don't have an unauthorized license
	license, err := driver.ReadFile(t, licenseFile)
	if err != nil {
		t.Errorf("Error reading license file for %s: %s", licenseFile, err.Error())
		return
	}
	contents := strings.ToUpper(string(license))
	for _, b := range blacklist {
		if strings.Contains(contents, b) {
			t.Errorf("Invalid license for %s, license contains %s", licenseFile, b)
			return
		}
	}
}

func checkLicenses(t *testing.T, tt LicenseTest, driver drivers.Driver) {
	if tt.Debian {
		root := "/usr/share/doc"
		packages, err := driver.ReadDir(t, root)
		if err != nil {
			t.Fatalf("%s", err)
		}
		for _, p := range packages {
			if !p.IsDir() {
				continue
			}

			t.Logf(p.Name())

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
			licenseFile := path.Join(root, p.Name(), "copyright")
			_, err := driver.StatFile(t, licenseFile)
			if err != nil {
				t.Errorf("Error reading license file for %s: %s", p.Name(), err.Error())
				continue
			}

			checkFile(t, licenseFile, driver)
		}
	}

	for _, file := range tt.Files {
		checkFile(t, file, driver)
	}
}

func (lt LicenseTest) LogName(num int) string {
	return fmt.Sprintf("License Test #%d", num)
}
