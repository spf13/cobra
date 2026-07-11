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
)

// ArgSpec documents one positional argument of a command. It is descriptive
// only: Cobra renders it in the help output and exposes it for introspection,
// but does not use it to validate input. Set Command.Args for validation.
type ArgSpec struct {
	// Name is the placeholder shown in help, rendered as <name>.
	Name string

	// Description explains what the argument is. Keep it to a short phrase; it
	// is aligned against the other arguments in the help output.
	Description string

	// Example is an optional value a caller can copy. When set, and not already
	// present in Description, it is appended as " (e.g. <example>)".
	Example string
}

// HasAvailableArguments reports whether the command documents any positional
// arguments to show in the help output.
func (c *Command) HasAvailableArguments() bool {
	return len(c.Arguments) > 0
}

// placeholder renders the argument name as it appears in help, wrapping a bare
// name in angle brackets. A name that already carries its own notation is left
// untouched, so the author keeps control of it: a bracketed form such as
// "[note]", "<id>", or "{a|b}", or a variadic suffix such as "files...". Any
// other name — including a bare dotted one like "config.file" — is wrapped.
func (a ArgSpec) placeholder() string {
	name := a.Name
	if name == "" {
		return name
	}
	if strings.HasSuffix(name, "...") {
		return name
	}
	switch name[0] {
	case '[', '<', '{', '(':
		return name
	}
	return "<" + name + ">"
}

// ArgumentUsages returns a formatted, aligned listing of the command's
// documented positional arguments, one per line, ready to print under an
// "Arguments:" heading. It mirrors the layout of flag usages.
func (c *Command) ArgumentUsages() string {
	if len(c.Arguments) == 0 {
		return ""
	}

	width := 0
	placeholders := make([]string, len(c.Arguments))
	for i, a := range c.Arguments {
		placeholders[i] = a.placeholder()
		if l := len(placeholders[i]); l > width {
			width = l
		}
	}

	var b strings.Builder
	for i, a := range c.Arguments {
		desc := a.Description
		if a.Example != "" && !strings.Contains(desc, a.Example) {
			if desc != "" {
				desc += " "
			}
			desc += "(e.g. " + a.Example + ")"
		}
		fmt.Fprintf(&b, "  %-*s   %s\n", width, placeholders[i], desc)
	}
	return b.String()
}
