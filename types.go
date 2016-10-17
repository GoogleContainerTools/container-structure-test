package main

import (
	"encoding/json"
	"io/ioutil"
	"strings"

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

func Parse(fp string, st *StructureTest) error {
	testContents, err := ioutil.ReadFile(fp)
	if err != nil {
		return err
	}

	switch {
	case strings.HasSuffix(fp, ".json"):
		if err := json.Unmarshal(testContents, &st); err != nil {
			return err
		}
	case strings.HasSuffix(fp, ".yaml"):
		if err := yaml.Unmarshal(testContents, &st); err != nil {
			return err
		}
	}
	return nil
}
