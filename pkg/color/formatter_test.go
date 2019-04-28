/*
Copyright 2019 Google Inc. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package color

import (
	"bytes"
	"io"
	"testing"

	"github.com/GoogleContainerTools/container-structure-test/testutil"
)

func compareText(t *testing.T, expected, actual string, expectedN int, actualN int, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("did not expect error when formatting text but got %s", err)
	}
	if actualN != expectedN {
		t.Errorf("expected formatter to have written %d bytes but wrote %d", expectedN, actualN)
	}
	if actual != expected {
		t.Errorf("formatting not applied to text. Expected \"%s\" but got \"%s\"", expected, actual)
	}
}

func TestFprint(t *testing.T) {
	defer func(f func(io.Writer) bool) { ColoredOutput = f }(ColoredOutput)
	ColoredOutput = func(io.Writer) bool { return true }

	var b bytes.Buffer
	n, err := Green.Fprint(&b, "It's not easy being")
	expected := "\033[32mIt's not easy being\033[0m"
	compareText(t, expected, b.String(), 28, n, err)
}

func TestFprintln(t *testing.T) {
	defer func(f func(io.Writer) bool) { ColoredOutput = f }(ColoredOutput)
	ColoredOutput = func(io.Writer) bool { return true }

	var b bytes.Buffer
	n, err := Green.Fprintln(&b, "2", "less", "chars!")
	expected := "\033[32m2 less chars!\033[0m\n"
	compareText(t, expected, b.String(), 23, n, err)
}

func TestFprintf(t *testing.T) {
	defer func(f func(io.Writer) bool) { ColoredOutput = f }(ColoredOutput)
	ColoredOutput = func(io.Writer) bool { return true }

	var b bytes.Buffer
	n, err := Green.Fprintf(&b, "It's been %d %s", 1, "week")
	expected := "\033[32mIt's been 1 week\033[0m"
	compareText(t, expected, b.String(), 25, n, err)
}

func TestFprintNoTTY(t *testing.T) {
	var b bytes.Buffer
	expected := "It's not easy being"
	n, err := Green.Fprint(&b, expected)
	compareText(t, expected, b.String(), 19, n, err)
}

func TestFprintlnNoTTY(t *testing.T) {
	var b bytes.Buffer
	n, err := Green.Fprintln(&b, "2", "less", "chars!")
	expected := "2 less chars!\n"
	compareText(t, expected, b.String(), 14, n, err)
}

func TestFprintfNoTTY(t *testing.T) {
	var b bytes.Buffer
	n, err := Green.Fprintf(&b, "It's been %d %s", 1, "week")
	expected := "It's been 1 week"
	compareText(t, expected, b.String(), 16, n, err)
}

func TestFprintTTYNoColor(t *testing.T) {
	defer func(f func(io.Writer) bool) { IsTerminal = f }(IsTerminal)
	IsTerminal = func(io.Writer) bool { return true }
	defer func() { NoColor = false }()
	NoColor = true

	var b bytes.Buffer
	expected := "It's not easy being"
	n, err := Green.Fprint(&b, expected)
	compareText(t, expected, b.String(), 19, n, err)
}

func TestFprintlnTTYNoColor(t *testing.T) {
	defer func(f func(io.Writer) bool) { IsTerminal = f }(IsTerminal)
	IsTerminal = func(io.Writer) bool { return true }
	defer func() { NoColor = false }()
	NoColor = true

	var b bytes.Buffer
	n, err := Green.Fprintln(&b, "2", "less", "chars!")
	expected := "2 less chars!\n"
	compareText(t, expected, b.String(), 14, n, err)
}

func TestFprintfTTYNoColor(t *testing.T) {
	defer func(f func(io.Writer) bool) { IsTerminal = f }(IsTerminal)
	IsTerminal = func(io.Writer) bool { return true }
	defer func() { NoColor = false }()
	NoColor = true

	var b bytes.Buffer
	n, err := Green.Fprintf(&b, "It's been %d %s", 1, "week")
	expected := "It's been 1 week"
	compareText(t, expected, b.String(), 16, n, err)
}

func TestOverwriteDefault(t *testing.T) {
	testutil.CheckDeepEqual(t, None, Default)
	OverwriteDefault(Red)
	testutil.CheckDeepEqual(t, Red, Default)
}
