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
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
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
	assertNoErr(t, rootCmd.GenNushellCompletion(buf, true))
}

func TestGenNushellCompletionFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "cobra-test")
	if err != nil {
		log.Fatal(err.Error())
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

func TestNushellCompletionNoActiveHelp(t *testing.T) {
	c := &Command{Use: "c", Run: emptyRun}

	buf := new(bytes.Buffer)
	assertNoErr(t, c.GenNushellCompletion(buf, true))
	output := buf.String()

	// check that active help is being disabled
	activeHelpVar := activeHelpGlobalEnvVar
	check(t, output, fmt.Sprintf("%s=0", activeHelpVar))
}
