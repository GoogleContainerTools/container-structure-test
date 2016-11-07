package main

import (
	"strings"
	"testing"
)

type StructureTest interface {
	RunAll(t *testing.T)
}

type arrayFlags []string

func (a *arrayFlags) String() string {
	return strings.Join(*a, ", ")
}

func (a *arrayFlags) Set(value string) error {
	*a = append(*a, value)
	return nil
}

var schemaVersions map[string]VersionHolder = map[string]VersionHolder{
	"1.0.0": new(VersionHolderv1),
}

type SchemaVersion struct {
	SchemaVersion string
}

type Unmarshaller func([]byte, interface{}) error

type VersionHolder interface {
	New() StructureTest
}

type VersionHolderv1 struct{}

func (v VersionHolderv1) New() StructureTest {
	return new(StructureTestv1)
}
