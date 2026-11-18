// Copyright 2013-2022 The Cobra Authors
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
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// nushellTestProgEnvVar makes the test binary act as the cobra program that
// the nushell integration test completes.
const nushellTestProgEnvVar = "COBRA_NUSHELL_TEST_PROG"

func TestMain(m *testing.M) {
	if os.Getenv(nushellTestProgEnvVar) == "1" {
		if err := nushellTestProg().Execute(); err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func nushellTestProg() *Command {
	root := &Command{Use: "nuprog", Run: emptyRun}
	root.PersistentFlags().String("namespace", "", "the namespace")
	_ = root.RegisterFlagCompletionFunc("namespace", func(*Command, []string, string) ([]string, ShellCompDirective) {
		return []string{"default", "kube-system", "kube-public", "mytest", "test"}, ShellCompDirectiveNoFileComp
	})
	directive := func(comps []string, d ShellCompDirective) CompletionFunc {
		return func(*Command, []string, string) ([]string, ShellCompDirective) { return comps, d }
	}
	root.AddCommand(
		&Command{Use: "get", Short: "Get resources", ValidArgs: []string{"pods\tThe pods", "nodes"}, Run: emptyRun},
		&Command{Use: "nospace", Run: emptyRun, ValidArgsFunction: directive([]string{"bitnami/"}, ShellCompDirectiveNoSpace)},
		&Command{Use: "order", Run: emptyRun, ValidArgsFunction: directive([]string{"zz", "aa", "mm"}, ShellCompDirectiveKeepOrder)},
		&Command{Use: "nofiles", Run: emptyRun, ValidArgsFunction: directive(nil, ShellCompDirectiveNoFileComp)},
		&Command{Use: "files", Run: emptyRun},
		&Command{Use: "ext", Run: emptyRun, ValidArgsFunction: directive([]string{"go"}, ShellCompDirectiveFilterFileExt)},
		&Command{Use: "dirs", Run: emptyRun, ValidArgsFunction: directive(nil, ShellCompDirectiveFilterDirs)},
		&Command{Use: "subdirs", Run: emptyRun, ValidArgsFunction: directive([]string{"d"}, ShellCompDirectiveFilterDirs)},
		&Command{Use: "err", Run: emptyRun, ValidArgsFunction: directive([]string{"x"}, ShellCompDirectiveError)},
		&Command{Use: "help-active", Run: emptyRun, ValidArgsFunction: func(*Command, []string, string) ([]string, ShellCompDirective) {
			return AppendActiveHelp([]string{"value"}, "some help"), ShellCompDirectiveNoFileComp
		}},
	)
	return root
}

func TestGenNushellCompletion(t *testing.T) {
	rootCmd := &Command{Use: "kubectl", Run: emptyRun}
	rootCmd.AddCommand(&Command{Use: "get", Run: emptyRun})

	buf := new(bytes.Buffer)
	assertNoErr(t, rootCmd.GenNushellCompletion(buf, true))
	output := buf.String()

	check(t, output, `def "nu-complete kubectl" [place: record]`)
	check(t, output, `@complete "nu-complete kubectl"`)
	check(t, output, `extern "kubectl" [...args]`)
	check(t, output, fmt.Sprintf(`run-external "kubectl" "%s"`, ShellCompRequestCmd))
}

func TestGenNushellCompletionNoDesc(t *testing.T) {
	rootCmd := &Command{Use: "kubectl", Run: emptyRun}

	buf := new(bytes.Buffer)
	assertNoErr(t, rootCmd.GenNushellCompletion(buf, false))
	output := buf.String()

	check(t, output, fmt.Sprintf(`run-external "kubectl" "%s"`, ShellCompNoDescRequestCmd))
}

func TestGenNushellCompletionSpecialChars(t *testing.T) {
	rootCmd := &Command{Use: "my-prog:v2", Run: emptyRun}

	buf := new(bytes.Buffer)
	assertNoErr(t, rootCmd.GenNushellCompletion(buf, true))
	output := buf.String()

	check(t, output, `extern "my-prog:v2" [...args]`)
	check(t, output, "MY_PROG_V2_ACTIVE_HELP")
}

func TestGenNushellCompletionFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "cobra-test")
	if err != nil {
		t.Fatal(err.Error())
	}

	defer os.RemoveAll(tmpFile.Name())

	rootCmd := &Command{Use: "root", Args: NoArgs, Run: emptyRun}
	child := &Command{
		Use:               "child",
		ValidArgsFunction: validArgsFunc,
		Run:               emptyRun,
	}
	rootCmd.AddCommand(child)

	assertNoErr(t, rootCmd.GenNushellCompletionFile(tmpFile.Name(), true))
}

func TestFailGenNushellCompletionFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cobra-test")
	if err != nil {
		t.Fatal(err.Error())
	}

	defer os.RemoveAll(tmpDir)

	f, _ := os.OpenFile(filepath.Join(tmpDir, "test"), os.O_CREATE, 0o400)
	defer f.Close()

	rootCmd := &Command{Use: "root", Args: NoArgs, Run: emptyRun}
	child := &Command{
		Use:               "child",
		ValidArgsFunction: validArgsFunc,
		Run:               emptyRun,
	}
	rootCmd.AddCommand(child)

	got := rootCmd.GenNushellCompletionFile(f.Name(), true)
	if !errors.Is(got, os.ErrPermission) {
		t.Errorf("got: %s, want: %s", got.Error(), os.ErrPermission.Error())
	}
}

func TestNushellCompletionNoActiveHelp(t *testing.T) {
	c := &Command{Use: "c", Run: emptyRun}

	buf := new(bytes.Buffer)
	assertNoErr(t, c.GenNushellCompletion(buf, true))
	output := buf.String()

	// check that active help is being disabled
	activeHelpVar := activeHelpEnvVar(c.Name())
	check(t, output, fmt.Sprintf(`%s: "0"`, activeHelpVar))
}

// nushellMinVersion is the oldest nushell the generated script supports.
var nushellMinVersion = [2]int{0, 116}

// requireNushell returns the path to a usable nu binary, or skips the test.
func requireNushell(t *testing.T) string {
	t.Helper()
	nu, err := exec.LookPath("nu")
	if err != nil {
		t.Skip("nu not found in PATH")
	}
	out, err := exec.Command(nu, "--version").Output()
	if err != nil {
		t.Skipf("cannot get nu version: %v", err)
	}
	m := regexp.MustCompile(`^(\d+)\.(\d+)`).FindStringSubmatch(strings.TrimSpace(string(out)))
	if m == nil {
		t.Skipf("cannot parse nu version %q", out)
	}
	major, _ := strconv.Atoi(m[1])
	minor, _ := strconv.Atoi(m[2])
	if major < nushellMinVersion[0] || (major == nushellMinVersion[0] && minor < nushellMinVersion[1]) {
		t.Skipf("nu %s is older than %d.%d", m[0], nushellMinVersion[0], nushellMinVersion[1])
	}
	return nu
}

// TestNushellCompletionIntegration loads the generated script into nushell and
// checks the completions that nushell offers for a range of command lines.
func TestNushellCompletionIntegration(t *testing.T) {
	nu := requireNushell(t)

	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	binDir := t.TempDir()
	if err := os.Symlink(exe, filepath.Join(binDir, "nuprog")); err != nil {
		t.Skipf("cannot create symlink: %v", err)
	}

	workDir := t.TempDir()
	assertNoErr(t, os.MkdirAll(filepath.Join(workDir, "d", "sub"), 0o755))
	for _, name := range []string{"a.go", "b.txt", filepath.Join("d", "c.go")} {
		assertNoErr(t, os.WriteFile(filepath.Join(workDir, name), nil, 0o600))
	}

	script := filepath.Join(t.TempDir(), "nuprog.nu")
	assertNoErr(t, nushellTestProg().GenNushellCompletionFile(script, true))

	type suggestion struct {
		Value            string `json:"value"`
		Description      string `json:"description"`
		AppendWhitespace bool   `json:"append_whitespace"`
	}

	testCases := []struct {
		line string
		want []suggestion
	}{
		{"nuprog g", []suggestion{{"get", "Get resources", true}}},
		{"nuprog get ", []suggestion{{"nodes", "", true}, {"pods", "The pods", true}}},
		{"nuprog get p", []suggestion{{"pods", "The pods", true}}},
		{"nuprog --namespace k", []suggestion{{"kube-public", "", true}, {"kube-system", "", true}}},
		// Prefix matches win over substring matches.
		{"nuprog --namespace te", []suggestion{{"test", "", true}}},
		// Only substring matches.
		{"nuprog --namespace ytes", []suggestion{{"mytest", "", true}}},
		{"nuprog --namespace=te", []suggestion{{"--namespace=test", "", true}}},
		{"nuprog get --names", []suggestion{{"--namespace", "the namespace", true}}},
		{"nuprog nospace ", []suggestion{{"bitnami/", "", false}}},
		{"nuprog order ", []suggestion{{"zz", "", true}, {"aa", "", true}, {"mm", "", true}}},
		{"nuprog nofiles ", nil},
		{"nuprog err ", nil},
		{"nuprog help-active ", []suggestion{{"value", "", true}}},
		{"nuprog files ", []suggestion{{"a.go", "", false}, {"b.txt", "", false}, {"d/", "", false}}},
		{"nuprog ext ", []suggestion{{"a.go", "", false}, {"d/", "", false}}},
		{"nuprog dirs ", []suggestion{{"d/", "", false}}},
		{"nuprog subdirs ", []suggestion{{"sub/", "", false}}},
		// Other commands keep nushell's own completion.
		{"cat d/", []suggestion{{"d/c.go", "", false}, {"d/sub/", "", false}}},
	}

	for _, tc := range testCases {
		t.Run(tc.line, func(t *testing.T) {
			//nolint:gosec // G204: nu is found by the test and the arguments are test data.
			cmd := exec.Command(nu, "--no-config-file", "--commands",
				fmt.Sprintf("source %q; %q | commandline complete --detailed | to json --raw", script, tc.line))
			cmd.Dir = workDir
			cmd.Env = append(os.Environ(),
				"PATH="+binDir+string(os.PathListSeparator)+os.Getenv("PATH"),
				nushellTestProgEnvVar+"=1")
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("nu failed: %v\n%s", err, out)
			}
			var got []suggestion
			if err := json.Unmarshal(out, &got); err != nil {
				t.Fatalf("cannot parse nu output %q: %v", out, err)
			}
			if len(got) == 0 && len(tc.want) == 0 {
				return
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("completions for %q:\n got: %+v\nwant: %+v", tc.line, got, tc.want)
			}
		})
	}
}
