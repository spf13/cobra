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
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCompleteNoDesCmdInFishScript(t *testing.T) {
	rootCmd := &Command{Use: "root", Args: NoArgs, Run: emptyRun}
	child := &Command{
		Use:               "child",
		ValidArgsFunction: validArgsFunc,
		Run:               emptyRun,
	}
	rootCmd.AddCommand(child)

	buf := new(bytes.Buffer)
	assertNoErr(t, rootCmd.GenFishCompletion(buf, false))
	output := buf.String()

	check(t, output, ShellCompNoDescRequestCmd)
}

func TestCompleteCmdInFishScript(t *testing.T) {
	rootCmd := &Command{Use: "root", Args: NoArgs, Run: emptyRun}
	child := &Command{
		Use:               "child",
		ValidArgsFunction: validArgsFunc,
		Run:               emptyRun,
	}
	rootCmd.AddCommand(child)

	buf := new(bytes.Buffer)
	assertNoErr(t, rootCmd.GenFishCompletion(buf, true))
	output := buf.String()

	check(t, output, ShellCompRequestCmd)
	checkOmit(t, output, ShellCompNoDescRequestCmd)
}

func TestProgWithDash(t *testing.T) {
	rootCmd := &Command{Use: "root-dash", Args: NoArgs, Run: emptyRun}
	buf := new(bytes.Buffer)
	assertNoErr(t, rootCmd.GenFishCompletion(buf, false))
	output := buf.String()

	// Functions name should have replace the '-'
	check(t, output, "__root_dash_perform_completion")
	checkOmit(t, output, "__root-dash_perform_completion")

	// The command name should not have replaced the '-'
	check(t, output, "-c root-dash")
	checkOmit(t, output, "-c root_dash")
}

func TestProgWithColon(t *testing.T) {
	rootCmd := &Command{Use: "root:colon", Args: NoArgs, Run: emptyRun}
	buf := new(bytes.Buffer)
	assertNoErr(t, rootCmd.GenFishCompletion(buf, false))
	output := buf.String()

	// Functions name should have replace the ':'
	check(t, output, "__root_colon_perform_completion")
	checkOmit(t, output, "__root:colon_perform_completion")

	// The command name should not have replaced the ':'
	check(t, output, "-c root:colon")
	checkOmit(t, output, "-c root_colon")
}

func TestFishCompletionNoActiveHelp(t *testing.T) {
	c := &Command{Use: "c", Run: emptyRun}

	buf := new(bytes.Buffer)
	assertNoErr(t, c.GenFishCompletion(buf, true))
	output := buf.String()

	// check that active help is being disabled
	activeHelpVar := activeHelpEnvVar(c.Name())
	check(t, output, fmt.Sprintf("%s=0", activeHelpVar))
}

func TestFishCompletionFlagPrefix(t *testing.T) {
	fish, err := exec.LookPath("fish")
	if err != nil {
		t.Skip("fish is required to execute the generated completion script")
	}

	rootCmd := &Command{Use: "cobra-fish-test", Run: emptyRun}
	buf := new(bytes.Buffer)
	assertNoErr(t, rootCmd.GenFishCompletion(buf, false))

	// Keep the completion backend fixed so these tests exercise the generated
	// shell adapter, including Fish's handling of flag and value boundaries.
	script := `
function cobra-fish-test
    if test "$argv[1]" = "__completeNoDesc"
        printf '%s\n' "$COBRA_FISH_TEST_CANDIDATE" ':4'
    end
end
` + buf.String() + `
complete --do-complete "$COBRA_FISH_TEST_LINE"
`
	tests := []struct {
		name, arguments, candidate, expected string
	}{
		{"long flag", "--selector=w", "web", "--selector=web"},
		{"short flag", "-s=w", "web", "-s=web"},
		{"long flag empty value", "--selector=", "web", "--selector=web"},
		{"short flag empty value", "-s=", "web", "-s=web"},
		{"long flag equals in value", "--selector=app=w", "app=web", "--selector=app=web"},
		{"short flag equals in value", "-s=app=w", "app=web", "-s=app=web"},
		{"multiple equals in value", "--selector=app=x=w", "app=x=web", "--selector=app=x=web"},
		{"not equals in value", "--selector=app!=w", "app!=web", "--selector=app!=web"},
		{"double equals in value", "--selector=app==w", "app==web", "--selector=app==web"},
		{"comma separated value", "--selector=env=prod,app=w", "env=prod,app=web", "--selector=env=prod,app=web"},
		{"hyphen in flag", "--label-selector=app=w", "app=web", "--label-selector=app=web"},
		{"hyphen and equals in value", "--selector=app-name=w", "app-name=web", "--selector=app-name=web"},
		{"leading hyphen in value", "--selector=-app=w", "-app=web", "--selector=-app=web"},
		{"separate long flag value", "--selector app=w", "app=web", "app=web"},
		{"separate short flag value", "-s app=w", "app=web", "app=web"},
		{"separate long flag hyphenated value", "--selector app-name=w", "app-name=web", "app-name=web"},
		{"separate short flag hyphenated value", "-s app-name=w", "app-name=web", "app-name=web"},
		{"positional hyphenated value", "app-name=w", "app-name=web", "app-name=web"},
		{"positional equals value", "app=w", "app=web", "app=web"},
		{"positional plain value", "w", "web", "web"},
		{"open double quoted value", `--selector="w`, "web", "--selector=web"},
		{"closed double quoted value", `--selector="w"`, "web", "--selector=web"},
		{"open double quoted token", `"--selector=w`, "web", "--selector=web"},
		{"closed double quoted token", `"--selector=w"`, "web", "--selector=web"},
		{"open single quoted value", "--selector='w", "web", "--selector=web"},
		{"quoted value containing equals", `--selector="app=w`, "app=web", "--selector=app=web"},
		{"quoted token containing equals", `"--selector=app=w`, "app=web", "--selector=app=web"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, fish, "--no-config")
			cmd.Stdin = strings.NewReader(script)
			configDir := t.TempDir()
			cmd.Env = append(os.Environ(),
				"HOME="+configDir,
				"XDG_CONFIG_HOME="+configDir,
				"BASH_COMP_DEBUG_FILE=",
				"COBRA_FISH_TEST_LINE=cobra-fish-test "+tc.arguments,
				"COBRA_FISH_TEST_CANDIDATE="+tc.candidate,
			)
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("fish completion failed: %v\n%s", err, output)
			}
			if got := strings.TrimSuffix(string(output), "\n"); got != tc.expected {
				t.Errorf("completion for %q = %q, want %q", tc.arguments, got, tc.expected)
			}
		})
	}
}

func TestGenFishCompletionFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "cobra-test")
	if err != nil {
		t.Fatal(err.Error())
	}

	defer os.Remove(tmpFile.Name())

	rootCmd := &Command{Use: "root", Args: NoArgs, Run: emptyRun}
	child := &Command{
		Use:               "child",
		ValidArgsFunction: validArgsFunc,
		Run:               emptyRun,
	}
	rootCmd.AddCommand(child)

	assertNoErr(t, rootCmd.GenFishCompletionFile(tmpFile.Name(), false))
}

func TestFailGenFishCompletionFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cobra-test")
	if err != nil {
		t.Fatal(err.Error())
	}

	defer os.RemoveAll(tmpDir)

	f, _ := os.OpenFile(filepath.Join(tmpDir, "test"), os.O_CREATE, 0400)
	defer f.Close()

	rootCmd := &Command{Use: "root", Args: NoArgs, Run: emptyRun}
	child := &Command{
		Use:               "child",
		ValidArgsFunction: validArgsFunc,
		Run:               emptyRun,
	}
	rootCmd.AddCommand(child)

	got := rootCmd.GenFishCompletionFile(f.Name(), false)
	if !errors.Is(got, os.ErrPermission) {
		t.Errorf("got: %s, want: %s", got.Error(), os.ErrPermission.Error())
	}
}
