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
	"bytes"
	"fmt"
	"path/filepath"
	"strings"

	types "github.com/GoogleContainerTools/container-structure-test/pkg/types/unversioned"
)

var bannerLength = 27 // default banner length

func OutputResult(result *types.TestResult, isQuiet bool) string {
	var strBuffer bytes.Buffer
	strBuffer.WriteString(fmt.Sprintf("=== RUN: %s\n", result.Name))
	if result.Pass {
		strBuffer.WriteString(green("--- PASS\n"))
	} else {
		strBuffer.WriteString(red("--- FAIL\n"))
	}
	if result.Stdout != "" && !isQuiet {
		strBuffer.WriteString(blue(fmt.Sprintf("stdout: %s", result.Stdout)))
	}
	if result.Stderr != "" && !isQuiet {
		strBuffer.WriteString(blue(fmt.Sprintf("stderr: %s", result.Stderr)))
	}
	for _, s := range result.Errors {
		strBuffer.WriteString(orange(fmt.Sprintf("Error: %s\n", s)))
	}
	return strBuffer.String()
}

func Banner(filename string) string {
	var strBuffer bytes.Buffer
	fileStr := fmt.Sprintf("====== Test file: %s ======", filepath.Base(filename))
	bannerLength = len(fileStr)
	strBuffer.WriteString(strings.Repeat("=", bannerLength) + "\n")
	strBuffer.WriteString(fileStr + "\n")
	strBuffer.WriteString(strings.Repeat("=", bannerLength) + "\n")
	return purple(strBuffer.String())
}

func FinalResults(result types.SummaryObject) string {
	if bannerLength%2 == 0 {
		bannerLength++
	}
	var strBuffer bytes.Buffer
	strBuffer.WriteString("\n" + strings.Repeat("=", bannerLength) + "\n")
	strBuffer.WriteString(strings.Repeat("=", (bannerLength-9)/2))
	strBuffer.WriteString(" RESULTS ")
	strBuffer.WriteString(strings.Repeat("=", (bannerLength-9)/2))
	strBuffer.WriteString("\n" + strings.Repeat("=", bannerLength) + "\n")
	strBuffer.WriteString(lightGreen(fmt.Sprintf("Passes:      %d\n", result.Pass)))
	strBuffer.WriteString(lightRed(fmt.Sprintf("Failures:    %d\n", result.Fail)))
	strBuffer.WriteString(cyan(fmt.Sprintf("Total tests: %d\n", result.Total)))
	if result.Fail == 0 {
		strBuffer.WriteString(green("\nPASS\n"))
	} else {
		strBuffer.WriteString(red("\nFAIL\n"))
	}
	return strBuffer.String()
}
