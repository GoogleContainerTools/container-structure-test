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

package version

import (
	"fmt"
	"runtime"
)

// The current version of container-structure-test
// This is a private field and is set through a compilation flag from the Makefile

var version = "v0.0.0-unset"

var (
	buildDate string
	platform  = fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
)

type Info struct {
	Version    string
	GitVersion string
	BuildDate  string
	GoVersion  string
	Compiler   string
	Platform   string
}

// Get returns the version and buildtime information about the binary.
func GetVersion() *Info {
	// These variables typically come from -ldflags settings to `go build`
	return &Info{
		Version:   version,
		BuildDate: buildDate,
		GoVersion: runtime.Version(),
		Compiler:  runtime.Compiler,
		Platform:  platform,
	}
}

// func GetVersion() string {
// 	return version
// }
