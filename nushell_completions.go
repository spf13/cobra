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
	"io"
	"os"
)

func (c *Command) GenNushellCompletion(w io.Writer) error {
	buf := new(bytes.Buffer)
	WriteStringAndCheck(buf, `
# An external configurator that works with any cobra based
# command line application (e.g. kubectl, minikube)
let cobra_configurator = {|spans| 
 
  let cmd = $spans.0

  # skip the first entry in the span (the command) and join the rest of the span to create __complete args
  let cmd_args = ($spans | skip 1 | str join ' ') 

  # If the last span entry was empty add "" to the end of the command args
  let cmd_args = if ($spans | last | str trim | is-empty) {
    $'($cmd_args) ""'
  } else {
    $cmd_args
  }

  # The full command to be executed
  let full_cmd = $'($cmd) __complete ($cmd_args)'

  # Since nushell doesn't have anything like eval, execute in a subshell
  let result = (do -i { nu -c $"'($full_cmd)'" } | complete)

  # Create a record with all completion related info. 
  # directive and directive_str are for posterity
  let stdout_lines = ($result.stdout | lines)
  let $completions = ($stdout_lines | drop | parse -r '([\w\-\.:\+]*)\t?(.*)' | rename value description)

  let result = ({
    completions: $completions
    directive_str: ($result.stderr)
    directive: ($stdout_lines | last)
  })

  $result.completions
}`)

	_, err := buf.WriteTo(w)
	return err
}

func (c *Command) GenNushellCompletionFile(filename string) error {
	outFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer outFile.Close()

	return c.GenNushellCompletion(outFile)
}
