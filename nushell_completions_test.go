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
	"log"
	"os"
	"testing"
)

func TestGenNushellCompletion(t *testing.T) {
	rootCmd := &Command{Use: "kubectl", Run: emptyRun}
	rootCmd.PersistentFlags().String("server", "s", "The address and port of the Kubernetes API server")
	rootCmd.PersistentFlags().BoolP("skip-headers", "", false, "The address and port of the Kubernetes API serverIf true, avoid header prefixes in the log messages")
	getCmd := &Command{
		Use:        "get",
		Short:      "Display one or many resources",
		ArgAliases: []string{"pods", "nodes", "services", "replicationcontrollers", "po", "no", "svc", "rc"},
		ValidArgs:  []string{"pod", "node", "service", "replicationcontroller"},
		Run:        emptyRun,
	}
	rootCmd.AddCommand(getCmd)

	buf := new(bytes.Buffer)
	assertNoErr(t, rootCmd.GenNushellCompletion(buf))
	output := buf.String()

	check(t, output, "let full_cmd = $'($cmd)_ACTIVE_HELP=0 ($cmd) __complete ($cmd_args)'")
}

func TestGenNushellCompletionFile(t *testing.T) {
	err := os.Mkdir("./tmp", 0o755)
	if err != nil {
		log.Fatal(err.Error())
	}

	defer os.RemoveAll("./tmp")

	rootCmd := &Command{Use: "root", Args: NoArgs, Run: emptyRun}
	child := &Command{
		Use:               "child",
		ValidArgsFunction: validArgsFunc,
		Run:               emptyRun,
	}
	rootCmd.AddCommand(child)

	assertNoErr(t, rootCmd.GenNushellCompletionFile("./tmp/test"))
}

func TestFailGenNushellCompletionFile(t *testing.T) {
	err := os.Mkdir("./tmp", 0o755)
	if err != nil {
		log.Fatal(err.Error())
	}

	defer os.RemoveAll("./tmp")

	f, _ := os.OpenFile("./tmp/test", os.O_CREATE, 0o400)
	defer f.Close()

	rootCmd := &Command{Use: "root", Args: NoArgs, Run: emptyRun}
	child := &Command{
		Use:               "child",
		ValidArgsFunction: validArgsFunc,
		Run:               emptyRun,
	}
	rootCmd.AddCommand(child)

	got := rootCmd.GenNushellCompletionFile("./tmp/test")
	if got == nil {
		t.Error("should raise permission denied error")
	}

	if os.Getenv("MSYSTEM") == "MINGW64" {
		if got.Error() != "open ./tmp/test: Access is denied." {
			t.Errorf("got: %s, want: %s", got.Error(), "open ./tmp/test: Access is denied.")
		}
	} else {
		if got.Error() != "open ./tmp/test: permission denied" {
			t.Errorf("got: %s, want: %s", got.Error(), "open ./tmp/test: permission denied")
		}
	}
}
