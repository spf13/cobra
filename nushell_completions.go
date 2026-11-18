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

// genNushellComp writes a nushell completion script for the program called name.
//
// The script declares the program as an extern and attaches a command-wide
// completer to it with the @complete attribute, so only this program is
// affected. The completer calls the program's __complete command and turns
// its output into nushell suggestions. It requires nushell 0.116.0 or later.
func genNushellComp(buf io.StringWriter, name string, includeDesc bool) {
	compCmd := ShellCompRequestCmd
	if !includeDesc {
		compCmd = ShellCompNoDescRequestCmd
	}
	WriteStringAndCheck(buf, fmt.Sprintf(`# nushell completion for %-36[1]s -*- shell-script -*-
#
# Requires nushell 0.116.0 or later. To load completions in every new session,
# save this script in a nushell autoload directory ($nu.user-autoload-dirs).
#
# Set $env.BASH_COMP_DEBUG_FILE to a file path to log what the completer does.

def "nu-complete %[1]s debug" [message: string] {
    let file = $env.BASH_COMP_DEBUG_FILE? | default ""
    if $file != "" {
        $"($message)\n" | save --append $file
    }
}

# Completes the arguments of %[1]s by calling "%[1]s %[2]s".
def "nu-complete %[1]s" [place: record] {
    let ShellCompDirectiveError = %[3]d
    let ShellCompDirectiveNoSpace = %[4]d
    let ShellCompDirectiveNoFileComp = %[5]d
    let ShellCompDirectiveFilterFileExt = %[6]d
    let ShellCompDirectiveFilterDirs = %[7]d
    let ShellCompDirectiveKeepOrder = %[8]d

    # $place.command holds the words of the command line, with aliases already
    # expanded, and ends with "" when the cursor is on an empty argument. This
    # is exactly what the %[2]s command expects after the program name.
    let args = $place.command | skip 1
    if ($args | is-empty) {
        return {completions: [], fallback: true}
    }
    let current = $args | last

    nu-complete %[1]s debug "========= starting completion logic =========="
    nu-complete %[1]s debug $"Calling: %[1]s %[2]s ($args | to nuon)"

    # Active help is not supported by nushell.
    let result = try {
        with-env {%[9]s: "0"} { run-external "%[1]s" "%[2]s" ...$args | complete }
    } catch {
        null
    }
    if $result == null or $result.exit_code != 0 {
        nu-complete %[1]s debug "Completion command failed; falling back to file completion"
        return {completions: [], fallback: true}
    }

    # The last line of the output is the directive, as ":<number>".
    let lines = $result.stdout | lines
    if ($lines | is-empty) or not (($lines | last) =~ '^:\d+$') {
        nu-complete %[1]s debug "No directive found; falling back to file completion"
        return {completions: [], fallback: true}
    }
    let directive = $lines | last | str substring 1.. | into int
    let candidates = $lines | drop | where {|line| $line != "" and not ($line | str starts-with "%[10]s") }
    nu-complete %[1]s debug $"Directive: ($directive), candidates: ($candidates | length)"

    if ($directive bit-and $ShellCompDirectiveError) != 0 {
        return {completions: []}
    }

    # With the --flag=value form, complete the value and keep the flag in front of it.
    let flag_form = $current | parse --regex '^(?<flag>-[^=]+=)(?<value>.*)$'
    let flag_prefix = if ($flag_form | is-empty) { "" } else { $flag_form.0.flag }
    let to_complete = if ($flag_form | is-empty) { $current } else { $flag_form.0.value }

    let filter_ext = ($directive bit-and $ShellCompDirectiveFilterFileExt) != 0
    let filter_dirs = ($directive bit-and $ShellCompDirectiveFilterDirs) != 0
    if $filter_ext or $filter_dirs {
        # The candidates are file extensions, or at most one directory to complete in.
        let values = $candidates | each {|line| $line | split row "\t" | first }
        let paths = if $filter_ext {
            $to_complete | commandline complete --type path --detailed
            | where {|path| $path.kind? == "directory" or ($values | any {|ext| $path.value | str ends-with $".($ext)" }) }
        } else if ($values | is-empty) {
            $to_complete | commandline complete --type directory --detailed
        } else {
            let base = $values | first | str trim --right --char "/"
            $"($base)/($to_complete)" | commandline complete --type directory --detailed
            | each {|path| $path | update value ($path.value | str replace $"($base)/" "") }
        }
        # Keep nushell's own suggestion fields so these behave like its file completion.
        return {
            completions: ($paths | each {|path|
                $path | reject --optional span type | update value $"($flag_prefix)($path.value)"
            })
            options: {filter: false}
        }
    }

    let completions = $candidates | each {|line|
        let parts = $line | split row "\t"
        let description = $parts | skip 1 | str join "\t" | str trim
        if $description == "" {
            {value: ($parts | first)}
        } else {
            {value: ($parts | first), description: $description}
        }
    }

    # Keep the completions that start with what was typed. When there are none,
    # keep the ones that contain it.
    let prefixed = $completions | where {|it| $it.value | str starts-with $to_complete }
    let completions = if ($prefixed | is-empty) {
        $completions | where {|it| $it.value | str contains $to_complete }
    } else {
        $prefixed
    }

    if ($completions | is-empty) {
        if ($directive bit-and $ShellCompDirectiveNoFileComp) != 0 {
            return {completions: []}
        }
        return {completions: [], fallback: true}
    }

    # nushell only sorts the completions it filters, so sort them here.
    let completions = if ($directive bit-and $ShellCompDirectiveKeepOrder) == 0 {
        $completions | sort-by value
    } else {
        $completions
    }

    let space = ($directive bit-and $ShellCompDirectiveNoSpace) == 0
    {
        completions: ($completions | each {|it|
            $it | update value $"($flag_prefix)($it.value)" | insert append_whitespace $space
        })
        # The script does its own filtering and sorting. The completions depend
        # on state outside nushell, so they must not be cached.
        options: {filter: false, sort: false, cache: false}
    }
}

@complete "nu-complete %[1]s"
extern "%[1]s" [...args]
`, name, compCmd,
		ShellCompDirectiveError, ShellCompDirectiveNoSpace, ShellCompDirectiveNoFileComp,
		ShellCompDirectiveFilterFileExt, ShellCompDirectiveFilterDirs, ShellCompDirectiveKeepOrder,
		activeHelpEnvVar(name), activeHelpMarker))
}

// GenNushellCompletion generates a nushell completion script and writes it to the passed writer.
// The script requires nushell 0.116.0 or later.
func (c *Command) GenNushellCompletion(w io.Writer, includeDesc bool) error {
	buf := new(bytes.Buffer)
	genNushellComp(buf, c.Name(), includeDesc)
	_, err := buf.WriteTo(w)
	return err
}

// GenNushellCompletionFile generates a nushell completion script file.
func (c *Command) GenNushellCompletionFile(filename string, includeDesc bool) error {
	outFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer outFile.Close()

	return c.GenNushellCompletion(outFile, includeDesc)
}
