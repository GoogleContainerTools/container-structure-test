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

package utils

import (
	"fmt"
	"os"
	"regexp"

	"github.com/sirupsen/logrus"
)

var (
	yesResponses = []string{"y", "Y", "yes", "Yes", "YES"}
	noResponses  = []string{"n", "N", "no", "No", "NO"}
)

const (
	NoopCommand string = "NOOP_COMMAND_DO_NOT_RUN"
	DebianRoot  string = "/usr/share/doc"
	LicenseFile string = "copyright"
)

func CompileAndRunRegex(regex string, base string, shouldMatch bool) bool {
	r, rErr := regexp.Compile(regex)
	if rErr != nil {
		logrus.Errorf("Error compiling regex %s : %s", regex, rErr.Error())
		return false
	}
	return shouldMatch == r.MatchString(base)
}

// adapted from https://gist.github.com/albrow/5882501
func UserConfirmation(message string, force bool) bool {
	fmt.Println(message)
	if force {
		fmt.Println("Forcing test run!")
		return true
	}

	var input string
	_, err := fmt.Scanln(&input)
	if err != nil {
		logrus.Errorf("error reading input from stdin: %s", err.Error())
		return false
	}
	for _, response := range yesResponses {
		if input == response {
			return true
		}
	}
	for _, response := range noResponses {
		if input == response {
			return false
		}
	}
	fmt.Println("Please type yes or no to continue or exit")
	return UserConfirmation(message, force)
}

func ValueInList(target string, list []string) bool {
	for _, value := range list {
		if target == value {
			return true
		}
	}
	return false
}

func SubstituteEnvVar(arg string, env map[string]string) string {
	f := func(key string) string {
		return env[key]
	}
	subbed := os.Expand(arg, f)
	return subbed
}

func SubstituteEnvVars(args []string, env map[string]string) []string {
	finalArgs := []string{}
	for _, arg := range args {
		finalArgs = append(finalArgs, SubstituteEnvVar(arg, env))
	}
	return finalArgs
}
