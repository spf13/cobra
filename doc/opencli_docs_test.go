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
	"testing"

	"github.com/spf13/cobra"
	"go.yaml.in/yaml/v3"
)

func TestGenOpenCLIRootCommand(t *testing.T) {
	opencli, err := ExperimentalGenOpenCLI(rootCmd)
	if err != nil {
		t.Fatal(err)
	}

	// Check OpenCLI structure
	if opencli.Opencli == "" {
		t.Error("Expected opencli version to be set")
	}

	// Check root command info
	if opencli.Info.Title != rootCmd.Name() {
		t.Errorf("Expected title %q, got %q", rootCmd.Name(), opencli.Info.Title)
	}
	if rootCmd.Short != "" && (opencli.Info.Summary == nil || *opencli.Info.Summary != rootCmd.Short) {
		t.Errorf("Expected summary %q, got %v", rootCmd.Short, opencli.Info.Summary)
	}

	// Check commands - rootCmd has 3 commands: printCmd, echoCmd, dummyCmd
	expectedCommandCount := 3
	if len(opencli.Commands) != expectedCommandCount {
		t.Errorf("Expected %d commands, got %d", expectedCommandCount, len(opencli.Commands))
	}

	// Check root-level options - rootCmd has 2 persistent flags: rootflag, strtwo, plus help flag
	expectedRootOptionCount := 3 // rootflag, strtwo, help
	if len(opencli.Options) != expectedRootOptionCount {
		t.Errorf("Expected %d root options, got %d", expectedRootOptionCount, len(opencli.Options))
	}

	// Check for root-level flags
	var foundRootflag, foundStrtwo bool
	for _, opt := range opencli.Options {
		if opt.Name == "rootflag" {
			foundRootflag = true
		}
		if opt.Name == "strtwo" {
			foundStrtwo = true
		}
	}
	if !foundRootflag {
		t.Error("Expected to find rootflag option in root options")
	}
	if !foundStrtwo {
		t.Error("Expected to find strtwo option in root options")
	}

	//TODO:
}

func TestGenOpenCLIWithVersion(t *testing.T) {
	testCmd := &cobra.Command{
		Use:     "test",
		Short:   "Test command",
		Version: "1.2.3",
	}

	opencli, err := ExperimentalGenOpenCLI(testCmd)
	if err != nil {
		t.Fatal(err)
	}

	if opencli.Info.Version != "1.2.3" {
		t.Errorf("Expected version %q, got %q", "1.2.3", opencli.Info.Version)
	}

	// Check commands length - should have 0 commands (no subcommands added)
	if len(opencli.Commands) != 0 {
		t.Errorf("Expected 0 commands, got %d", len(opencli.Commands))
	}

	// Check options length - should have 1 option (help flag is added by InitDefaultHelpFlag)
	if len(opencli.Options) != 1 {
		t.Errorf("Expected 1 option (help flag), got %d", len(opencli.Options))
	}
}

func TestGenOpenCLIWithArguments(t *testing.T) {
	testCmd := &cobra.Command{
		Use:   "test [optional] required [variadic...]",
		Short: "Test command with arguments",
	}

	opencli, err := ExperimentalGenOpenCLI(testCmd)
	if err != nil {
		t.Fatal(err)
	}

	// Check that arguments are extracted - should have 3 arguments: optional, required, variadic
	expectedArgCount := 3
	if len(opencli.Arguments) != expectedArgCount {
		t.Errorf("Expected %d arguments, got %d", expectedArgCount, len(opencli.Arguments))
	}

	// Check for required argument
	var foundRequired bool
	for _, arg := range opencli.Arguments {
		if arg.Name == "required" {
			foundRequired = true
			if arg.Required == nil || !*arg.Required {
				t.Error("Expected required argument to be marked as required")
			}
		}
		if arg.Name == "optional" {
			if arg.Required == nil || *arg.Required {
				t.Error("Expected optional argument to be marked as optional")
			}
		}
		if arg.Name == "variadic" {
			if arg.Arity == nil {
				t.Error("Expected variadic argument to have arity set")
			}
		}
	}
	if !foundRequired {
		t.Error("Expected to find required argument")
	}
}

func TestGenOpenCLIWithOptions(t *testing.T) {
	testCmd := &cobra.Command{
		Use:   "test",
		Short: "Test command",
	}
	testCmd.Flags().StringP("flag1", "f", "default", "flag 1")
	testCmd.Flags().BoolP("flag2", "b", false, "flag 2")

	opencli, err := ExperimentalGenOpenCLI(testCmd)
	if err != nil {
		t.Fatal(err)
	}

	// Check options length - should have 3 options (flag1, flag2, help)
	expectedOptionCount := 3
	if len(opencli.Options) != expectedOptionCount {
		t.Errorf("Expected %d options, got %d", expectedOptionCount, len(opencli.Options))
	}

	// Check for specific flags
	var foundFlag1, foundFlag2 bool
	for _, opt := range opencli.Options {
		if opt.Name == "flag1" {
			foundFlag1 = true
		}
		if opt.Name == "flag2" {
			foundFlag2 = true
		}
	}
	if !foundFlag1 {
		t.Error("Expected to find flag1 option")
	}
	if !foundFlag2 {
		t.Error("Expected to find flag2 option")
	}
}

func TestGenOpenCLIYAMLFormat(t *testing.T) {
	opencli, err := ExperimentalGenOpenCLI(rootCmd)
	if err != nil {
		t.Fatal(err)
	}

	// Should be valid YAML structure with required fields
	if opencli.Opencli == "" {
		t.Error("Expected opencli version to be set")
	}

	if opencli.Info.Title == "" {
		t.Error("Expected info.title to be set")
	}

	// Version can be empty string (it's valid, just not set on rootCmd)
	// The field is required in the struct but can be empty string
}

func BenchmarkGenOpenCLIToFile(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ExperimentalGenOpenCLI(rootCmd)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestGenOpenCLIWithFixtures(t *testing.T) {
	fixtureData := loadFixture(t, "opencli", "1.rootCmd", "output.yaml")
	actualData := generateOpenCLI(t, rootCmd)
	compareNormalizedYAML(t, fixtureData, actualData)
}

// generateOpenCLI generates OpenCLI YAML from a command and returns the bytes
func generateOpenCLI(t *testing.T, cmd *cobra.Command) []byte {
	opencli, err := ExperimentalGenOpenCLI(cmd)
	if err != nil {
		t.Fatalf("Failed to generate OpenCLI: %v", err)
	}
	data, err := yaml.Marshal(opencli)
	if err != nil {
		t.Fatalf("Failed to marshal OpenCLI to YAML: %v", err)
	}
	return data
}
