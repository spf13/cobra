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
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestGenSkillsDoc(t *testing.T) {
	buf := new(bytes.Buffer)
	config := SkillsConfig{
		Description: "Root command for testing",
	}
	if err := GenSkills(rootCmd, buf, config); err != nil {
		t.Fatal(err)
	}
	output := buf.String()

	checkStringContains(t, output, "---\nname: root\n")
	checkStringContains(t, output, "description: Root command for testing\n")
	checkStringContains(t, output, "---\n")
	checkStringContains(t, output, "# root")
	checkStringContains(t, output, rootCmd.Long)
	checkStringContains(t, output, "## Available Commands")
	checkStringContains(t, output, "`root echo` - "+echoCmd.Short)
	checkStringContains(t, output, "`root echo times` - "+timesCmd.Short)
	checkStringOmits(t, output, deprecatedCmd.Short)
	checkStringContains(t, output, "root --help")
	checkStringOmits(t, output, "references/REFERENCE.md")
}

func TestGenSkillsDocDefaultDescription(t *testing.T) {
	buf := new(bytes.Buffer)
	if err := GenSkills(rootCmd, buf, SkillsConfig{}); err != nil {
		t.Fatal(err)
	}
	output := buf.String()

	checkStringContains(t, output, "description: "+rootCmd.Long)
}

func TestGenSkillsFrontmatter(t *testing.T) {
	buf := new(bytes.Buffer)
	config := SkillsConfig{
		Name:                   "my-cli",
		Description:            "Manage widgets and gadgets",
		License:                "Apache-2.0",
		Compatibility:          "Requires git and docker",
		AllowedTools:           "Bash(git:*) Read",
		DisableModelInvocation: true,
		Metadata: map[string]string{
			"author":  "test-org",
			"version": "1.0",
		},
	}
	if err := GenSkills(rootCmd, buf, config); err != nil {
		t.Fatal(err)
	}
	output := buf.String()

	checkStringContains(t, output, "name: my-cli\n")
	checkStringContains(t, output, "description: Manage widgets and gadgets\n")
	checkStringContains(t, output, "license: Apache-2.0\n")
	checkStringContains(t, output, "compatibility: Requires git and docker\n")
	checkStringContains(t, output, "disable-model-invocation: true\n")
	checkStringContains(t, output, "allowed-tools: Bash(git:*) Read\n")
	checkStringContains(t, output, "metadata:\n")
	checkStringContains(t, output, "  author: test-org\n")
	checkStringContains(t, output, "  version: 1.0\n")
}

func TestGenSkillsDir(t *testing.T) {
	tmpdir, err := os.MkdirTemp("", "test-gen-skills")
	if err != nil {
		t.Fatalf("Failed to create tmpdir: %v", err)
	}
	defer os.RemoveAll(tmpdir)

	config := SkillsConfig{
		Name:        "my-tool",
		Description: "A test tool",
		Notes: []string{
			"Most list commands support -o json.",
		},
	}
	if err := GenSkillsDir(rootCmd, tmpdir, config); err != nil {
		t.Fatalf("GenSkillsDir failed: %v", err)
	}

	skillFile := filepath.Join(tmpdir, "my-tool", "SKILL.md")
	if _, err := os.Stat(skillFile); err != nil {
		t.Fatalf("Expected file 'my-tool/SKILL.md' to exist")
	}

	skill, err := os.ReadFile(skillFile)
	if err != nil {
		t.Fatalf("Failed to read SKILL.md: %v", err)
	}
	skillContent := string(skill)
	checkStringContains(t, skillContent, "name: my-tool\n")
	checkStringContains(t, skillContent, "# root")
	checkStringContains(t, skillContent, "references/root_echo.md")
	checkStringContains(t, skillContent, "references/root.md")
	checkStringOmits(t, skillContent, "### Examples")
	checkStringOmits(t, skillContent, "### Options")
	checkStringContains(t, skillContent, "## Notes")
	checkStringContains(t, skillContent, "- Most list commands support -o json.")

	refDir := filepath.Join(tmpdir, "my-tool", "references")
	expectedRefs := []string{
		"root.md",
		"root_echo.md",
		"root_echo_echosub.md",
		"root_echo_times.md",
	}
	for _, name := range expectedRefs {
		if _, err := os.Stat(filepath.Join(refDir, name)); err != nil {
			t.Fatalf("Expected reference file %q to exist", name)
		}
	}

	echoRef, err := os.ReadFile(filepath.Join(refDir, "root_echo.md"))
	if err != nil {
		t.Fatalf("Failed to read root_echo.md: %v", err)
	}
	echoContent := string(echoRef)
	checkStringContains(t, echoContent, "# root echo")
	checkStringContains(t, echoContent, echoCmd.Long)
	checkStringContains(t, echoContent, echoCmd.Example)
	checkStringContains(t, echoContent, "boolone")
	checkStringContains(t, echoContent, "### Tips")
	checkStringContains(t, echoContent, "- Supports JSON output via -o json.")

	timesRef, err := os.ReadFile(filepath.Join(refDir, "root_echo_times.md"))
	if err != nil {
		t.Fatalf("Failed to read root_echo_times.md: %v", err)
	}
	timesContent := string(timesRef)
	checkStringContains(t, timesContent, "# root echo times")
	checkStringContains(t, timesContent, timesCmd.Short)
	checkStringContains(t, timesContent, "Options inherited from parent commands")
	checkStringOmits(t, timesContent, "### Tips")
}

func TestGenSkillsDirDefaultName(t *testing.T) {
	tmpdir, err := os.MkdirTemp("", "test-gen-skills")
	if err != nil {
		t.Fatalf("Failed to create tmpdir: %v", err)
	}
	defer os.RemoveAll(tmpdir)

	if err := GenSkillsDir(rootCmd, tmpdir, SkillsConfig{}); err != nil {
		t.Fatalf("GenSkillsDir failed: %v", err)
	}

	filename := filepath.Join(tmpdir, "root", "SKILL.md")
	if _, err := os.Stat(filename); err != nil {
		t.Fatalf("Expected file 'root/SKILL.md' to exist")
	}
}

func TestGenSkillsSingleCommand(t *testing.T) {
	cmd := &cobra.Command{
		Use:   "solo",
		Short: "A standalone command",
		Long:  "A standalone command with no subcommands",
		Run:   emptyRun,
	}

	buf := new(bytes.Buffer)
	if err := GenSkills(cmd, buf, SkillsConfig{}); err != nil {
		t.Fatal(err)
	}
	output := buf.String()

	checkStringContains(t, output, "name: solo\n")
	checkStringContains(t, output, "# solo")
	checkStringContains(t, output, "A standalone command with no subcommands")
	checkStringOmits(t, output, "## Available Commands")
}

func TestToSkillName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"MyApp", "myapp"},
		{"my-app", "my-app"},
		{"my_app", "my-app"},
		{"My App", "my-app"},
		{"--my--app--", "my-app"},
	}
	for _, tt := range tests {
		got := toSkillName(tt.input)
		if got != tt.expected {
			t.Errorf("toSkillName(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestYamlEscapeString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"simple", "simple"},
		{"has: colon", `"has: colon"`},
		{"has #comment", `"has #comment"`},
	}
	for _, tt := range tests {
		got := yamlEscapeString(tt.input)
		if got != tt.expected {
			t.Errorf("yamlEscapeString(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}

	input := `has "quotes"`
	got := yamlEscapeString(input)
	if !strings.HasPrefix(got, `"`) || !strings.HasSuffix(got, `"`) {
		t.Errorf("yamlEscapeString(%q) should be quoted, got %q", input, got)
	}
}

func TestGenSkillsNotes(t *testing.T) {
	buf := new(bytes.Buffer)
	config := SkillsConfig{
		Notes: []string{
			"Most list commands support `-o json` for machine-readable output.",
			"Use `--workspace` to target a specific workspace.",
		},
	}
	if err := GenSkills(rootCmd, buf, config); err != nil {
		t.Fatal(err)
	}
	output := buf.String()

	checkStringContains(t, output, "## Notes")
	checkStringContains(t, output, "- Most list commands support `-o json` for machine-readable output.")
	checkStringContains(t, output, "- Use `--workspace` to target a specific workspace.")
}

func TestGenSkillsNoNotesSection(t *testing.T) {
	buf := new(bytes.Buffer)
	if err := GenSkills(rootCmd, buf, SkillsConfig{}); err != nil {
		t.Fatal(err)
	}
	output := buf.String()

	checkStringOmits(t, output, "## Notes")
}

func TestGenRefFileWithTips(t *testing.T) {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List items",
		Long:  "List all items in the workspace",
		Run:   emptyRun,
		Annotations: map[string]string{
			"skills:tip:output": "Use `-o json` for machine-readable output.",
			"skills:tip:mode":   "Use `--mode draft` to see unpublished changes.",
		},
	}

	buf := new(bytes.Buffer)
	if err := genRefFile(cmd, buf); err != nil {
		t.Fatal(err)
	}
	output := buf.String()

	checkStringContains(t, output, "### Tips")
	checkStringContains(t, output, "- Use `--mode draft` to see unpublished changes.")
	checkStringContains(t, output, "- Use `-o json` for machine-readable output.")
}

func TestGenRefFileWithoutTips(t *testing.T) {
	cmd := &cobra.Command{
		Use:   "pull",
		Short: "Pull items",
		Run:   emptyRun,
	}

	buf := new(bytes.Buffer)
	if err := genRefFile(cmd, buf); err != nil {
		t.Fatal(err)
	}
	output := buf.String()

	checkStringOmits(t, output, "### Tips")
}

func TestCollectTips(t *testing.T) {
	cmd := &cobra.Command{
		Use:   "test",
		Short: "Test command",
		Run:   emptyRun,
		Annotations: map[string]string{
			"skills:tip:output": "Use `-o json` for machine-readable output.",
			"skills:tip:slug":   "Pass slug as positional argument or `--slug` flag.",
			"unrelated":         "should be ignored",
		},
	}
	tips := collectTips(cmd)
	if len(tips) != 2 {
		t.Fatalf("expected 2 tips, got %d", len(tips))
	}
	// Should be sorted by key for deterministic output
	if tips[0] != "Use `-o json` for machine-readable output." {
		t.Errorf("unexpected first tip: %s", tips[0])
	}
	if tips[1] != "Pass slug as positional argument or `--slug` flag." {
		t.Errorf("unexpected second tip: %s", tips[1])
	}
}

func TestCollectTipsEmpty(t *testing.T) {
	cmd := &cobra.Command{
		Use:   "test",
		Short: "Test command",
		Run:   emptyRun,
	}
	tips := collectTips(cmd)
	if len(tips) != 0 {
		t.Fatalf("expected 0 tips, got %d", len(tips))
	}
}

func BenchmarkGenSkillsToFile(b *testing.B) {
	file, err := os.CreateTemp("", "")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(file.Name())
	defer file.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := GenSkills(rootCmd, file, SkillsConfig{}); err != nil {
			b.Fatal(err)
		}
	}
}
