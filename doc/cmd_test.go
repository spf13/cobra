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

package doc

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"go.yaml.in/yaml/v3"
)

func emptyRun(*cobra.Command, []string) {}

func init() {
	rootCmd.PersistentFlags().StringP("rootflag", "r", "two", "")
	rootCmd.PersistentFlags().StringP("strtwo", "t", "two", "help message for parent flag strtwo")

	echoCmd.PersistentFlags().StringP("strone", "s", "one", "help message for flag strone")
	echoCmd.PersistentFlags().BoolP("persistentbool", "p", false, "help message for flag persistentbool")
	echoCmd.Flags().IntP("intone", "i", 123, "help message for flag intone")
	echoCmd.Flags().BoolP("boolone", "b", true, "help message for flag boolone")

	timesCmd.PersistentFlags().StringP("strtwo", "t", "2", "help message for child flag strtwo")
	timesCmd.Flags().IntP("inttwo", "j", 234, "help message for flag inttwo")
	timesCmd.Flags().BoolP("booltwo", "c", false, "help message for flag booltwo")

	printCmd.PersistentFlags().StringP("strthree", "s", "three", "help message for flag strthree")
	printCmd.Flags().IntP("intthree", "i", 345, "help message for flag intthree")
	printCmd.Flags().BoolP("boolthree", "b", true, "help message for flag boolthree")

	echoCmd.AddCommand(timesCmd, echoSubCmd, deprecatedCmd)
	rootCmd.AddCommand(printCmd, echoCmd, dummyCmd)
}

var rootCmd = &cobra.Command{
	Use:   "root",
	Short: "Root short description",
	Long:  "Root long description",
	Run:   emptyRun,
}

var echoCmd = &cobra.Command{
	Use:     "echo [string to echo]",
	Aliases: []string{"say"},
	Short:   "Echo anything to the screen",
	Long:    "an utterly useless command for testing",
	Example: "Just run cobra-test echo",
}

var echoSubCmd = &cobra.Command{
	Use:   "echosub [string to print]",
	Short: "second sub command for echo",
	Long:  "an absolutely utterly useless command for testing gendocs!.",
	Run:   emptyRun,
}

var timesCmd = &cobra.Command{
	Use:        "times [# times] [string to echo]",
	SuggestFor: []string{"counts"},
	Short:      "Echo anything to the screen more times",
	Long:       `a slightly useless command for testing.`,
	Run:        emptyRun,
}

var deprecatedCmd = &cobra.Command{
	Use:        "deprecated [can't do anything here]",
	Short:      "A command which is deprecated",
	Long:       `an absolutely utterly useless command for testing deprecation!.`,
	Deprecated: "Please use echo instead",
}

var printCmd = &cobra.Command{
	Use:   "print [string to print]",
	Short: "Print anything to the screen",
	Long:  `an absolutely utterly useless command for testing.`,
}

var dummyCmd = &cobra.Command{
	Use:   "dummy [action]",
	Short: "Performs a dummy action",
}

func checkStringContains(t *testing.T, got, expected string) {
	if !strings.Contains(got, expected) {
		t.Errorf("Expected to contain: \n %v\nGot:\n %v\n", expected, got)
	}
}

func checkStringOmits(t *testing.T, got, expected string) {
	if strings.Contains(got, expected) {
		t.Errorf("Expected to not contain: \n %v\nGot: %v", expected, got)
	}
}

// loadFixture loads a YAML fixture file from __fixtures__ directory and returns its contents
func loadFixture(t *testing.T, parts ...string) []byte {
	path := filepath.Join(append([]string{"__fixtures__"}, parts...)...)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read fixture file %s: %v", path, err)
	}
	return data
}

// compareNormalizedYAML compares two YAML byte slices after normalizing them
func compareNormalizedYAML(t *testing.T, expected, actual []byte) {
	expectedNormalized, err := normalizeYAML(expected)
	if err != nil {
		t.Fatalf("Failed to normalize expected YAML: %v", err)
	}

	actualNormalized, err := normalizeYAML(actual)
	if err != nil {
		t.Fatalf("Failed to normalize actual YAML: %v", err)
	}

	if string(expectedNormalized) != string(actualNormalized) {
		t.Errorf("Generated OpenCLI does not match fixture")
		t.Logf("Expected:\n%s", string(expectedNormalized))
		t.Logf("Actual:\n%s", string(actualNormalized))
	}
}

// normalizeYAML normalizes YAML data by recursively sorting map keys
// and re-marshaling to ensure consistent ordering
func normalizeYAML(data []byte) ([]byte, error) {
	var v interface{}
	if err := yaml.Unmarshal(data, &v); err != nil {
		return nil, err
	}
	normalized := sortYAMLKeys(v)
	return yaml.Marshal(normalized)
}

// sortYAMLKeys recursively sorts map keys in a YAML structure
func sortYAMLKeys(v interface{}) interface{} {
	switch val := v.(type) {
	case map[interface{}]interface{}:
		// Create a new map with sorted keys
		sorted := make(map[interface{}]interface{})
		keys := make([]string, 0, len(val))
		keyMap := make(map[string]interface{})

		// Collect all keys and convert to strings for sorting
		for k := range val {
			keyStr := toString(k)
			keys = append(keys, keyStr)
			keyMap[keyStr] = k
		}

		// Sort keys
		sort.Strings(keys)

		// Rebuild map with sorted keys, recursively sorting values
		for _, keyStr := range keys {
			originalKey := keyMap[keyStr]
			sorted[originalKey] = sortYAMLKeys(val[originalKey])
		}
		return sorted
	case []interface{}:
		// Recursively sort elements in slices
		sorted := make([]interface{}, len(val))
		for i, item := range val {
			sorted[i] = sortYAMLKeys(item)
		}
		return sorted
	default:
		// Primitive types, return as-is
		return v
	}
}

// toString converts a key to string for sorting
func toString(k interface{}) string {
	switch v := k.(type) {
	case string:
		return v
	case int:
		return string(rune(v))
	default:
		// For other types, try to convert to string
		return ""
	}
}
