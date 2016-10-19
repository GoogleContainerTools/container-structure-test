package main

import (
	"encoding/json"
	"io/ioutil"
	"strings"

	"github.com/ghodss/yaml"
)

type arrayFlags []string

func (a *arrayFlags) String() string {
	return strings.Join(*a, ", ")
}

func (a *arrayFlags) Set(value string) error {
	*a = append(*a, value)
	return nil
}

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

func combineTests(tests *StructureTest, tmpTests *StructureTest) {
	tests.CommandTests = append(tests.CommandTests, tmpTests.CommandTests...)
	tests.FileExistenceTests = append(tests.FileExistenceTests, tmpTests.FileExistenceTests...)
	tests.FileContentTests = append(tests.FileContentTests, tmpTests.FileContentTests...)
}

func parseFile(tests *StructureTest, configFile string) error {
	var tmpTests StructureTest
	testContents, err := ioutil.ReadFile(configFile)
	if err != nil {
		return err
	}

	switch {
	case strings.HasSuffix(configFile, ".json"):
		if err := json.Unmarshal(testContents, &tests); err != nil {
			return err
		}
	case strings.HasSuffix(configFile, ".yaml"):
		if err := yaml.Unmarshal(testContents, &tests); err != nil {
			return err
		}
	}
	combineTests(tests, &tmpTests)
	return nil
}

func Parse(configFiles []string, tests *StructureTest) error {
	for _, file := range configFiles {
		if err := parseFile(tests, file); err != nil {
			return err
		}
	}
	return nil
}
