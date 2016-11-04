package main

import (
	"testing"
)

type FileContentTestv1 struct {
	Name             string   // name of test
	Path             string   // file to check existence of
	ExpectedContents []string // list of expected contents of file
	ExcludedContents []string // list of excluded contents of file
}

func validateFileContentTestV1(t *testing.T, tt FileContentTestv1) {
	if tt.Name == "" {
		t.Fatalf("Please provide a valid name for every test!")
	}
	if tt.Path == "" {
		t.Fatalf("Please provide a valid file path for test %s", tt.Name)
	}
	t.Logf("FILE CONTENT TEST: %s", tt.Name)
}
