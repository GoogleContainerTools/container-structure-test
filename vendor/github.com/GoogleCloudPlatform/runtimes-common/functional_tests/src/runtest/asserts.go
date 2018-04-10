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

package main

import (
	"fmt"
	"regexp"
	"strings"
)

// DoStringAssert asserts the string value against the rule, returning
// an empty string if the assertion succeeds, or the error message otherwise.
func DoStringAssert(value string, rule StringAssert) string {
	if len(rule.Exactly) > 0 {
		if value != rule.Exactly {
			return fmt.Sprintf("Should have matched exactly:\n%s\n... but was:\n%s", rule.Exactly, value)
		}
	}
	if len(rule.Equals) > 0 {
		trimmed := strings.TrimSpace(value)
		if trimmed != rule.Equals {
			return fmt.Sprintf("Should have been:\n%s\n... but was:\n%s", rule.Equals, trimmed)
		}
	}
	if len(rule.Matches) > 0 {
		r, err := regexp.Compile(rule.Matches)
		if err != nil {
			return fmt.Sprintf("Regex failed to compile: %s", rule.Matches)
		}
		if !r.MatchString(value) {
			return fmt.Sprintf("Should have matched regex:\n%s\n... but was:\n%s", rule.Matches, value)
		}
	}
	if rule.MustBeEmpty {
		if len(value) > 0 {
			return fmt.Sprintf("Should have been empty, but was:\n%s", value)
		}
	}
	return ""
}
