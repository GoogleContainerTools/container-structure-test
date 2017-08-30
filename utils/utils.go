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
	"bytes"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"syscall"
	"testing"

	"github.com/GoogleCloudPlatform/runtimes-common/structure_tests/types/unversioned"
)

// given a list of environment variable key/value pairs, set these in the current environment.
// also, keep track of the previous values of these vars to reset after test execution.
func SetEnvVars(t *testing.T, vars []unversioned.EnvVar) []unversioned.EnvVar {
	var originalVars []unversioned.EnvVar
	for _, env_var := range vars {
		originalVars = append(originalVars, unversioned.EnvVar{env_var.Key, os.Getenv(env_var.Key)})
		if err := os.Setenv(env_var.Key, os.ExpandEnv(env_var.Value)); err != nil {
			t.Fatalf("error setting env var: %s", err)
		}
	}
	return originalVars
}

func ResetEnvVars(t *testing.T, vars []unversioned.EnvVar) {
	for _, env_var := range vars {
		var err error
		if env_var.Value == "" {
			// if the previous value was empty string, the variable did not
			// exist in the environment; unset it
			err = os.Unsetenv(env_var.Key)
		} else {
			// otherwise, set it back to its previous value
			err = os.Setenv(env_var.Key, env_var.Value)
		}
		if err != nil {
			t.Fatalf("error resetting env var: %s", err)
		}
	}
}

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

// given an array of command parts, construct a full command and execute it against the
// current environment. a list of environment variables can be passed to be set in the
// environment before the command is executed. additionally, a boolean flag is passed
// to specify whether or not we care about the output of the command.
func ProcessCommand(t *testing.T, envVars []unversioned.EnvVar, fullCommand []string,
	shellMode bool, checkOutput bool) (string, string, int) {
	var cmd *exec.Cmd
	if len(fullCommand) == 0 {
		t.Logf("empty command provided: skipping...")
		return "", "", -1
	}
	var command string
	var flags []string
	if shellMode {
		command = "/bin/sh"
		flags = []string{"-c", strings.Join(fullCommand, " ")}
	} else {
		command = fullCommand[0]
		flags = fullCommand[1:]
	}
	originalVars := SetEnvVars(t, envVars)
	defer ResetEnvVars(t, originalVars)
	if len(flags) > 0 {
		cmd = exec.Command(command, flags...)
	} else {
		cmd = exec.Command(command)
	}

	if checkOutput {
		t.Logf("Executing: %s", cmd.Args)
	} else {
		t.Logf("Executing setup/teardown: %s", cmd.Args)
	}

	var outbuf, errbuf bytes.Buffer

	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf

	err := cmd.Run()
	stdout := outbuf.String()
	if stdout != "" {
		t.Logf("stdout: %s", stdout)
	}
	stderr := errbuf.String()
	if stderr != "" {
		t.Logf("stderr: %s", stderr)
	}
	var exitCode int
	if err != nil {
		if checkOutput {
			// The test might be designed to run a command that exits with an error.
			t.Logf("Error running command: %s. Continuing.", err)
		} else {
			t.Fatalf("Error running setup/teardown command: %s.", err)
		}
		switch err := err.(type) {
		default:
			t.Errorf("Command failed to start! Unable to retrieve error info!")
		case *exec.ExitError:
			exitCode = err.Sys().(syscall.WaitStatus).ExitStatus()
		case *exec.Error:
			// Command started but failed to finish, so we can at least check the stderr
			stderr = err.Error()
		}
	} else {
		exitCode = cmd.ProcessState.Sys().(syscall.WaitStatus).ExitStatus()
	}
	return stdout, stderr, exitCode
}
