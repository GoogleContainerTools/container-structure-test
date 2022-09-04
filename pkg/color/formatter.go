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
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/crypto/ssh/terminal"
)

var (
	// ColoredOutput will check if color escape codes should be printed. This can be changed
	// for testing to an arbitrary method.
	ColoredOutput = coloredOutput
	// IsTerminal will check if the specified output stream is a terminal. This can be changed
	// for testing to an arbitrary method.
	IsTerminal = isTerminal
	// NoColor allow to force to not output colors.
	NoColor = false
)

// Color can be used to format text using ANSI escape codes so it can be printed to
// the terminal in color.
type Color int

// Define some Color instances that can format text to be displayed to the terminal in color, using ANSI escape codes.
var (
	LightRed    = Color(91)
	LightGreen  = Color(92)
	LightYellow = Color(93)
	LightBlue   = Color(94)
	LightPurple = Color(95)
	Red         = Color(31)
	Green       = Color(32)
	Yellow      = Color(33)
	Blue        = Color(34)
	Purple      = Color(35)
	Cyan        = Color(36)
	White       = Color(37)
	// None uses ANSI escape codes to reset all formatting.
	None = Color(0)

	// Default default output color for output from container-structure-test to the user.
	Default = None
)

// Fprint wraps the operands in c's ANSI escape codes, and outputs the result to
// out. If out is not a terminal, the escape codes will not be added.
// It returns the number of bytes written and any errors encountered.
func (c Color) Fprint(out io.Writer, a ...interface{}) (n int, err error) {
	if ColoredOutput(out) {
		return fmt.Fprintf(out, "\033[%dm%s\033[0m", c, fmt.Sprint(a...))
	}
	return fmt.Fprint(out, a...)
}

// Fprintln wraps the operands in c's ANSI escape codes, and outputs the result to
// out, followed by a newline. If out is not a terminal, the escape codes will not be added.
// It returns the number of bytes written and any errors encountered.
func (c Color) Fprintln(out io.Writer, a ...interface{}) (n int, err error) {
	if ColoredOutput(out) {
		return fmt.Fprintf(out, "\033[%dm%s\033[0m\n", c, strings.TrimSuffix(fmt.Sprintln(a...), "\n"))
	}
	return fmt.Fprintln(out, a...)
}

// Fprintf applies formats according to the format specifier (and the optional interfaces provided),
// wraps the result in c's ANSI escape codes, and outputs the result to
// out, followed by a newline. If out is not a terminal, the escape codes will not be added.
// It returns the number of bytes written and any errors encountered.
func (c Color) Fprintf(out io.Writer, format string, a ...interface{}) (n int, err error) {
	if ColoredOutput(out) {
		return fmt.Fprintf(out, "\033[%dm%s\033[0m", c, fmt.Sprintf(format, a...))
	}
	return fmt.Fprintf(out, format, a...)
}

// ColoredWriteCloser forces printing with colors to an io.WriteCloser.
type ColoredWriteCloser struct {
	io.WriteCloser
}

// ColoredWriter forces printing with colors to an io.Writer.
type ColoredWriter struct {
	io.Writer
}

// OverwriteDefault overwrites default color.
func OverwriteDefault(color Color) {
	Default = color
}

func coloredOutput(w io.Writer) bool {
	if NoColor {
		return false
	}
	return IsTerminal(w)
}

// This implementation comes from logrus (https://github.com/sirupsen/logrus/blob/master/terminal_check_notappengine.go),
// unfortunately logrus doesn't expose a public interface we can use to call it.
func isTerminal(w io.Writer) bool {
	if _, ok := w.(ColoredWriteCloser); ok {
		return true
	}
	if _, ok := w.(ColoredWriter); ok {
		return true
	}

	switch v := w.(type) {
	case *os.File:
		return terminal.IsTerminal(int(v.Fd()))
	default:
		return false
	}
}
