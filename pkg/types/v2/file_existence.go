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
	"os"

	"github.com/GoogleCloudPlatform/runtimes-common/ctc_lib"
	"github.com/pkg/errors"

	"github.com/GoogleContainerTools/container-structure-test/pkg/drivers"
	types "github.com/GoogleContainerTools/container-structure-test/pkg/types/unversioned"
	"github.com/GoogleContainerTools/container-structure-test/pkg/utils"
)

type FileExistenceTest struct {
	Name        string `yaml:"name"`        // name of test
	Path        string `yaml:"path"`        // file to check existence of
	ShouldExist bool   `yaml:"shouldExist"` // whether or not the file should exist
	Permissions string `yaml:"permissions"` // expected Unix permission string of the file, e.g. drwxrwxrwx
}

func (fe FileExistenceTest) MarshalYAML() (interface{}, error) {
	return FileExistenceTest{ShouldExist: true}, nil
}

func (fe *FileExistenceTest) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// Create a type Alias and call unmarshal on this type to unmarshal the yaml text into
	// struct. Else, calling unmarshal on FileExistence will result in to infinite recursive loop.
	type feAlias FileExistenceTest
	feTest := feAlias{
		ShouldExist: true,
	}
	err := unmarshal(&feTest)
	if err != nil {
		return err
	}
	*fe = FileExistenceTest(feTest)
	return nil
}

func (ft FileExistenceTest) Validate(channel chan interface{}) bool {
	res := &types.TestResult{}
	if ft.Name == "" {
		res.Errorf("Please provide a valid name for every test")
	}
	res.Name = ft.Name
	if ft.Path == "" {
		res.Errorf("Please provide a valid file path for test %s", ft.Name)
	}
	if len(res.Errors) > 0 {
		channel <- res
		return false
	}
	return true
}

func (ft FileExistenceTest) LogName() string {
	return fmt.Sprintf("File Existence Test: %s", ft.Name)
}

func (ft FileExistenceTest) Run(driver drivers.Driver) *types.TestResult {
	result := &types.TestResult{
		Name:   ft.LogName(),
		Pass:   true,
		Errors: make([]string, 0),
	}
	ctc_lib.Log.Info(ft.LogName())
	var info os.FileInfo
	config, err := driver.GetConfig()
	if err != nil {
		ctc_lib.Log.Errorf("error retrieving image config: %s", err.Error())
	}
	info, err = driver.StatFile(utils.SubstituteEnvVar(ft.Path, config.Env))
	if info == nil && ft.ShouldExist {
		result.Errorf(errors.Wrap(err, "Error examining file in container").Error())
		result.Fail()
		return result
	}
	if ft.ShouldExist && err != nil {
		result.Errorf("File %s should exist but does not, got error: %s", ft.Path, err)
		result.Fail()
	} else if !ft.ShouldExist && err == nil {
		result.Errorf("File %s should not exist but does", ft.Path)
		result.Fail()
	}
	if ft.Permissions != "" {
		perms := info.Mode()
		if perms.String() != ft.Permissions {
			result.Errorf("%s has incorrect permissions. Expected: %s, Actual: %s", ft.Path, ft.Permissions, perms.String())
			result.Fail()
		}
	}
	return result
}
