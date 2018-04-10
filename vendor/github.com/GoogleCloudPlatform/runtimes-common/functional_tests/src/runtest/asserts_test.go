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

import "testing"

func TestStringAssertExactly(t *testing.T) {
	rule := StringAssert{Exactly: "to be exact"}
	assertShouldPass(t, "to be exact", rule)
	assertShouldFail(t, "\nto be exact", rule)
	assertShouldFail(t, "to be exact\n", rule)
	assertShouldFail(t, " to be exact", rule)
	assertShouldFail(t, "to be exact ", rule)
}

func TestStringAssertEquals(t *testing.T) {
	rule := StringAssert{Equals: "to be equal"}
	assertShouldPass(t, "to be equal", rule)
	assertShouldPass(t, "\nto be equal", rule)
	assertShouldPass(t, "to be equal\n", rule)
	assertShouldPass(t, " to be equal", rule)
	assertShouldPass(t, "to be equal ", rule)
	assertShouldFail(t, "to be different", rule)
	assertShouldFail(t, "to be equal\nbut", rule)
}

func TestStringAssertMatches(t *testing.T) {
	rule := StringAssert{Matches: "a[bc]d"}
	assertShouldPass(t, "abd", rule)
	assertShouldPass(t, "123acd456", rule)
	assertShouldFail(t, "ad", rule)
}

func TestStringAssertMatchesBadRegex(t *testing.T) {
	rule := StringAssert{Matches: `\`}
	assertShouldFail(t, `\`, rule)
}

func TestStringAssertMustBeEmpty(t *testing.T) {
	rule := StringAssert{MustBeEmpty: true}
	assertShouldPass(t, "", rule)
	assertShouldFail(t, "a", rule)
}

func assertShouldPass(t *testing.T, value string, rule StringAssert) {
	outcome := DoStringAssert(value, rule)
	if len(outcome) > 0 {
		t.Errorf("Expected to pass for string: %s\n...but failed with error: %s", value, outcome)
	}
}

func assertShouldFail(t *testing.T, value string, rule StringAssert) {
	outcome := DoStringAssert(value, rule)
	if len(outcome) == 0 {
		t.Errorf("Expected to fail but was passing for string: %s", value)
	} else {
		t.Logf("DoStringAssert() output: %s", outcome)
	}
}
