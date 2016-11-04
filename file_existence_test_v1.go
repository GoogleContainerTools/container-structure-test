package main

import (
	"testing"
)

type FileExistenceTestv1 struct {
	Name        string // name of test
	Path        string // file to check existence of
	IsDirectory bool   // whether or not the path points to a directory
	ShouldExist bool   // whether or not the file should exist
}

func validateFileExistenceTestV1(t *testing.T, tt FileExistenceTestv1) {
	if tt.Name == "" {
		t.Fatalf("Please provide a valid name for every test!")
	}
	if tt.Path == "" {
		t.Fatalf("Please provide a valid file path for test %s", tt.Name)
	}
	t.Logf("FILE EXISTENCE TEST: %s", tt.Name)
}
