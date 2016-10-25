package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"regexp"
	"testing"
)

func TestRunCommand(t *testing.T) {
	for _, tt := range tests.CommandTests {
		validateCommandTest(t, tt)
		var cmd *exec.Cmd
		if tt.Flags != nil && len(tt.Flags) > 0 {
			cmd = exec.Command(tt.Command, tt.Flags...)
		} else {
			cmd = exec.Command(tt.Command)
		}
		t.Logf("Executing: %s", cmd.Args)
		var outbuf, errbuf bytes.Buffer

		cmd.Stdout = &outbuf
		cmd.Stderr = &errbuf

		if err := cmd.Run(); err != nil {
			// The test might be designed to run a command that exits with an error.
			t.Logf("Error running command: %s. Continuing.", err)
		}

		stdout := outbuf.String()
		if stdout != "" {
			t.Logf("stdout: %s", stdout)
		}
		stderr := errbuf.String()
		if stderr != "" {
			t.Logf("stderr: %s", stderr)
		}

		for _, errStr := range tt.ExpectedError {
			errMsg := fmt.Sprintf("Expected string '%s' not found in error!", errStr)
			compileAndRunRegex(errStr, stderr, t, errMsg, true)
		}
		for _, errStr := range tt.ExcludedError {
			errMsg := fmt.Sprintf("Excluded string '%s' found in error!", errStr)
			compileAndRunRegex(errStr, stderr, t, errMsg, false)
		}

		for _, outStr := range tt.ExpectedOutput {
			errMsg := fmt.Sprintf("Expected string '%s' not found in output!", outStr)
			compileAndRunRegex(outStr, stdout, t, errMsg, true)
		}
		for _, outStr := range tt.ExcludedError {
			errMsg := fmt.Sprintf("Excluded string '%s' found in output!", outStr)
			compileAndRunRegex(outStr, stdout, t, errMsg, false)
		}
	}
}

func TestFileExists(t *testing.T) {
	for _, tt := range tests.FileExistenceTests {
		validateFileExistenceTest(t, tt)
		var err error
		if tt.IsDirectory {
			_, err = ioutil.ReadDir(tt.Path)
		} else {
			_, err = ioutil.ReadFile(tt.Path)
		}
		if tt.ShouldExist && err != nil {
			t.Errorf("File %s should exist but does not!", tt.Path)
		} else if !tt.ShouldExist && err == nil {
			t.Errorf("File %s should not exist but does!", tt.Path)
		}
	}
}

func TestFileContents(t *testing.T) {
	for _, tt := range tests.FileContentTests {
		validateFileContentTest(t, tt)
		actualContents, err := ioutil.ReadFile(tt.Path)
		if err != nil {
			t.Errorf("Failed to open %s. Error: %s", tt.Path, err)
		}

		contents := string(actualContents[:])

		var errMessage string
		for _, s := range tt.ExpectedContents {
			errMessage = "Expected string " + s + " not found in file contents!"
			compileAndRunRegex(s, contents, t, errMessage, true)
		}
		for _, s := range tt.ExcludedContents {
			errMessage = "Excluded string " + s + " found in file contents!"
			compileAndRunRegex(s, contents, t, errMessage, false)
		}
	}
}

func compileAndRunRegex(regex string, base string, t *testing.T, err string, shouldMatch bool) {
	r, rErr := regexp.Compile(regex)
	if rErr != nil {
		t.Errorf("Error compiling regex %s : %s", regex, rErr.Error())
		return
	}
	if shouldMatch != r.MatchString(base) {
		t.Errorf(err)
	}
}

func validateCommandTest(t *testing.T, tt CommandTest) {
	if tt.Name == "" {
		t.Fatalf("Please provide a valid name for every test!")
	}
	if tt.Command == "" {
		t.Fatalf("Please provide a valid command to run for test %s", tt.Name)
	}
	t.Logf("COMMAND TEST: %s", tt.Name)
}

func validateFileExistenceTest(t *testing.T, tt FileExistenceTest) {
	if tt.Name == "" {
		t.Fatalf("Please provide a valid name for every test!")
	}
	if tt.Path == "" {
		t.Fatalf("Please provide a valid file path for test %s", tt.Name)
	}
	t.Logf("FILE EXISTENCE TEST: %s", tt.Name)
}

func validateFileContentTest(t *testing.T, tt FileContentTest) {
	if tt.Name == "" {
		t.Fatalf("Please provide a valid name for every test!")
	}
	if tt.Path == "" {
		t.Fatalf("Please provide a valid file path for test %s", tt.Name)
	}
	t.Logf("FILE CONTENT TEST: %s", tt.Name)
}

var configFiles arrayFlags
var tests StructureTest

func init() {
	flag.Var(&configFiles, "config", "path to the .yaml file containing test definitions.")
	flag.Parse()

	if len(configFiles) == 0 {
		configFiles = append(configFiles, "/workspace/structure_test.json")
	}

	if err := Parse(configFiles, &tests); err != nil {
		log.Fatalf("Error parsing config file: %s", err)
	}
}
