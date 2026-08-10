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
	"fmt"
	"testing"
)

func TestPwshCompletionNoActiveHelp(t *testing.T) {
	c := &Command{Use: "c", Run: emptyRun}

	buf := new(bytes.Buffer)
	assertNoErr(t, c.GenPowerShellCompletion(buf))
	output := buf.String()

	// check that active help is being disabled
	activeHelpVar := activeHelpEnvVar(c.Name())
	check(t, output, fmt.Sprintf("${env:%s}=0", activeHelpVar))
}

// Regression for https://github.com/spf13/cobra/issues/2481: empty/single-line
// completion output must not call methods on null or System.Char, and stripping
// the directive must not filter by value.
func TestPwshCompletionDirectiveHandling(t *testing.T) {
	c := &Command{Use: "c", Run: emptyRun}

	buf := new(bytes.Buffer)
	assertNoErr(t, c.GenPowerShellCompletion(buf))

	for _, want := range []string{
		"$Out = @($Out)",
		"Select-Object -First ($Out.Count - 1)",
		"([string]$Out[-1]).TrimStart(':')",
	} {
		if !bytes.Contains(buf.Bytes(), []byte(want)) {
			t.Errorf("generated PowerShell completion missing %q", want)
		}
	}
	// Old buggy patterns must not reappear
	for _, bad := range []string{
		"$Out = $Out | Where-Object { $_ -ne $Out[-1] }",
		"[int]$Directive = $Out[-1].TrimStart(':')",
	} {
		if bytes.Contains(buf.Bytes(), []byte(bad)) {
			t.Errorf("generated PowerShell completion still contains buggy pattern %q", bad)
		}
	}
}
