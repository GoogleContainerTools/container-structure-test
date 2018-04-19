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

package output

const (
	RED         = "\033[0;31m"
	LIGHT_RED   = "\033[1;31m"
	GREEN       = "\033[0;32m"
	LIGHT_GREEN = "\033[1;32m"
	YELLOW      = "\033[1;33m"
	ORANGE      = "\033[0;33m"
	CYAN        = "\033[0;36m"
	BLUE        = "\033[0;34m"
	PURPLE      = "\033[0;35m"
	NC          = "\033[0m" // No Color
)

// ANSI Color Escape Codes
// Black        0;30     Dark Gray     1;30
// Red          0;31     Light Red     1;31
// Green        0;32     Light Green   1;32
// Brown/Orange 0;33     Yellow        1;33
// Blue         0;34     Light Blue    1;34
// Purple       0;35     Light Purple  1;35
// Cyan         0;36     Light Cyan    1;36

func green(s string) string {
	return GREEN + s + NC
}

func blue(s string) string {
	return BLUE + s + NC
}

func lightGreen(s string) string {
	return LIGHT_GREEN + s + NC
}

func yellow(s string) string {
	return YELLOW + s + NC
}

func red(s string) string {
	return RED + s + NC
}

func lightRed(s string) string {
	return LIGHT_RED + s + NC
}

func cyan(s string) string {
	return CYAN + s + NC
}

func orange(s string) string {
	return ORANGE + s + NC
}

func purple(s string) string {
	return PURPLE + s + NC
}
