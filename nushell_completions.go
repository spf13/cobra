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
    "fmt"
    "io"
    "os"
)

func (c *Command) GenNushellCompletion(w io.Writer, includeDesc bool) error {
    buf := new(bytes.Buffer)
    WriteStringAndCheck(buf, fmt.Sprintf(`
let cobra_completer = {|spans|
    let ShellCompDirectiveError = %[1]d
    let ShellCompDirectiveNoSpace = %[2]d
    let ShellCompDirectiveNoFileComp = %[3]d
    let ShellCompDirectiveFilterFileExt = %[4]d
    let ShellCompDirectiveFilterDirs = %[5]d

    let cmd = $spans | first 
    let rest = $spans | skip

    def exec_complete [
        spans: list<string>
    ] {
        # This will catch the stderr message related to the directive and any other errors,
        # such as the command not being a cobra based command
        let result = do --ignore-errors { cobra_active_help=0 run-external $cmd "__complete" ...$spans | complete }

        if $result != null and $result.exit_code == 0 {
            let completions = $result.stdout | lines

            # the directive is the last line
            let directive = do -i { $completions | last | str replace ':' '' | into int }

            let completions = $completions | drop | each { |it| 
                # the first word is the command, the rest is the description
                let words = $it | split row -r '\s{1}'

                # If the last span contains a hypen and equals, attach it to the name
                let last_span = $spans | last
                let words = if ($last_span =~ '^-') and ($last_span =~ '=$') {
                    $words | each {|it| $"($last_span)($it)" }
                } else {
                    $words
                }

                {value: ($words | first | str trim), description: ($words | skip | str join ' ')}
            }

            {completions: $completions, directive: $directive}
        } else {
            {completions: [], directive: -1}
        }
    }

    if (not ($rest | is-empty)) {
        let result = exec_complete $rest
        let completions = $result.completions
        let directive = $result.directive

        # Add space at the end of each completion
        let completions = if $directive != $ShellCompDirectiveNoSpace {
          $completions | each {|it| {value: $"($it.value) ", description: $it.description}}
        } else {
          $completions
        }

        # Cobra returns a list of completions that are supported with this directive
        # There is no way to currently support this in a nushell external completer
        let completions = if $directive == $ShellCompDirectiveFilterFileExt {
          []
        } else {
          $completions
        }

        if $directive == $ShellCompDirectiveNoFileComp {
          # Allow empty results as this will stop file completion
          $completions
        } else if ($completions | is-empty)  or  $directive == $ShellCompDirectiveError {
          # Not returning null causes file completions to break
          # Return null if there are no completions or ShellCompDirectiveError
          null
        } else {
          $completions
        }
            
        if ($completions | is-empty) {
            null
        } else {
            $completions
        }
    } else {
        null
    }
}
`, ShellCompDirectiveError, ShellCompDirectiveNoSpace, ShellCompDirectiveNoFileComp,
        ShellCompDirectiveFilterFileExt, ShellCompDirectiveFilterDirs))

    _, err := buf.WriteTo(w)
    return err
}

func (c *Command) GenNushellCompletionFile(filename string, includeDesc bool) error {
    outFile, err := os.Create(filename)
    if err != nil {
        return err
    }
    defer outFile.Close()

    return c.GenNushellCompletion(outFile, includeDesc)
}
