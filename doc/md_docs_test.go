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

func TestGenMdDoc(t *testing.T) {
	// We generate on subcommand so we have both subcommands and parents.
	buf := new(bytes.Buffer)
	if err := GenMarkdown(echoCmd, buf); err != nil {
		t.Fatal(err)
	}
	output := buf.String()

	checkStringContains(t, output, echoCmd.Long)
	checkStringContains(t, output, echoCmd.Example)
	checkStringContains(t, output, "boolone")
	checkStringContains(t, output, "rootflag")
	checkStringContains(t, output, rootCmd.Short)
	checkStringContains(t, output, echoSubCmd.Short)
	checkStringOmits(t, output, deprecatedCmd.Short)
	checkStringContains(t, output, "Options inherited from parent commands")
}

func TestGenMdDocWithNoLongOrSynopsis(t *testing.T) {
	// We generate on subcommand so we have both subcommands and parents.
	buf := new(bytes.Buffer)
	if err := GenMarkdown(dummyCmd, buf); err != nil {
		t.Fatal(err)
	}
	output := buf.String()

	checkStringContains(t, output, dummyCmd.Example)
	checkStringContains(t, output, dummyCmd.Short)
	checkStringContains(t, output, "Options inherited from parent commands")
	checkStringOmits(t, output, "### Synopsis")
}

func TestGenMdNoHiddenParents(t *testing.T) {
	// We generate on subcommand so we have both subcommands and parents.
	for _, name := range []string{"rootflag", "strtwo"} {
		f := rootCmd.PersistentFlags().Lookup(name)
		f.Hidden = true
		defer func() { f.Hidden = false }()
	}
	buf := new(bytes.Buffer)
	if err := GenMarkdown(echoCmd, buf); err != nil {
		t.Fatal(err)
	}
	output := buf.String()

	checkStringContains(t, output, echoCmd.Long)
	checkStringContains(t, output, echoCmd.Example)
	checkStringContains(t, output, "boolone")
	checkStringOmits(t, output, "rootflag")
	checkStringContains(t, output, rootCmd.Short)
	checkStringContains(t, output, echoSubCmd.Short)
	checkStringOmits(t, output, deprecatedCmd.Short)
	checkStringOmits(t, output, "Options inherited from parent commands")
}

func TestGenMdNoTag(t *testing.T) {
	rootCmd.DisableAutoGenTag = true
	defer func() { rootCmd.DisableAutoGenTag = false }()

	buf := new(bytes.Buffer)
	if err := GenMarkdown(rootCmd, buf); err != nil {
		t.Fatal(err)
	}
	output := buf.String()

	checkStringOmits(t, output, "Auto generated")
}

func TestGenMdTree(t *testing.T) {
	c := &cobra.Command{Use: "do [OPTIONS] arg1 arg2"}
	tmpdir, err := os.MkdirTemp("", "test-gen-md-tree")
	if err != nil {
		t.Fatalf("Failed to create tmpdir: %v", err)
	}
	defer os.RemoveAll(tmpdir)

	if err := GenMarkdownTree(c, tmpdir); err != nil {
		t.Fatalf("GenMarkdownTree failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(tmpdir, "do.md")); err != nil {
		t.Fatalf("Expected file 'do.md' to exist")
	}
}

const wrapTestCmdUse = "wraptest"

func TestGenMdDocWithOptions_NoWrap(t *testing.T) {
	long := "This is a very long description that exceeds a typical wrap width and should not be wrapped when WrapWidth is zero"
	cmd := &cobra.Command{
		Use:   wrapTestCmdUse,
		Short: long,
		Long:  long,
		Run:   emptyRun,
	}

	buf := new(bytes.Buffer)
	if err := GenMarkdownCustomWithOptions(cmd, buf, func(s string) string { return s }, MarkdownOptions{}); err != nil {
		t.Fatal(err)
	}
	output := buf.String()
	checkStringContains(t, output, long)
}

func TestGenMdDocWithOptions_WrapWidth(t *testing.T) {
	long := "This is a very long description that exceeds the wrap width and should be wrapped across multiple lines"
	cmd := &cobra.Command{
		Use:   wrapTestCmdUse,
		Short: long,
		Long:  long,
		Run:   emptyRun,
	}

	buf := new(bytes.Buffer)
	if err := GenMarkdownCustomWithOptions(cmd, buf, func(s string) string { return s }, MarkdownOptions{WrapWidth: 40}); err != nil {
		t.Fatal(err)
	}
	output := buf.String()

	for _, line := range strings.Split(output, "\n") {
		if len(line) > 40 {
			// Lines from flags and other sections may exceed width; only check prose lines.
			if !strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "*") && !strings.HasPrefix(line, "  ") {
				t.Errorf("line exceeds wrap width of 40: %q", line)
			}
		}
	}
}

func TestGenMdDocWithOptions_PreservesExistingNewlines(t *testing.T) {
	long := "First sentence.\nSecond sentence on its own line."
	cmd := &cobra.Command{
		Use:  "test",
		Long: long,
		Run:  emptyRun,
	}

	buf := new(bytes.Buffer)
	if err := GenMarkdownCustomWithOptions(cmd, buf, func(s string) string { return s }, MarkdownOptions{WrapWidth: 80}); err != nil {
		t.Fatal(err)
	}
	output := buf.String()
	checkStringContains(t, output, "First sentence.")
	checkStringContains(t, output, "Second sentence on its own line.")
}

func TestGenMdDocWithOptions_MatchesGenMarkdownCustom(t *testing.T) {
	buf1 := new(bytes.Buffer)
	buf2 := new(bytes.Buffer)
	linkHandler := func(s string) string { return s }

	if err := GenMarkdownCustom(echoCmd, buf1, linkHandler); err != nil {
		t.Fatal(err)
	}
	if err := GenMarkdownCustomWithOptions(echoCmd, buf2, linkHandler, MarkdownOptions{}); err != nil {
		t.Fatal(err)
	}
	if buf1.String() != buf2.String() {
		t.Error("GenMarkdownCustomWithOptions with zero options should produce identical output to GenMarkdownCustom")
	}
}

func TestGenMdTreeCustomWithOptions(t *testing.T) {
	c := &cobra.Command{Use: "do [OPTIONS] arg1 arg2"}
	tmpdir, err := os.MkdirTemp("", "test-gen-md-tree-opts")
	if err != nil {
		t.Fatalf("Failed to create tmpdir: %v", err)
	}
	defer os.RemoveAll(tmpdir)

	identity := func(s string) string { return s }
	emptyStr := func(s string) string { return "" }
	if err := GenMarkdownTreeCustomWithOptions(c, tmpdir, emptyStr, identity, MarkdownOptions{WrapWidth: 80}); err != nil {
		t.Fatalf("GenMarkdownTreeCustomWithOptions failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmpdir, "do.md")); err != nil {
		t.Fatalf("Expected file 'do.md' to exist")
	}
}

func BenchmarkGenMarkdownToFile(b *testing.B) {
	file, err := os.CreateTemp("", "")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(file.Name())
	defer file.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := GenMarkdown(rootCmd, file); err != nil {
			b.Fatal(err)
		}
	}
}
