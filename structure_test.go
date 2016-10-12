package structure_tests

import (
	"bytes"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os/exec"
	"regexp"
	"testing"
)

type CommandTest struct {
	Name           string   // name of test
	Command        string   // command to run
	Flags          string   // optional flags
	ExpectedOutput []string // expected output of running command
	ExpectedError  []string // expected error from running command
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
}

type StructureTest struct {
	Commands           []CommandTest       `json:"commands"`
	FileExistenceTests []FileExistenceTest `json:"file_existence"`
	FileContentTests   []FileContentTest   `json:"file_contents"`
}

func TestRunCommand(t *testing.T) {
	for _, tt := range tests.Commands {
		t.Logf("COMMAND TEST: %s", tt.Name)
		var cmd *exec.Cmd
		if tt.Flags != "" {
			cmd = exec.Command(tt.Command, tt.Flags)
		} else {
			cmd = exec.Command(tt.Command)
		}
		var outbuf, errbuf bytes.Buffer

		cmd.Stdout = &outbuf
		cmd.Stderr = &errbuf

		err := cmd.Run()
		stdout := outbuf.String()
		stderr := errbuf.String()

		var errMessage string
		if err != nil {
			for _, errStr := range tt.ExpectedError {
				errMessage = "Expected string " + errStr + " not found in error!"
				compileAndRunRegex(errStr, stderr, t, errMessage)
			}
		}

		for _, outStr := range tt.ExpectedOutput {
			errMessage = "Expected string " + outStr + " not found in output!"
			compileAndRunRegex(outStr, stdout, t, errMessage)
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
			compileAndRunRegex(s, contents, t, errMessage)
		}
	}
}

func compileAndRunRegex(regex string, base string, t *testing.T, err string) {
	r, rErr := regexp.Compile(regex)
	if rErr != nil {
		t.Errorf("Error compiling regex %s : %s", regex, rErr.Error())
		return
	}
	if !r.MatchString(base) {
		t.Errorf(err)
	}
}

var configFile string
var tests StructureTest

func init() {
	flag.StringVar(&configFile, "config", "/workspace/structure_test.json",
		"path to the .yaml file containing test definitions.")
	flag.Parse()

	var err error
	var testJson []byte
	testJson, err = ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatal(err)
	}
	marshalErr := json.Unmarshal(testJson, &tests)
	if marshalErr != nil {
		log.Fatal(err)
	}
}
