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
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"
)

const FAILED = "FAILED"
const PASSED = "PASSED"

var verbose = flag.Bool("verbose", false, "Verbose logging")
var vars = FlagStringMap("vars", "Variable substitutions")

func main() {
	testSpec := flag.String("test_spec", "", "Path to a yaml or json file containing the test spec")
	flag.Parse()

	if *testSpec == "" {
		log.Fatal("--test_spec must be specified")
	}

	suite := LoadSuite(*testSpec)
	doSetup(suite)
	report := doRunTests(suite)
	reportAndExit := func() {
		failureCount := report()
		os.Exit(failureCount)
	}
	defer reportAndExit()
	doTeardown(suite)
}

func info(text string, arg ...interface{}) {
	log.Printf(text, arg...)
}

func runCommand(name string, args ...string) (err error, stdout string, stderr string) {
	expandedName, expandedArgs := expandCommand(name, args...)
	if *verbose {
		originalCommand := fmt.Sprintf("%v", append([]string{name}, args...))
		expandedCommand := fmt.Sprintf("%v", append([]string{expandedName}, expandedArgs...))
		if originalCommand == expandedCommand {
			info("Running command: %s", originalCommand)
		} else {
			info("Running command (original): %s", originalCommand)
			info("Running command (expanded): %s", expandedCommand)
		}
	}

	cmd := exec.Command(expandedName, expandedArgs...)
	var stdoutBuffer bytes.Buffer
	var stderrBuffer bytes.Buffer
	cmd.Stdout = &stdoutBuffer
	cmd.Stderr = &stderrBuffer
	err = cmd.Run()
	stdout = stdoutBuffer.String()
	stderr = stderrBuffer.String()
	if *verbose {
		commandOutput("STDOUT", stdout)
		commandOutput("STDERR", stderr)
	}
	return
}

func expandCommand(name string, args ...string) (nameOut string, argsOut []string) {
	mapping := func(key string) string {
		return (*vars)[key]
	}
	nameOut = os.Expand(name, mapping)
	argsOut = make([]string, 0, len(args))
	for _, arg := range args {
		argsOut = append(argsOut, os.Expand(arg, mapping))
	}
	return
}

func commandOutput(name string, content string) {
	if len(content) <= 0 {
		return
	}

	info("%s>>%s<<%s", name, content, name)
}

func doSetup(suite Suite) {
	info(">>> Setting up...")
	for _, setup := range suite.Setup {
		err, _, _ := runCommand(setup.Command[:1][0], setup.Command[1:]...)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func doTeardown(suite Suite) {
	info(">>> Tearing down...")
	for _, teardown := range suite.Teardown {
		err, _, _ := runCommand(teardown.Command[:1][0], teardown.Command[1:]...)
		if err != nil {
			info(" > Warning: Teardown command failed: %s", err)
		}
	}
}

// doRunTests executes the tests in the suite and returns a function
// to do a summary report. This report function returns the number of
// test failures, i.e. its returning 0 means all tests are passing.
func doRunTests(suite Suite) func() int {
	info(">>> Testing...")
	results := make(map[int]string)
	passing := make(map[int]bool)
	for index, test := range suite.Tests {
		doOneTest(index, test, suite, results, passing)
	}

	report := func() int {
		failureCount := 0
		for index := range suite.Tests {
			if !passing[index] {
				failureCount++
			}
		}
		if failureCount == 0 {
			info(">>> Summary: %s", PASSED)
		} else {
			info(">>> Summary: %s", FAILED)
		}
		for index := range suite.Tests {
			info(" > %s", results[index])
		}
		return failureCount
	}
	return report
}

func doOneTest(index int, test Test, suite Suite, results map[int]string, passing map[int]bool) {
	var name string
	var msg string
	var result string

	if len(test.Name) > 0 {
		name = fmt.Sprintf("test-%d (%s)", index, test.Name)
	} else {
		name = fmt.Sprintf("test-%d", index)
	}
	info(" > %s", name)

	recordResult := func(b *string) {
		if *b == PASSED {
			results[index] = fmt.Sprintf("%s: %s", name, PASSED)
			passing[index] = true
		} else {
			results[index] = fmt.Sprintf("%s: %s", name, FAILED)
			passing[index] = false
		}
		info(" %s", *b)
	}
	defer recordResult(&result)

	args := append([]string{"exec", suite.Target}, test.Command...)
	err, stdout, stderr := runCommand("docker", args...)

	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				result = fmt.Sprintf("%s: Exit status %d", FAILED, status.ExitStatus())
			}
		} else {
			result = fmt.Sprintf("%s: Encountered error: %v", FAILED, err)
		}
		if len(stdout) > 0 {
			result = fmt.Sprintf("%s\nSTDOUT>>>%s<<<STDOUT", result, stdout)
		}
		if len(stderr) > 0 {
			result = fmt.Sprintf("%s\nSTDERR>>>%s<<<STDERR", result, stderr)
		}
		return
	}

	msg = DoStringAssert(stdout, test.Expect.Stdout)
	if len(msg) > 0 {
		result = fmt.Sprintf("%s: stdout assertion failure\n%s", FAILED, msg)
		return
	}
	msg = DoStringAssert(stderr, test.Expect.Stderr)
	if len(msg) > 0 {
		result = fmt.Sprintf("%s: stderr assertion failure\n%s", FAILED, msg)
		return
	}

	result = PASSED
}
