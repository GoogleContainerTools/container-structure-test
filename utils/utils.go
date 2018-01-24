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
	"regexp"
	"testing"
)

var yesResponses = []string{"y", "Y", "yes", "Yes", "YES"}
var noResponses = []string{"n", "N", "no", "No", "NO"}

func CompileAndRunRegex(regex string, base string, t *testing.T, err string, shouldMatch bool) {
	r, rErr := regexp.Compile(regex)
	if rErr != nil {
		t.Errorf("Error compiling regex %s : %s", regex, rErr.Error())
		return
	}
	if shouldMatch != r.MatchString(base) {
		t.Errorf(err)
	}
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
		// should maybe log something here
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
