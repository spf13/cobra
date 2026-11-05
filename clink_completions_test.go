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
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

type clinkFailWriter struct{}

func (clinkFailWriter) Write([]byte) (int, error) { return 0, errors.New("write failed") }

func TestClinkGeneration(t *testing.T) {
	for _, desc := range []bool{false, true} {
		c := &Command{Use: "test-cli"}
		var b bytes.Buffer
		assertNoErr(t, c.GenClinkCompletion(&b, desc))
		request := ShellCompRequestCmd
		if !desc {
			request = ShellCompNoDescRequestCmd
		}
		check(t, b.String(), `local request_command = "`+request+`"`)
		check(t, b.String(), `cobra.register("test-cli")`)
		for _, unwanted := range []string{"@NAME@", "@REQUEST@", "require("} {
			checkOmit(t, b.String(), unwanted)
		}
		file := filepath.Join(t.TempDir(), "test-cli.lua")
		assertNoErr(t, c.GenClinkCompletionFile(file, desc))
		contents, err := os.ReadFile(file)
		assertNoErr(t, err)
		if string(contents) != b.String() {
			t.Fatal("file and writer output differ")
		}
	}
	c := &Command{Use: "app"}
	if c.GenClinkCompletion(clinkFailWriter{}, true) == nil {
		t.Fatal("lost writer error")
	}
	if c.GenClinkCompletionFile(filepath.Join(t.TempDir(), "missing", "x.lua"), true) == nil {
		t.Fatal("lost file creation error")
	}
}

func TestClinkDefaultCommand(t *testing.T) {
	for _, disable := range []bool{false, true} {
		for _, flag := range []bool{false, true} {
			if disable && flag {
				continue
			}
			c := &Command{Use: "app", CompletionOptions: CompletionOptions{DisableDescriptions: disable}}
			args := []string{"completion", "clink"}
			if flag {
				args = append(args, "--no-descriptions")
			}
			output, err := executeCommand(c, args...)
			assertNoErr(t, err)
			request := ShellCompRequestCmd
			if disable || flag {
				request = ShellCompNoDescRequestCmd
			}
			check(t, output, `local request_command = "`+request+`"`)
		}
	}
	c := &Command{Use: "app"}
	_, err := executeCommand(c, "completion", "clink", "unexpected")
	if err == nil {
		t.Fatal("accepted unexpected argument")
	}
}

func TestClinkLuaQuoting(t *testing.T) {
	if got := quoteClinkLua("a\"\\\n\x00é"); got != `"a\"\\\010\000\195\169"` {
		t.Fatalf("unsafe Lua string: %s", got)
	}
	c := &Command{Use: "APP.EXE"}
	var b bytes.Buffer
	assertNoErr(t, c.GenClinkCompletion(&b, true))
	check(t, b.String(), `cobra.register("app")`)
}

// Set COBRA_CLINK to clink_x64.exe for the Windows integration suite.
// Without Clink, a Lua 5.2+ interpreter runs the deterministic protocol suite.
func TestClinkLuaRuntime(t *testing.T) {
	clink := os.Getenv("COBRA_CLINK")
	interpreter := clink
	if interpreter == "" {
		for _, name := range []string{"lua5.2", "lua5.3", "lua5.4", "lua"} {
			if path, err := exec.LookPath(name); err == nil {
				interpreter = path
				break
			}
		}
	}
	if interpreter == "" {
		t.Skip("install Lua 5.2+ or set COBRA_CLINK to run the generated Lua tests")
	}
	dir := t.TempDir()
	script := filepath.Join(dir, "cobrafixture.lua")
	c := &Command{Use: "cobrafixture"}
	assertNoErr(t, c.GenClinkCompletionFile(script, true))
	fixture := os.Getenv("COBRA_CLINK_FIXTURE")
	if clink != "" && runtime.GOOS == "windows" {
		if fixture == "" {
			fixture = filepath.Join(dir, "cobrafixture.exe")
			cmd := exec.Command("go", "build", "-o", fixture, "./testdata/clink")
			if output, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("fixture: %v\n%s", err, output)
			}
		}
	}
	args := []string{"testdata/clink/run.lua", script, fixture, "descriptions"}
	if clink != "" {
		args = append([]string{"lua"}, args...)
	}
	cmd := exec.Command(interpreter, args...) //nolint:gosec // The test runner explicitly selects the interpreter; no shell is used.
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Lua runtime: %v\n%s", err, output)
	}
	if !strings.Contains(string(output), "PASS") {
		t.Fatalf("no test result: %s", output)
	}
	t.Log(string(output))
	assertNoErr(t, c.GenClinkCompletionFile(script, false))
	args[len(args)-1] = "no-descriptions"
	cmd = exec.Command(interpreter, args...) //nolint:gosec // The test runner explicitly selects the interpreter; no shell is used.
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Lua runtime without descriptions: %v\n%s", err, output)
	}
	t.Log(string(output))
}
