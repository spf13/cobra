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
	"os/exec"
	"testing"
)

func TestZshCompletionWithActiveHelp(t *testing.T) {
	c := &Command{Use: "c", Run: emptyRun}

	buf := new(bytes.Buffer)
	assertNoErr(t, c.GenZshCompletion(buf))
	output := buf.String()

	// check that active help is not being disabled
	activeHelpVar := activeHelpEnvVar(c.Name())
	checkOmit(t, output, fmt.Sprintf("%s=0", activeHelpVar))
}

func TestZshCompletionFlagPrefix(t *testing.T) {
	zsh, err := exec.LookPath("zsh")
	if err != nil {
		t.Skip("zsh is not installed")
	}

	c := &Command{Use: "c", Run: emptyRun}
	buf := new(bytes.Buffer)
	assertNoErr(t, c.GenZshCompletion(buf))

	// Capture the prefix passed to zsh's completion system by the generated script.
	script := "compdef() {}\n" + buf.String() + `
c() { printf '%s\n' 'key=value' ':4' }
_describe() {
    while (( $# )); do
        if [[ "$1" == -P ]]; then
            printf '%s' "$2"
            return
        fi
        shift
    done
}
words=(c "$1")
CURRENT=2
_c
`

	for _, tc := range []struct {
		name       string
		word       string
		wantPrefix string
	}{
		{"long flag", "--output=value", "--output="},
		{"short flag", "-o=value", "-o="},
		{"empty value", "--set=", "--set="},
		{"key value", "--set=key=value", "--set="},
		{"combined key values", "--set=k1=v1,k2=v", "--set="},
		{"short key values", "-s=k1=v1,k2=v", "-s="},
		{"flag without equals", "--output", ""},
		{"separate key value", "key=value", ""},
		{"hyphen in value", "key-name=value", ""},
		{"escaped flag", `\--output=value`, "--output="},
		{"double quoted flag", `"--output=value"`, "--output="},
		{"single quoted flag", "'--output=value'", "--output="},
		{"quoted key values", `"--set=k1=v1,k2=v"`, "--set="},
		{"quoted hyphen in value", `"key-name=value"`, ""},
		{"quoted flag fragment", "--out'put'=value", "--output="},
	} {
		t.Run(tc.name, func(t *testing.T) {
			output, err := exec.Command(zsh, "-f", "-c", script, "--", tc.word).CombinedOutput()
			assertNoErr(t, err)
			if string(output) != tc.wantPrefix {
				t.Errorf("completion prefix for %q: got %q, want %q", tc.word, output, tc.wantPrefix)
			}
		})
	}
}
