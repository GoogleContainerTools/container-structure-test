package structure_tests

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"regexp"
	"strings"
	"testing"

	"github.com/ghodss/yaml"
)

type CommandTest struct {
	Name           string
	Command        string
	Flags          string
	ExpectedOutput []string
	ExcludedOutput []string
	ExpectedError  []string
	ExcludedError  []string // excluded error from running command
}

type FileExistenceTest struct {
	Name        string // name of test
	Path        string // file to check existence of
	IsDirectory bool   // whether or not the path points to a directory
	ShouldExist bool   // whether or not the file should exist
}

type FileContentTest struct {
	Name             string   // name of test
	Path             string   // file to check existence of
	ExpectedContents []string // list of expected contents of file
	ExcludedContents []string // list of excluded contents of file
}

type StructureTest struct {
	CommandTests       []CommandTest
	FileExistenceTests []FileExistenceTest
	FileContentTests   []FileContentTest
}

func TestRunCommand(t *testing.T) {
	for _, tt := range tests.CommandTests {
		t.Logf("COMMAND TEST: %s", tt.Name)
		var cmd *exec.Cmd
		if tt.Flags != "" {
			cmd = exec.Command(tt.Command, tt.Flags)
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
		stderr := errbuf.String()

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
		t.Logf("FILE EXISTENCE TEST: %s", tt.Name)
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
		t.Logf("FILE CONTENT TEST: %s", tt.Name)
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

var configFile string
var tests StructureTest

func init() {
	flag.StringVar(&configFile, "config", "/workspace/structure_test.json",
		"path to the .yaml file containing test definitions.")
	flag.Parse()

	testContents, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatalf("Error reading file: %s. %s", configFile, err)
	}

	switch {
	case strings.HasSuffix(configFile, ".json"):
		if err := json.Unmarshal(testContents, &tests); err != nil {
			log.Fatal(err)
		}
	case strings.HasSuffix(configFile, ".yaml"):
		if err := yaml.Unmarshal(testContents, &tests); err != nil {
			log.Fatal(err)
		}
	}
}
