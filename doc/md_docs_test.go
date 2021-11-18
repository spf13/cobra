package doc

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

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
	tmpdir, err := ioutil.TempDir("", "test-gen-md-tree")
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

func BenchmarkGenMarkdownToFile(b *testing.B) {
	file, err := ioutil.TempFile("", "")
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

func TestGenMdTreeCustom(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "test-gen-md-tree")
	if err != nil {
		t.Fatalf("Failed to create tmpdir: %v", err)
	}
	defer os.RemoveAll(tmpdir)

	prepender := func(s string) string { return "Prepended" }
	identity := func(s string) string { return s }

	if err := GenMarkdownTreeCustom(rootCmd, tmpdir, prepender, identity); err != nil {
		t.Fatalf("GenMarkdownTree failed: %v", err)
	}

	gotRoot := fileContents(t, tmpdir, "root.md")
	checkStringContains(t, gotRoot, "Prepended")
	checkStringContains(t, gotRoot, rootCmd.Long)

	gotEcho := fileContents(t, tmpdir, "root_echo.md")
	checkStringContains(t, gotEcho, "Prepended")
	checkStringContains(t, gotEcho, echoCmd.Long)

	gotEchoSub := fileContents(t, tmpdir, "root_echo_echosub.md")
	checkStringContains(t, gotEchoSub, "Prepended")
	checkStringContains(t, gotEchoSub, echoSubCmd.Long)
}

func TestGenMarkdownTreeCustomHeaderAndFooter(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "test-gen-md-tree")
	if err != nil {
		t.Fatalf("Failed to create tmpdir: %v", err)
	}
	defer os.RemoveAll(tmpdir)

	prependStr := fmt.Sprintf("pre%d", time.Now().UnixNano())
	postpendStr := fmt.Sprintf("post%d", time.Now().UnixNano())

	prepender := func(s string) string { return prependStr }
	identity := func(s string) string { return s }
	postpender := func(s string) string { return postpendStr }

	if err := GenMarkdownTreeCustomHeaderAndFooter(rootCmd, tmpdir, prepender, postpender, identity); err != nil {
		t.Fatalf("GenMarkdownTree failed: %v", err)
	}

	gotRoot := fileContents(t, tmpdir, "root.md")
	checkStringContains(t, gotRoot, prependStr)
	checkStringContains(t, gotRoot, rootCmd.Long)
	checkStringContains(t, gotRoot, postpendStr)

	gotEcho := fileContents(t, tmpdir, "root_echo.md")
	checkStringContains(t, gotEcho, prependStr)
	checkStringContains(t, gotEcho, echoCmd.Long)
	checkStringContains(t, gotEcho, postpendStr)

	gotEchoSub := fileContents(t, tmpdir, "root_echo_echosub.md")
	checkStringContains(t, gotEchoSub, prependStr)
	checkStringContains(t, gotEchoSub, echoSubCmd.Long)
	checkStringContains(t, gotEchoSub, postpendStr)
}
