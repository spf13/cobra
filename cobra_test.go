// Copyright 2013-2023 The Cobra Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cobra

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"text/template"
)

func assertNoErr(t *testing.T, e error) {
	if e != nil {
		t.Error(e)
	}
}

func TestAddTemplateFunctions(t *testing.T) {
	AddTemplateFunc("t", func() bool { return true })
	AddTemplateFuncs(template.FuncMap{
		"f": func() bool { return false },
		"h": func() string { return "Hello," },
		"w": func() string { return "world." }})

	c := &Command{}
	c.SetUsageTemplate(`{{if t}}{{h}}{{end}}{{if f}}{{h}}{{end}} {{w}}`)

	const expected = "Hello, world."
	if got := c.UsageString(); got != expected {
		t.Errorf("Expected UsageString: %v\nGot: %v", expected, got)
	}
}

func TestLevenshteinDistance(t *testing.T) {
	tests := []struct {
		name       string
		s          string
		t          string
		ignoreCase bool
		expected   int
	}{
		{
			name:       "Equal strings (case-sensitive)",
			s:          "hello",
			t:          "hello",
			ignoreCase: false,
			expected:   0,
		},
		{
			name:       "Equal strings (case-insensitive)",
			s:          "Hello",
			t:          "hello",
			ignoreCase: true,
			expected:   0,
		},
		{
			name:       "Different strings (case-sensitive)",
			s:          "kitten",
			t:          "sitting",
			ignoreCase: false,
			expected:   3,
		},
		{
			name:       "Different strings (case-insensitive)",
			s:          "Kitten",
			t:          "Sitting",
			ignoreCase: true,
			expected:   3,
		},
		{
			name:       "Empty strings",
			s:          "",
			t:          "",
			ignoreCase: false,
			expected:   0,
		},
		{
			name:       "One empty string",
			s:          "abc",
			t:          "",
			ignoreCase: false,
			expected:   3,
		},
		{
			name:       "Both empty strings",
			s:          "",
			t:          "",
			ignoreCase: true,
			expected:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			got := ld(tt.s, tt.t, tt.ignoreCase)

			// Assert
			if got != tt.expected {
				t.Errorf("Expected ld: %v\nGot: %v", tt.expected, got)
			}
		})
	}
}

func TestStringInSlice(t *testing.T) {
	tests := []struct {
		name     string
		a        string
		list     []string
		expected bool
	}{
		{
			name:     "String in slice (case-sensitive)",
			a:        "apple",
			list:     []string{"orange", "banana", "apple", "grape"},
			expected: true,
		},
		{
			name:     "String not in slice (case-sensitive)",
			a:        "pear",
			list:     []string{"orange", "banana", "apple", "grape"},
			expected: false,
		},
		{
			name:     "String in slice (case-insensitive)",
			a:        "APPLE",
			list:     []string{"orange", "banana", "apple", "grape"},
			expected: false,
		},
		{
			name:     "Empty slice",
			a:        "apple",
			list:     []string{},
			expected: false,
		},
		{
			name:     "Empty string",
			a:        "",
			list:     []string{"orange", "banana", "apple", "grape"},
			expected: false,
		},
		{
			name:     "Empty strings match",
			a:        "",
			list:     []string{"orange", ""},
			expected: true,
		},
		{
			name:     "Empty string in empty slice",
			a:        "",
			list:     []string{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			got := stringInSlice(tt.a, tt.list)

			// Assert
			if got != tt.expected {
				t.Errorf("Expected stringInSlice: %v\nGot: %v", tt.expected, got)
			}
		})
	}
}

func TestRpad(t *testing.T) {
	tests := []struct {
		name        string
		inputString string
		padding     int
		expected    string
	}{
		{
			name:        "Padding required",
			inputString: "Hello",
			padding:     10,
			expected:    "Hello     ",
		},
		{
			name:        "No padding required",
			inputString: "World",
			padding:     5,
			expected:    "World",
		},
		{
			name:        "Empty string",
			inputString: "",
			padding:     8,
			expected:    "        ",
		},
		{
			name:        "Zero padding",
			inputString: "cobra",
			padding:     0,
			expected:    "cobra",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			got := rpad(tt.inputString, tt.padding)

			// Assert
			if got != tt.expected {
				t.Errorf("Expected rpad: %v\nGot: %v", tt.expected, got)
			}
		})
	}
}

// TestDeadcodeElimination checks that a simple program using cobra in its
// default configuration is linked taking full advantage of the linker's
// deadcode elimination step.
//
// If reflect.Value.MethodByName/reflect.Value.Method are reachable the
// linker will not always be able to prove that exported methods are
// unreachable, making deadcode elimination less effective. Using
// text/template and html/template makes reflect.Value.MethodByName
// reachable.
// Since cobra can use text/template templates this test checks that in its
// default configuration that code path can be proven to be unreachable by
// the linker.
//
// See also: https://github.com/spf13/cobra/pull/1956
func TestDeadcodeElimination(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("go tool nm fails on windows")
	}

	// check that a simple program using cobra in its default configuration is
	// linked with deadcode elimination enabled.
	const (
		dirname  = "test_deadcode"
		progname = "test_deadcode_elimination"
	)
	_ = os.Mkdir(dirname, 0770)
	defer os.RemoveAll(dirname)
	filename := filepath.Join(dirname, progname+".go")
	err := os.WriteFile(filename, []byte(`package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Version: "1.0",
	Use:     "example_program",
	Short:   "example_program - test fixture to check that deadcode elimination is allowed",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("hello world")
	},
	Aliases: []string{"alias1", "alias2"},
	Example: "stringer --help",
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Whoops. There was an error while executing your CLI '%s'", err)
		os.Exit(1)
	}
}
`), 0600)
	if err != nil {
		t.Fatalf("could not write test program: %v", err)
	}
	buf, err := exec.Command("go", "build", filename).CombinedOutput()
	if err != nil {
		t.Fatalf("could not compile test program: %s", string(buf))
	}
	defer os.Remove(progname)
	buf, err = exec.Command("go", "tool", "nm", progname).CombinedOutput()
	if err != nil {
		t.Fatalf("could not run go tool nm: %v", err)
	}
	if strings.Contains(string(buf), "MethodByName") {
		t.Error("compiled programs contains MethodByName symbol")
	}
}
