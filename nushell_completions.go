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
	"regexp"
	"strings"

	"github.com/spf13/pflag"
)

var carrageReturnRE = regexp.MustCompile(`\r?\n`)

func descriptionString(desc string) string {
	// Remove any carriage returns, this will break the extern
	desc = carrageReturnRE.ReplaceAllString(desc, " ")

	// Lets keep the descriptions short-ish
	if len(desc) > 100 {
		desc = desc[0:97] + "..."
	}
	return desc
}

func GenNushellComp(c *Command, buf io.StringWriter, nameBuilder *strings.Builder, isRoot bool, includeDesc bool) {
	processFlags := func(flags *pflag.FlagSet) {
		flags.VisitAll(func(f *pflag.Flag) {
			WriteStringAndCheck(buf, fmt.Sprintf("\t--%[1]s", f.Name))

			if f.Shorthand != "" {
				WriteStringAndCheck(buf, fmt.Sprintf("(-%[1]s)", f.Shorthand))
			}

			if includeDesc && f.Usage != "" {
				desc := descriptionString(f.Usage)
				WriteStringAndCheck(buf, fmt.Sprintf("\t# %[1]s", desc))
			}

			WriteStringAndCheck(buf, "\n")

		})
	}

	cmdName := c.Name()
	// commands after root name will be like "git pull"
	if !isRoot {
		nameBuilder.WriteString(" ")
	}
	nameBuilder.WriteString(cmdName)

	// only create an extern block if there is something to put in it
	if len(c.ValidArgs) > 0 || c.HasAvailableFlags() {
		builderString := nameBuilder.String()

		// ensure there is a space before any previous content
		// otherwise it will break descriptions
		WriteStringAndCheck(buf, "\n")

		funcName := builderString
		if !isRoot {
			funcName = fmt.Sprintf("\"%[1]s\"", builderString)
		}

		if includeDesc && c.Short != "" {
			desc := descriptionString(c.Short)
			WriteStringAndCheck(buf, fmt.Sprintf("# %[1]s\n", desc))
		}
		WriteStringAndCheck(buf, fmt.Sprintf("export extern %[1]s [\n", funcName))

		// valid args
		for _, arg := range c.ValidArgs {
			WriteStringAndCheck(buf, fmt.Sprintf("\t%[1]s?\n", arg))
		}

		processFlags(c.InheritedFlags())
		processFlags(c.LocalFlags())

		// End extern statement
		WriteStringAndCheck(buf, "]\n")
	}

	// process sub commands
	for _, child := range c.Commands() {
		childBuilder := strings.Builder{}
		childBuilder.WriteString(nameBuilder.String())
		GenNushellComp(child, buf, &childBuilder, false, includeDesc)
	}

}

func (c *Command) GenNushellCompletion(w io.Writer, includeDesc bool) error {
	var nameBuilder strings.Builder
	buf := new(bytes.Buffer)
	GenNushellComp(c, buf, &nameBuilder, true, includeDesc)

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
