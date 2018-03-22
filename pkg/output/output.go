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

	types "github.com/GoogleCloudPlatform/container-structure-test/pkg/types/unversioned"
)

type OutWriter struct {
	Format  string // TODO(nkubala): implement JSON type
	Verbose bool
	Quiet   bool
}

func (o *OutWriter) OutputResult(result *types.TestResult) (int, int) {
	// TODO(nkubala): use template
	pass := 0
	fail := 0
	o.Printf("=== RUN: %s", result.Name)
	if result.Pass {
		pass++
		o.green("--- PASS")
	} else {
		fail++
		o.red("--- FAIL")
	}
	if o.Verbose {
		if result.Stdout != "" {
			o.blue(fmt.Sprintf("stdout: %s", result.Stdout))
		}
		if result.Stderr != "" {
			o.blue(fmt.Sprintf("stderr: %s", result.Stderr))
		}
	}
	for _, s := range result.Errors {
		o.orange(fmt.Sprintf("Error: %s\n", s))
	}
	return pass, fail
}

func (o *OutWriter) Banner(filename string) {
	fileStr := fmt.Sprintf("====== Test file: %s ======", filepath.Base(filename))
	bannerStr := strings.Repeat("=", len(fileStr))
	o.purple(bannerStr)
	o.purple(fileStr)
	o.purple(bannerStr)
}

func (o *OutWriter) FinalResults(results []*types.TestResult) bool {
	totalPass := 0
	totalFail := 0
	for _, result := range results {
		if result.Pass {
			totalPass++
		} else {
			totalFail++
		}
	}
	totalTests := totalPass + totalFail
	if totalTests == 0 {
		o.red("No tests run! Check config file format.")
		return false
	}
	o.Print("===============")
	o.Print("=== RESULTS ===")
	o.Print("===============")
	o.lightGreen(fmt.Sprintf("Passes:      %d", totalPass))
	o.lightRed(fmt.Sprintf("Failures:    %d", totalFail))
	o.cyan(fmt.Sprintf("Total tests: %d", totalTests))
	if totalFail > 0 {
		o.red("\nFAIL")
		return false
	}
	o.green("\nPASS")
	return true
}
