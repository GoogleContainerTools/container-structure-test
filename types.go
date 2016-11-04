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

var schemaVersions map[string]interface{} = map[string]interface{}{
	"1.0.0": new(StructureTestv1),
}

type SchemaVersion struct {
	SchemaVersion string
}

type Unmarshaller func([]byte, interface{}) error
