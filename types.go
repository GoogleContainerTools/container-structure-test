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
	Name           string   // required
	Command        string   // required
	Flags          []string // optional
	ExpectedOutput []string // optional
	ExcludedOutput []string // optional
	ExpectedError  []string // optional
	ExcludedError  []string // optional
}

type FileExistenceTest struct {
	Name        string // required
	Path        string // required
	IsDirectory bool   // required
	ShouldExist bool   // required
}

type FileContentTest struct {
	Name             string   // required
	Path             string   // required
	ExpectedContents []string // optional
	ExcludedContents []string // optional
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
