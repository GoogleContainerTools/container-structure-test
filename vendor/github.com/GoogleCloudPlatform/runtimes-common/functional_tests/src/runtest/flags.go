// Copyright 2017 Google Inc. All rights reserved.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"errors"
	"flag"
	"fmt"
	"strings"
)

type StringMapFlag map[string]string

func (m *StringMapFlag) String() string {
	return fmt.Sprintf("StringMapFlag%v", *m)
}

func (m *StringMapFlag) Set(value string) error {
	split := strings.SplitN(value, "=", 2)
	if len(split) != 2 {
		return errors.New("Invalid flag format. Value should be key=value")
	}
	if *m == nil {
		*m = make(map[string]string)
	}
	(*m)[split[0]] = split[1]
	return nil
}

func FlagStringMap(name string, usage string) *map[string]string {
	var f StringMapFlag
	flag.Var(&f, name, usage)
	return (*map[string]string)(&f)
}
