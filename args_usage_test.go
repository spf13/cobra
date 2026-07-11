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
	"bytes"
	"strings"
	"testing"
)

func TestHasAvailableArguments(t *testing.T) {
	none := &Command{Use: "none", Run: emptyRun}
	if none.HasAvailableArguments() {
		t.Error("expected no documented arguments")
	}

	some := &Command{
		Use:       "some <id>",
		Run:       emptyRun,
		Arguments: []ArgSpec{{Name: "id", Description: "an id"}},
	}
	if !some.HasAvailableArguments() {
		t.Error("expected documented arguments")
	}
}

func TestArgumentUsages(t *testing.T) {
	c := &Command{
		Use: "get <id> <path>",
		Arguments: []ArgSpec{
			{Name: "id", Description: "item id", Example: "10239"},
			{Name: "path", Description: "destination file"},
			{Name: "[note]", Description: "optional note"},
		},
	}

	got := c.ArgumentUsages()
	want := "" +
		"  <id>     item id (e.g. 10239)\n" +
		"  <path>   destination file\n" +
		"  [note]   optional note\n"
	if got != want {
		t.Errorf("ArgumentUsages mismatch:\ngot:\n%q\nwant:\n%q", got, want)
	}
}

func TestArgSpecPlaceholder(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"id", "<id>"},
		{"config.file", "<config.file>"}, // bare dotted name is still wrapped
		{"user.email", "<user.email>"},
		{"[note]", "[note]"},         // authored optional notation kept
		{"<id>", "<id>"},             // already wrapped, kept
		{"{a|b}", "{a|b}"},           // alternation kept
		{"files...", "files..."},     // variadic suffix kept
		{"[files...]", "[files...]"}, // bracketed variadic kept
		{"", ""},                     // degenerate empty name
	}
	for _, tt := range tests {
		if got := (ArgSpec{Name: tt.name}).placeholder(); got != tt.want {
			t.Errorf("placeholder(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestArgumentUsagesEmpty(t *testing.T) {
	c := &Command{Use: "bare"}
	if got := c.ArgumentUsages(); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestArgumentUsagesExampleNotDuplicated(t *testing.T) {
	c := &Command{
		Use: "do",
		Arguments: []ArgSpec{
			{Name: "ref", Description: "submission ref such as DEVOOPS-8298", Example: "DEVOOPS-8298"},
		},
	}
	if n := strings.Count(c.ArgumentUsages(), "DEVOOPS-8298"); n != 1 {
		t.Errorf("expected the example to appear once, got %d", n)
	}
}

func TestArgumentsInHelpOutput(t *testing.T) {
	c := &Command{
		Use:   "get <id> <path>",
		Short: "download an item",
		Run:   emptyRun,
		Arguments: []ArgSpec{
			{Name: "id", Description: "item id", Example: "10239"},
			{Name: "path", Description: "destination file"},
		},
	}

	output, err := executeCommand(c, "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	checkStringContains(t, output, "Arguments:")
	checkStringContains(t, output, "<id>")
	checkStringContains(t, output, "item id (e.g. 10239)")
	checkStringContains(t, output, "<path>")
}

func TestArgumentsOmittedWhenAbsent(t *testing.T) {
	c := &Command{Use: "noargs", Short: "no positionals", Run: emptyRun}

	output, err := executeCommand(c, "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	checkStringOmits(t, output, "Arguments:")
}

// TestArgumentsTemplateMatchesFunc guards the invariant that defaultUsageFunc
// stays equivalent to defaultUsageTemplate for the Arguments block.
func TestArgumentsTemplateMatchesFunc(t *testing.T) {
	mk := func() *Command {
		return &Command{
			Use:   "get <id> <path>",
			Short: "download an item",
			Run:   emptyRun,
			Arguments: []ArgSpec{
				{Name: "id", Description: "item id", Example: "10239"},
				{Name: "path", Description: "destination file"},
			},
		}
	}

	fromFunc := new(bytes.Buffer)
	cf := mk()
	cf.SetOut(fromFunc)
	if err := cf.Usage(); err != nil {
		t.Fatal(err)
	}

	fromTemplate := new(bytes.Buffer)
	ct := mk()
	ct.SetOut(fromTemplate)
	ct.SetUsageTemplate(ct.UsageTemplate()) // force the text-template path
	if err := ct.Usage(); err != nil {
		t.Fatal(err)
	}

	if fromFunc.String() != fromTemplate.String() {
		t.Errorf("usage func and template differ:\nfunc:\n%s\ntemplate:\n%s", fromFunc.String(), fromTemplate.String())
	}
}
