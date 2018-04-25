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

package output

import (
	"fmt"
	"path/filepath"
	"strings"

	types "github.com/GoogleContainerTools/container-structure-test/pkg/types/unversioned"
)

func OutputResult(result *types.TestResult, isQuiet bool) string {
	resultStr := fmt.Sprintf("=== RUN: %s\n", result.Name)
	if result.Pass {
		resultStr += green("--- PASS\n")
	} else {
		resultStr += red("--- FAIL\n")
	}
	if result.Stdout != "" && !isQuiet {
		resultStr += blue(fmt.Sprintf("stdout: %s", result.Stdout))
	}
	if result.Stderr != "" && !isQuiet {
		resultStr += blue(fmt.Sprintf("stderr: %s", result.Stderr))
	}
	for _, s := range result.Errors {
		resultStr += orange(fmt.Sprintf("Error: %s\n", s))
	}
	return resultStr
}

func Banner(filename string) string {
	fileStr := fmt.Sprintf("====== Test file: %s ======", filepath.Base(filename))
	bannerStr := strings.Repeat("=", len(fileStr))
	return purple(bannerStr) + "\n" + purple(fileStr) + "\n" + purple(bannerStr) + "\n"
}

func FinalResults(result types.SummaryObject) string {
	resultStr := "\n===============\n"
	resultStr += "=== RESULTS ===\n"
	resultStr += "===============\n"
	resultStr += lightGreen(fmt.Sprintf("Passes:      %d\n", result.Pass))
	resultStr += lightRed(fmt.Sprintf("Failures:    %d\n", result.Fail))
	resultStr += cyan(fmt.Sprintf("Total tests: %d\n", result.Total))
	if result.Fail == 0 {
		resultStr += green("PASS\n")
	}
	return resultStr
}
