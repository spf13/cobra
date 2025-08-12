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

func TestZshCompletionWithActiveHelp(t *testing.T) {
	c := &Command{Use: "c", Run: emptyRun}

	buf := new(bytes.Buffer)
	assertNoErr(t, c.GenZshCompletion(buf))
	output := buf.String()

	// check that active help is not being disabled
	activeHelpVar := activeHelpEnvVar(c.Name())
	checkOmit(t, output, fmt.Sprintf("%s=0", activeHelpVar))
}

func TestZshCompletionWithInfluencedPermission(t *testing.T) {
	c := &Command{Use: "c", Run: emptyRun, InfluencedByPermissions: true}

	buf := new(bytes.Buffer)
	assertNoErr(t, c.GenZshCompletion(buf))
	output := buf.String()

	// check that related commands are being generated
	check(t, output, fmt.Sprintf(ZstyleGainPrivileges, c.Name()))
	check(t, output, fmt.Sprintf("_call_program -p %s-tag", c.Name()))
}
