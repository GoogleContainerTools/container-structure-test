// Copyright 2018 Google Inc. All rights reserved.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package types

import (
	"github.com/GoogleContainerTools/container-structure-test/pkg/drivers"
	v1 "github.com/GoogleContainerTools/container-structure-test/pkg/types/v1"
	v2 "github.com/GoogleContainerTools/container-structure-test/pkg/types/v2"
)

type StructureTest interface {
	SetDriverImpl(func(drivers.DriverConfig) (drivers.Driver, error), drivers.DriverConfig)
	NewDriver() (drivers.Driver, error)
	RunAll(chan interface{}, string)
}

var SchemaVersions map[string]func() StructureTest = map[string]func() StructureTest{
	"1.0.0": func() StructureTest { return new(v1.StructureTest) },
	"2.0.0": func() StructureTest { return new(v2.StructureTest) },
}

type SchemaVersion struct {
	SchemaVersion string `yaml:"schemaVersion"`
}

type Unmarshaller func([]byte, interface{}) error
