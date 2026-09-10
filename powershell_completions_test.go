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
	"strings"
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

func TestPwshCompletionOutputGuards(t *testing.T) {
	for _, includeDesc := range []bool{false, true} {
		t.Run(fmt.Sprintf("includeDesc=%t", includeDesc), func(t *testing.T) {
			buf := new(bytes.Buffer)
			genPowerShellComp(buf, "c", includeDesc)

			// Check the generated template without requiring PowerShell. Ignore
			// comments and whitespace so wording and indentation can change.
			var fields []string
			for _, line := range strings.Split(buf.String(), "\n") {
				if !strings.HasPrefix(strings.TrimSpace(line), "#") {
					fields = append(fields, strings.Fields(line)...)
				}
			}
			output := strings.Join(fields, " ")

			// These operations must occur in order: normalize scalar output,
			// guard empty output before indexing, return for directive-only
			// output, and remove only the last line (not equal-valued completions).
			for _, want := range []string{
				`[Array]$Directives = $Out`,
				`$Directive = 0 if ($Directives.Count -gt 0) { $DirectiveStr = $Directives[-1].TrimStart(':')`,
				`if ($Directives.Count -le 1) { return }`,
				`$Out = $Directives | Select-Object -First ($Directives.Count - 1)`,
			} {
				index := strings.Index(output, want)
				if index < 0 {
					t.Fatalf("generated completion is missing or misorders %q", want)
				}
				output = output[index+len(want):]
			}
		})
	}
}
