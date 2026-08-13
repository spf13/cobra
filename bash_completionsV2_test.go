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

func TestBashCompletionV2QuotesCommonPrefix(t *testing.T) {
	bash, err := exec.LookPath("bash")
	if err != nil {
		t.Skip("bash is not available")
	}

	c := &Command{Use: "root", Run: emptyRun}
	buf := new(bytes.Buffer)
	assertNoErr(t, c.GenBashCompletionV2(buf, true))

	script := buf.String() + `
type() {
    if [[ $1 == -t && $2 == compopt ]]; then
        printf builtin
    else
        builtin type "$@"
    fi
}
compopt() { printf 'compopt %s\n' "$*"; }

COMP_TYPE=9
COMPREPLY=('Foo Bar' 'Foo Baz')
__root_quote_common_prefix

COMPREPLY=($'Foo Bar\tfirst' $'Foo Baz\tsecond')
__root_quote_common_prefix

COMPREPLY=('Foo>Bar' 'Foo>Baz')
__root_quote_common_prefix

COMPREPLY=('Foo' 'Foobar')
__root_quote_common_prefix

COMPREPLY=('Foo Bar' 'Quux Baz')
__root_quote_common_prefix

COMP_TYPE=63
COMPREPLY=('Foo Bar' 'Foo Baz')
__root_quote_common_prefix
`

	cmd := exec.Command(bash)
	cmd.Stdin = strings.NewReader(script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("running generated Bash completion: %v\n%s", err, output)
	}

	const expected = "compopt -o filenames\ncompopt -o filenames\ncompopt -o filenames\n"
	if string(output) != expected {
		t.Fatalf("unexpected compopt calls:\n%s", output)
	}
}
