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
	"archive/tar"
	"fmt"
	"os"

	"github.com/GoogleContainerTools/container-structure-test/pkg/drivers"
	types "github.com/GoogleContainerTools/container-structure-test/pkg/types/unversioned"
	"github.com/GoogleContainerTools/container-structure-test/pkg/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var defaultOwnership = -1

type FileExistenceTest struct {
	Name           string `yaml:"name"`           // name of test
	Path           string `yaml:"path"`           // file to check existence of
	ShouldExist    bool   `yaml:"shouldExist"`    // whether or not the file should exist
	Permissions    string `yaml:"permissions"`    // expected Unix permission string of the file, e.g. drwxrwxrwx
	Uid            int    `yaml:"uid"`            // ID of the owner of the file
	Gid            int    `yaml:"gid"`            // ID of the group of the file
	IsExecutableBy string `yaml:"isExecutableBy"` // name of group that file should be executable by
}

func (fe FileExistenceTest) MarshalYAML() (interface{}, error) {
	return FileExistenceTest{ShouldExist: true}, nil
}

func (fe *FileExistenceTest) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// Create a type alias and call unmarshal on this type to unmarshal the yaml text into
	// struct, since calling unmarshal on FileExistenceTest will result in an infinite loop.
	type FileExistenceTestHolder FileExistenceTest
	holder := FileExistenceTestHolder{
		ShouldExist: true,
		Uid:         defaultOwnership,
		Gid:         defaultOwnership,
	}
	err := unmarshal(&holder)
	if err != nil {
		return err
	}
	*fe = FileExistenceTest(holder)
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
	logrus.Info(ft.LogName())
	var info os.FileInfo
	config, err := driver.GetConfig()
	if err != nil {
		logrus.Errorf("error retrieving image config: %s", err.Error())
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

	// Next assertions don't make sense if the file doesn't exist.
	if !ft.ShouldExist {
		return result
	}

	if ft.Permissions != "" && info != nil {
		perms := info.Mode()
		if perms.String() != ft.Permissions {
			result.Errorf("%s has incorrect permissions. Expected: %s, Actual: %s", ft.Path, ft.Permissions, perms.String())
			result.Fail()
		}
	}
	if ft.IsExecutableBy != "" {
		perms := info.Mode()
		switch ft.IsExecutableBy {
		case "any":
			if perms&0o111 == 0 {
				result.Errorf("%s has incorrect executable bit. Expected to be executable by any, Actual: %s", ft.Path, perms.String())
				result.Fail()
			}
		case "owner":
			if perms&0o100 == 0 {
				result.Errorf("%s has incorrect executable bit. Expected to be executable by owner, Actual: %s", ft.Path, perms.String())
				result.Fail()
			}
		case "group":
			if perms&0o010 == 0 {
				result.Errorf("%s has incorrect executable bit. Expected to be executable by group, Actual: %s", ft.Path, perms.String())
				result.Fail()
			}
		case "other":
			if perms&0o001 == 0 {
				result.Errorf("%s has incorrect executable bit. Expected to be executable by other, Actual: %s", ft.Path, perms.String())
				result.Fail()
			}
		default:
			result.Errorf("%s not recognized as a valid option", ft.IsExecutableBy)
			result.Fail()
		}
	}
	if ft.Uid != defaultOwnership || ft.Gid != defaultOwnership {
		header, ok := info.Sys().(*tar.Header)
		if ok {
			if ft.Uid != defaultOwnership && header.Uid != ft.Uid {
				result.Errorf("%s has incorrect user ownership. Expected: %d, Actual: %d", ft.Path, ft.Uid, header.Uid)
				result.Fail()
			}
			if ft.Gid != defaultOwnership && header.Gid != ft.Gid {
				result.Errorf("%s has incorrect group ownership. Expected: %d, Actual: %d", ft.Path, ft.Gid, header.Gid)
				result.Fail()
			}
		} else {
			result.Errorf("Error checking ownership of file %s", ft.Path)
			result.Fail()
		}
	}
	return result
}
