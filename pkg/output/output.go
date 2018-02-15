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

const (
	RED         = "\033[0;31m"
	GREEN       = "\033[0;32m"
	LIGHT_GREEN = "\033[1;32m"
	YELLOW      = "\033[1;33m"
	ORANGE      = "\033[0;33m"
	CYAN        = "\033[0;36m"
	BLUE        = "\033[0;34m"
	PURPLE      = "\033[0;35m"
	NC          = "\033[0m" // No Color
)

// ANSI Color Escape Codes
// Black        0;30     Dark Gray     1;30
// Red          0;31     Light Red     1;31
// Green        0;32     Light Green   1;32
// Brown/Orange 0;33     Yellow        1;33
// Blue         0;34     Light Blue    1;34
// Purple       0;35     Light Purple  1;35
// Cyan         0;36     Light Cyan    1;36

type OutWriter struct {
	Format  string // TODO(nkubala): implement JSON type
	Verbose bool
	Quiet   bool
}

func (o *OutWriter) Banner(filename string) {
	fileStr := fmt.Sprintf("====== Test file: %s ======", filepath.Base(filename))
	bannerStr := strings.Repeat("=", len(fileStr))
	o.Purple(bannerStr)
	o.Purple(fileStr)
	o.Purple(bannerStr)
}

func (o *OutWriter) OutputResults(fullResults []*types.FullResult) int {
	totalPass := 0
	totalFail := 0
	for _, fullResult := range fullResults {
		o.Banner(fullResult.FileName)
		pass := 0
		fail := 0
		if len(fullResult.Results) == 0 {
			o.Red("No tests run! Check config file format.")
			continue
		}
		o.Cyan(fmt.Sprintf("Total tests run: %d\n", len(fullResult.Results)))
		for _, result := range fullResult.Results {
			// TODO(nkubala): use template
			o.Printf("=== RUN: %s", result.Name)
			if result.Pass {
				pass++
				o.Green("--- PASS")
			} else {
				fail++
				o.Red("--- FAIL")
			}
			if o.Verbose {
				if result.Stdout != "" {
					o.Orange(fmt.Sprintf("stdout: %s", result.Stdout))
				}
				if result.Stderr != "" {
					o.Orange(fmt.Sprintf("stderr: %s", result.Stderr))
				}
			}
			for _, s := range result.Errors {
				o.Yellow(fmt.Sprintf("Error: %s\n", s))
			}
		}
		if pass > 0 {
			o.Green(fmt.Sprintf("PASSES: %d", pass))
		}
		if fail > 0 {
			o.Red(fmt.Sprintf("FAILURES: %d", fail))
		}
		totalPass += pass
		totalFail += fail
	}
	if totalFail > 0 {
		o.Red("FAIL")
		return 1
	}
	o.Green("PASS")
	return 0
}

func (o *OutWriter) Green(s string) {
	o.Print(GREEN + s + NC)
}

func (o *OutWriter) LightGreen(s string) {
	o.Print(LIGHT_GREEN + s + NC)
}

func (o *OutWriter) Yellow(s string) {
	o.Print(YELLOW + s + NC)
}

func (o *OutWriter) Red(s string) {
	o.Print(RED + s + NC)
}

func (o *OutWriter) Cyan(s string) {
	o.Print(CYAN + s + NC)
}

func (o *OutWriter) Orange(s string) {
	o.Print(ORANGE + s + NC)
}

func (o *OutWriter) Purple(s string) {
	o.Print(PURPLE + s + NC)
}

func (o *OutWriter) Print(s string) {
	if !o.Quiet {
		fmt.Println(s)
	}
}

func (o *OutWriter) Printf(s string, args ...interface{}) {
	o.Print(fmt.Sprintf(s, args))
}
