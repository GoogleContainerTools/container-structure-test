package main

import (
	"testing"
)

type CommandTestv1 struct {
	Name           string
	Command        string
	Flags          []string
	ExpectedOutput []string
	ExcludedOutput []string
	ExpectedError  []string
	ExcludedError  []string // excluded error from running command
}

func validateCommandTestV1(t *testing.T, tt CommandTestv1) {
	if tt.Name == "" {
		t.Fatalf("Please provide a valid name for every test!")
	}
	if tt.Command == "" {
		t.Fatalf("Please provide a valid command to run for test %s", tt.Name)
	}
	t.Logf("COMMAND TEST: %s", tt.Name)
}
