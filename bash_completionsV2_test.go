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
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestBashCompletionV2WithActiveHelp(t *testing.T) {
	c := &Command{Use: "c", Run: emptyRun}

	buf := new(bytes.Buffer)
	assertNoErr(t, c.GenBashCompletionV2(buf, true))
	output := buf.String()

	// check that active help is not being disabled
	activeHelpVar := activeHelpEnvVar(c.Name())
	checkOmit(t, output, fmt.Sprintf("%s=0", activeHelpVar))
}

func TestBashCompletionV2RejoinsColonWordbreak(t *testing.T) {
	bash, err := exec.LookPath("bash")
	if err != nil {
		t.Skip("bash is not available")
	}

	c := &Command{Use: "c", Run: emptyRun}

	buf := new(bytes.Buffer)
	assertNoErr(t, c.GenBashCompletionV2(buf, false))

	tempDir := t.TempDir()
	completionFile := tempDir + "/completion.bash"
	requestFile := tempDir + "/request.log"
	assertNoErr(t, os.WriteFile(completionFile, buf.Bytes(), 0o600))

	script := fmt.Sprintf(`
set -euo pipefail
source "$COMPLETION_FILE"
_init_completion() {
	cur="/"
	prev=":"
	words=(c ls local : /)
	cword=4
}
c() {
	printf '%%s\n' "$@" > "$REQUEST_FILE"
	printf ':%d\n'
}
COMP_TYPE=9
COMP_LINE='c ls local:/'
COMP_POINT=${#COMP_LINE}
__start_c
`, ShellCompDirectiveNoFileComp)

	cmd := exec.Command(bash, "-c", script)
	cmd.Env = append(os.Environ(),
		"COMPLETION_FILE="+completionFile,
		"REQUEST_FILE="+requestFile,
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("bash completion failed: %v\n%s", err, output)
	}

	request, err := os.ReadFile(requestFile)
	assertNoErr(t, err)

	expected := "__completeNoDesc\nls\nlocal:/"
	if got := strings.TrimSpace(string(request)); got != expected {
		t.Fatalf("expected request:\n%s\ngot:\n%s", expected, got)
	}
}
