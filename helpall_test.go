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
	"strings"
	"testing"
	"time"
)

// buildHelpAllTestTree returns a command tree exercising collectCommands edge
// cases: runnable commands at varying depths, a non-runnable parent, a hidden
// command with a visible child underneath, and a deprecated command.
func buildHelpAllTestTree() *Command {
	root := &Command{Use: "tool"}

	search := &Command{
		Use:   "search <keyword>",
		Short: "find public projects",
		RunE:  func(*Command, []string) error { return nil },
	}

	issues := &Command{
		Use:   "issues <project>",
		Short: "list issues",
		RunE:  func(*Command, []string) error { return nil },
	}
	var labelsFlag string
	issues.Flags().StringVar(&labelsFlag, "labels", "", "Filter by labels")

	file := &Command{
		Use:   "file <project> <path>",
		Short: "read file contents",
		RunE:  func(*Command, []string) error { return nil },
	}
	var debugFlag bool
	file.Flags().BoolVar(&debugFlag, "debug", false, "Debug output")
	file.Flags().MarkHidden("debug")

	parent := &Command{
		Use:   "parent",
		Short: "parent container",
	}

	notes := &Command{
		Use:   "notes <project> <iid>",
		Short: "list comments",
		RunE:  func(*Command, []string) error { return nil },
	}
	var systemFlag bool
	notes.Flags().BoolVar(&systemFlag, "system", false, "Include system notes")
	parent.AddCommand(notes)

	hidden := &Command{
		Use:    "hidden-cmd",
		Short:  "should not appear",
		Hidden: true,
		RunE:   func(*Command, []string) error { return nil },
	}
	// Visible child under hidden parent -- should be pruned with the subtree.
	hiddenChild := &Command{
		Use:   "nested",
		Short: "nested under hidden",
		RunE:  func(*Command, []string) error { return nil },
	}
	hidden.AddCommand(hiddenChild)

	deprecated := &Command{
		Use:        "old-cmd",
		Short:      "deprecated command",
		Deprecated: "use new-cmd instead",
		RunE:       func(*Command, []string) error { return nil },
	}

	version := &Command{
		Use:   "version",
		Short: "print version",
		Run:   func(*Command, []string) {},
	}

	root.AddCommand(search, issues, file, parent, hidden, deprecated, version)
	return root
}

func TestWalkCommandsSkipsHidden(t *testing.T) {
	root := buildHelpAllTestTree()
	cmds := collectCommands(root)

	names := make(map[string]bool)
	for _, c := range cmds {
		names[c.Path] = true
	}

	if names["tool hidden-cmd"] {
		t.Error("hidden command should be skipped")
	}
	if names["tool parent"] {
		t.Error("container-only parent should not emit a line")
	}
}

func TestWalkCommandsSkipsHiddenSubtree(t *testing.T) {
	root := buildHelpAllTestTree()
	cmds := collectCommands(root)

	for _, c := range cmds {
		if c.Path == "tool hidden-cmd nested" {
			t.Error("visible child under hidden parent should be pruned with subtree")
		}
	}
}

func TestWalkCommandsSkipsHiddenFlags(t *testing.T) {
	root := buildHelpAllTestTree()
	cmds := collectCommands(root)

	for _, c := range cmds {
		if c.Path == "tool file" {
			for _, f := range c.Flags {
				if strings.Contains(f, "debug") {
					t.Error("hidden flag --debug should not appear")
				}
			}
		}
	}
}

func TestWalkCommandsIncludesRunnableCommands(t *testing.T) {
	root := buildHelpAllTestTree()
	cmds := collectCommands(root)

	paths := make(map[string]bool)
	for _, c := range cmds {
		paths[c.Path] = true
	}

	for _, want := range []string{"tool search", "tool issues", "tool file", "tool parent notes", "tool version"} {
		if !paths[want] {
			t.Errorf("expected command %q in output", want)
		}
	}
}

func TestWalkCommandsIncludesDeprecated(t *testing.T) {
	root := buildHelpAllTestTree()
	cmds := collectCommands(root)

	found := false
	for _, c := range cmds {
		if c.Path == "tool old-cmd" {
			found = true
		}
	}
	if !found {
		t.Error("deprecated command should be included in walk")
	}
}

func TestWalkCommandsExtractsArgs(t *testing.T) {
	root := buildHelpAllTestTree()
	cmds := collectCommands(root)

	for _, c := range cmds {
		switch c.Path {
		case "tool search":
			if c.Args != "<keyword>" {
				t.Errorf("search args = %q, want %q", c.Args, "<keyword>")
			}
		case "tool file":
			if c.Args != "<project> <path>" {
				t.Errorf("file args = %q, want %q", c.Args, "<project> <path>")
			}
		case "tool parent notes":
			if c.Args != "<project> <iid>" {
				t.Errorf("notes args = %q, want %q", c.Args, "<project> <iid>")
			}
		}
	}
}

func TestWalkCommandsCollectsLocalFlags(t *testing.T) {
	root := buildHelpAllTestTree()
	cmds := collectCommands(root)

	for _, c := range cmds {
		if c.Path == "tool parent notes" {
			if len(c.Flags) != 1 || c.Flags[0] != "[--system]" {
				t.Errorf("notes flags = %v, want [[--system]]", c.Flags)
			}
		}
		if c.Path == "tool issues" {
			found := false
			for _, f := range c.Flags {
				if f == "[--labels LABELS]" {
					found = true
				}
			}
			if !found {
				t.Errorf("issues missing --labels flag, got %v", c.Flags)
			}
		}
	}
}

func TestFormatFlagTypes(t *testing.T) {
	root := &Command{Use: "test"}
	var (
		b bool
		s string
		n int
		d time.Duration
	)
	root.Flags().BoolVar(&b, "verbose", false, "verbose output")
	root.Flags().StringVar(&s, "state", "", "filter state")
	root.Flags().IntVar(&n, "tail", 50, "tail lines")
	root.Flags().DurationVar(&d, "interval", 30*time.Second, "poll interval")

	tests := []struct {
		name string
		want string
	}{
		{"verbose", "[--verbose]"},
		{"state", "[--state STATE]"},
		{"tail", "[--tail N]"},
		{"interval", "[--interval DURATION]"},
	}

	for _, tt := range tests {
		f := root.Flags().Lookup(tt.name)
		got := formatFlag(f)
		if got != tt.want {
			t.Errorf("formatFlag(%s) = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestFormatFlagWithShorthand(t *testing.T) {
	root := &Command{Use: "test"}
	var (
		b bool
		s string
		n int
		d time.Duration
	)
	root.Flags().BoolVarP(&b, "verbose", "v", false, "verbose output")
	root.Flags().StringVarP(&s, "output", "o", "", "output file")
	root.Flags().IntVarP(&n, "count", "c", 0, "count")
	root.Flags().DurationVarP(&d, "timeout", "t", 0, "timeout")

	tests := []struct {
		name string
		want string
	}{
		{"verbose", "[-v, --verbose]"},
		{"output", "[-o, --output OUTPUT]"},
		{"count", "[-c, --count N]"},
		{"timeout", "[-t, --timeout DURATION]"},
	}

	for _, tt := range tests {
		f := root.Flags().Lookup(tt.name)
		got := formatFlag(f)
		if got != tt.want {
			t.Errorf("formatFlag(%s) = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestFormatFlagCount(t *testing.T) {
	root := &Command{Use: "test"}
	root.Flags().CountP("verbose", "v", "verbosity level")

	f := root.Flags().Lookup("verbose")
	got := formatFlag(f)
	if got != "[-v, --verbose]" {
		t.Errorf("formatFlag(count) = %q, want %q", got, "[-v, --verbose]")
	}
}

func TestRenderCommandsAlignedOutput(t *testing.T) {
	cmds := []commandInfo{
		{Path: "tool search", Args: "<keyword>", Short: "find public projects"},
		{Path: "tool issues", Args: "<project>", Flags: []string{"[--state STATE]"}, Short: "list issues"},
		{Path: "tool parent notes", Args: "<project> <iid>", Flags: []string{"[--system]"}, Short: "list comments"},
	}

	// Default (no flags in output)
	got := renderCommands(cmds, false)
	lines := strings.Split(got, "\n")

	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d:\n%s", len(lines), got)
		return
	}

	for i, line := range lines {
		if !strings.HasPrefix(line, "    ") {
			t.Errorf("line %d missing 4-space indent: %q", i, line)
		}
	}

	cols := make([]int, len(lines))
	for i, line := range lines {
		cols[i] = strings.Index(line, "# ")
	}
	for i := 1; i < len(cols); i++ {
		if cols[i] != cols[0] {
			t.Errorf("# column mismatch: line 0 at %d, line %d at %d", cols[0], i, cols[i])
		}
	}

	if strings.Contains(got, "--state") {
		t.Error("default output should not contain flags")
	}

	// Verbose (flags in output)
	gotVerbose := renderCommands(cmds, true)
	if !strings.Contains(gotVerbose, "[--state STATE]") {
		t.Error("verbose output should contain flags")
	}
	if !strings.Contains(gotVerbose, "[--system]") {
		t.Error("verbose output should contain --system flag")
	}

	vLines := strings.Split(gotVerbose, "\n")
	vCols := make([]int, len(vLines))
	for i, line := range vLines {
		vCols[i] = strings.Index(line, "# ")
	}
	for i := 1; i < len(vCols); i++ {
		if vCols[i] != vCols[0] {
			t.Errorf("verbose # column mismatch: line 0 at %d, line %d at %d", vCols[0], i, vCols[i])
		}
	}
}

func TestRenderCommandsEmpty(t *testing.T) {
	got := renderCommands(nil, false)
	if got != "" {
		t.Errorf("expected empty string for nil input, got %q", got)
	}

	got = renderCommands([]commandInfo{}, true)
	if got != "" {
		t.Errorf("expected empty string for empty input, got %q", got)
	}
}

func TestRenderCommandsSingleEntry(t *testing.T) {
	cmds := []commandInfo{
		{Path: "tool version", Short: "print version"},
	}
	got := renderCommands(cmds, false)
	if strings.Contains(got, "\n") {
		t.Error("single entry should have no trailing newline")
	}
	if !strings.Contains(got, "tool version") {
		t.Error("missing command path")
	}
}

func TestCollectCommandsNilRoot(t *testing.T) {
	cmds := collectCommands(nil)
	if cmds != nil {
		t.Errorf("expected nil for nil root, got %v", cmds)
	}
}

func TestExtractArgsFromUse(t *testing.T) {
	tests := []struct {
		use  string
		want string
	}{
		{"search <keyword>", "<keyword>"},
		{"file <project> <path>", "<project> <path>"},
		{"version", ""},
		{"list [pattern]", "[pattern]"},
	}

	for _, tt := range tests {
		got := extractArgs(tt.use)
		if got != tt.want {
			t.Errorf("extractArgs(%q) = %q, want %q", tt.use, got, tt.want)
		}
	}
}

func TestHelpAllCommandOutput(t *testing.T) {
	root := buildHelpAllTestTree()
	root.AddCommand(NewHelpAllCommand())

	out, err := executeCommand(root, "help-all")
	if err != nil {
		t.Errorf("help-all failed: %v", err)
		return
	}

	if !strings.Contains(out, "tool search <keyword>") {
		t.Error("missing search command")
	}
	if !strings.Contains(out, "tool parent notes") {
		t.Error("missing parent notes command")
	}
	if strings.Contains(out, "hidden-cmd") {
		t.Error("should not contain hidden command")
	}
	if !strings.Contains(out, "tool help-all") {
		t.Error("help-all should list itself (it is a visible, runnable command)")
	}
	if strings.Contains(out, "[--labels") {
		t.Error("default output should not show flags")
	}
}

func TestHelpAllCommandVerbose(t *testing.T) {
	root := buildHelpAllTestTree()
	root.AddCommand(NewHelpAllCommand())

	out, err := executeCommand(root, "help-all", "--verbose")
	if err != nil {
		t.Errorf("help-all --verbose failed: %v", err)
		return
	}

	if !strings.Contains(out, "[--labels LABELS]") {
		t.Error("verbose output should include --labels flag")
	}
	if !strings.Contains(out, "[--system]") {
		t.Error("verbose output should include --system flag")
	}
}

func TestHelpAllCommandRejectsArgs(t *testing.T) {
	root := buildHelpAllTestTree()
	root.AddCommand(NewHelpAllCommand())

	_, err := executeCommand(root, "help-all", "bogus")
	if err == nil {
		t.Error("help-all should reject positional arguments")
	}
}

func TestHelpAllCommandEmptyTree(t *testing.T) {
	// Suppress cobra's auto-added completion and help commands so the only
	// runnable command in the tree is help-all itself.
	root := &Command{Use: "empty"}
	root.CompletionOptions.DisableDefaultCmd = true
	root.SetHelpCommand(&Command{Use: "no-help", Hidden: true})
	root.AddCommand(NewHelpAllCommand())

	out, err := executeCommand(root, "help-all")
	if err != nil {
		t.Errorf("help-all on empty tree failed: %v", err)
		return
	}

	if !strings.Contains(out, "empty help-all") {
		t.Error("help-all should list itself even in an otherwise empty tree")
	}
}
