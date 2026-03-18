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
	"fmt"
	"strings"

	"github.com/spf13/pflag"
)

// commandInfo is extracted data for one entry in the help-all reference output.
type commandInfo struct {
	Path  string   // full command path, e.g. "app serve"
	Args  string   // argument spec from Use field, e.g. "[session_id]"
	Flags []string // formatted flag placeholders, e.g. "[--all]"
	Short string   // one-line description
}

// collectCommands walks the command tree and returns a commandInfo for each
// visible, runnable command (those where Runnable() is true). Hidden commands
// and their subtrees are skipped. Deprecated commands are included.
func collectCommands(root *Command) []commandInfo {
	if root == nil {
		return nil
	}
	var out []commandInfo
	walkCommands(root, &out)
	return out
}

// walkCommands recursively visits cmd and its children, appending a
// commandInfo for each visible, runnable command. Hidden commands are
// pruned along with their entire subtree.
func walkCommands(cmd *Command, out *[]commandInfo) {
	if cmd.Hidden {
		return
	}

	if cmd.Runnable() {
		info := commandInfo{
			Path:  cmd.CommandPath(),
			Args:  extractArgs(cmd.Use),
			Short: cmd.Short,
		}

		cmd.LocalFlags().VisitAll(func(f *pflag.Flag) {
			if f.Hidden {
				return
			}
			info.Flags = append(info.Flags, formatFlag(f))
		})

		*out = append(*out, info)
	}

	for _, child := range cmd.Commands() {
		walkCommands(child, out)
	}
}

// extractArgs returns everything after the first space in a Use string,
// which represents the argument placeholders.
func extractArgs(use string) string {
	if i := strings.IndexByte(use, ' '); i >= 0 {
		return use[i+1:]
	}
	return ""
}

// formatFlag returns a bracket-wrapped flag placeholder appropriate for the
// flag's type (bool/count get no value, int/duration/string get a placeholder).
func formatFlag(f *pflag.Flag) string {
	long := f.Name
	short := f.Shorthand

	switch f.Value.Type() {
	case "bool", "count":
		if short != "" {
			return fmt.Sprintf("[-%s, --%s]", short, long)
		}
		return fmt.Sprintf("[--%s]", long)
	case "int":
		if short != "" {
			return fmt.Sprintf("[-%s, --%s N]", short, long)
		}
		return fmt.Sprintf("[--%s N]", long)
	case "duration":
		if short != "" {
			return fmt.Sprintf("[-%s, --%s DURATION]", short, long)
		}
		return fmt.Sprintf("[--%s DURATION]", long)
	default:
		placeholder := strings.ToUpper(long)
		if short != "" {
			return fmt.Sprintf("[-%s, --%s %s]", short, long, placeholder)
		}
		return fmt.Sprintf("[--%s %s]", long, placeholder)
	}
}

// renderCommands formats command info into aligned, indented lines. When
// verbose is true, flag placeholders are included before the description.
func renderCommands(cmds []commandInfo, verbose bool) string {
	if len(cmds) == 0 {
		return ""
	}

	lefts := make([]string, len(cmds))
	maxLen := 0
	for i, c := range cmds {
		parts := []string{c.Path}
		if c.Args != "" {
			parts = append(parts, c.Args)
		}
		if verbose {
			parts = append(parts, c.Flags...)
		}
		lefts[i] = strings.Join(parts, " ")
		if len(lefts[i]) > maxLen {
			maxLen = len(lefts[i])
		}
	}

	var b strings.Builder
	indent := "    "
	for i, c := range cmds {
		padding := strings.Repeat(" ", maxLen-len(lefts[i])+2)
		fmt.Fprintf(&b, "%s%s%s# %s", indent, lefts[i], padding, c.Short)
		if i < len(cmds)-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

// NewHelpAllCommand returns a help-all command that prints a complete, aligned
// command reference. Add it to root with root.AddCommand(). Pass --verbose to
// include flag placeholders in the output.
func NewHelpAllCommand() *Command {
	var verbose bool

	cmd := &Command{
		Use:    "help-all",
		Short:  "List all commands with their arguments and descriptions",
		Args: NoArgs,
		RunE: func(cmd *Command, args []string) error {
			cmds := collectCommands(cmd.Root())
			out := renderCommands(cmds, verbose)
			if out != "" {
				_, err := fmt.Fprintln(cmd.OutOrStdout(), out)
				return err
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&verbose, "verbose", false, "Include flags in output")

	return cmd
}
