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

func TestBashCompletionV2WithActiveHelp(t *testing.T) {
	c := &Command{Use: "c", Run: emptyRun}

	buf := new(bytes.Buffer)
	assertNoErr(t, c.GenBashCompletionV2(buf, true))
	output := buf.String()

	// check that active help is not being disabled
	activeHelpVar := activeHelpEnvVar(c.Name())
	checkOmit(t, output, fmt.Sprintf("%s=0", activeHelpVar))
}

func TestBashCompletionV2QuotesArgs(t *testing.T) {
	c := &Command{Use: "c", Run: emptyRun}

	buf := new(bytes.Buffer)
	assertNoErr(t, c.GenBashCompletionV2(buf, true))
	output := buf.String()

	// The completion request must quote each argument so that arguments
	// containing spaces or glob characters reach the program verbatim
	// instead of being re-split or expanded by eval.
	check(t, output, `printf -v quotedArg "%q" "$arg"`)
	checkOmit(t, output, "${args[*]}")
	// The empty word being completed is already passed as an argument by
	// the quoting loop, so no extra empty parameter may be appended.
	checkOmit(t, output, "Adding extra empty parameter")
}
