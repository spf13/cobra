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

package doc

import (
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// opencliGenerator generates OpenCLI YAML output from cobra commands
type opencliGenerator struct {
	rootCmd *cobra.Command
}

// ExperimentalGenOpenCLI creates OpenCLI YAML structure from a cobra command.
func ExperimentalGenOpenCLI(cmd *cobra.Command) (*OpencliYaml, error) {
	// Find root command first
	rootCmd := cmd
	for rootCmd.HasParent() {
		rootCmd = rootCmd.Parent()
	}

	// Initialize help on root command to ensure all commands are available
	rootCmd.InitDefaultHelpCmd()
	rootCmd.InitDefaultHelpFlag()

	gen := &opencliGenerator{
		rootCmd: rootCmd,
	}

	// Build OpenCLI structure
	opencli := &OpencliYaml{
		Opencli: "1.0.0", // OpenCLI version
		Info: CliInfo{
			Title:   rootCmd.Name(),
			Version: rootCmd.Version,
		},
	}

	// Set summary and description from root command
	if rootCmd.Short != "" {
		opencli.Info.Summary = &rootCmd.Short
	}
	if rootCmd.Long != "" {
		opencli.Info.Description = &rootCmd.Long
	}

	// Set conventions (default values are already set in the struct)
	opencli.Conventions = &Conventions{
		GroupOptions:    true,
		OptionSeparator: " ",
	}

	// Always generate from root command to get complete OpenCLI spec
	opencli.Arguments = gen.extractArguments(rootCmd)
	opencli.Options = gen.extractOptions(rootCmd)
	opencli.Commands = gen.extractCommands(rootCmd)
	opencli.Examples = gen.extractExamples(rootCmd)

	return opencli, nil
}

// extractArguments parses the Use string to extract positional arguments
func (g *opencliGenerator) extractArguments(cmd *cobra.Command) []Argument {
	var args []Argument

	// Parse the Use string to find arguments
	// Pattern: [arg] or [arg...] or arg or arg...
	use := cmd.Use
	if use == "" {
		return args
	}

	// Find the command name (first word)
	parts := strings.Fields(use)
	if len(parts) == 0 {
		return args
	}

	// Get everything after the command name
	useAfterCmd := strings.TrimSpace(strings.TrimPrefix(use, parts[0]))
	if useAfterCmd == "" {
		return args
	}

	// Parse arguments, handling brackets properly
	// This regex matches: [optional arg], required arg, [arg...], arg...
	// We'll parse manually to handle nested brackets and multi-word arguments
	var current strings.Builder
	var bracketDepth int

	for i, r := range useAfterCmd {
		switch r {
		case '[':
			bracketDepth++
			if bracketDepth > 1 {
				current.WriteRune(r)
			}
		case ']':
			bracketDepth--
			if bracketDepth == 0 {
				// End of optional argument
				argName := strings.TrimSpace(current.String())
				if argName != "" {
					arg := g.parseArgumentName(argName, false, cmd)
					args = append(args, arg)
				}
				current.Reset()
			} else if bracketDepth > 0 {
				current.WriteRune(r)
			}
		case ' ':
			if bracketDepth == 0 {
				// End of required argument
				argName := strings.TrimSpace(current.String())
				if argName != "" {
					arg := g.parseArgumentName(argName, true, cmd)
					args = append(args, arg)
				}
				current.Reset()
			} else {
				current.WriteRune(r)
			}
		default:
			current.WriteRune(r)
		}

		// Handle last argument if we're at the end
		if i == len(useAfterCmd)-1 && current.Len() > 0 {
			argName := strings.TrimSpace(current.String())
			if argName != "" {
				required := bracketDepth == 0
				arg := g.parseArgumentName(argName, required, cmd)
				args = append(args, arg)
			}
		}
	}

	return args
}

// parseArgumentName parses an argument name and creates an Argument struct
func (g *opencliGenerator) parseArgumentName(name string, required bool, cmd *cobra.Command) Argument {
	arg := Argument{
		Name: name,
	}

	// Check for variadic (...)
	if strings.HasSuffix(name, "...") {
		v := 0
		arg.Name = strings.TrimSuffix(name, "...")
		// Set arity for variadic arguments
		arity := &Arity{
			Minimum: &v,
		}
		arg.Arity = arity
	}

	// Set required flag
	arg.Required = &required

	// Check for ValidArgs to set acceptedValues
	if len(cmd.ValidArgs) > 0 {
		acceptedValues := make([]string, 0, len(cmd.ValidArgs))
		for _, v := range cmd.ValidArgs {
			// Remove description if present (tab-separated)
			val := strings.SplitN(v, "\t", 2)[0]
			acceptedValues = append(acceptedValues, val)
		}
		if len(acceptedValues) > 0 {
			arg.AcceptedValues = acceptedValues
		}
	}

	return arg
}

// extractOptions converts cobra flags to OpenCLI Options
func (g *opencliGenerator) extractOptions(cmd *cobra.Command) []Option {
	var options []Option

	// Process non-inherited flags
	flags := cmd.NonInheritedFlags()
	flags.VisitAll(func(flag *pflag.Flag) {
		if flag.Hidden {
			return
		}

		opt := Option{
			Name:   flag.Name,
			Hidden: flag.Hidden,
		}

		// Set description
		if flag.Usage != "" {
			opt.Description = &flag.Usage
		}

		// Set aliases (shorthand)
		if flag.Shorthand != "" && len(flag.ShorthandDeprecated) == 0 {
			opt.Aliases = []string{flag.Shorthand}
		}

		// Check if required (flags with NoOptDefVal are typically required)
		// This is a heuristic - cobra doesn't have a direct "required" flag concept
		// but we can check if it has a default value
		if flag.DefValue == "" && flag.Value.Type() != "bool" {
			required := true
			opt.Required = &required
		}

		// Check if recursive (persistent flags are recursive)
		// Persistent flags are accessible from subcommands
		if cmd.PersistentFlags().Lookup(flag.Name) != nil {
			opt.Recursive = true
		}

		// Extract option arguments if the flag takes a value
		if flag.Value.Type() != "bool" {
			arg := Argument{
				Name: flag.Name,
			}
			if flag.Usage != "" {
				arg.Description = &flag.Usage
			}
			opt.Arguments = []Argument{arg}
		}

		options = append(options, opt)
	})

	return options
}

// extractCommands converts cobra commands to OpenCLI Commands recursively
func (g *opencliGenerator) extractCommands(cmd *cobra.Command) []Command {
	var commands []Command

	children := cmd.Commands()
	sort.Sort(byName(children))

	for _, child := range children {
		// Skip hidden and deprecated commands
		if child.Hidden || len(child.Deprecated) != 0 {
			continue
		}
		// Skip the auto-generated help command (added by InitDefaultHelpCmd)
		// The help command is identified by name "help" and IsAvailableCommand returns false for it
		// because it's the helpCommand of the parent
		if child.Name() == "help" && !child.IsAvailableCommand() {
			continue
		}
		// For OpenCLI, we include all user-defined commands, even if they're not runnable
		// or are help topic commands. We only exclude the auto-generated help command above.

		opencliCmd := Command{
			Name:   child.Name(),
			Hidden: child.Hidden,
		}

		// Set aliases
		if len(child.Aliases) > 0 {
			opencliCmd.Aliases = child.Aliases
		}

		// Set description
		if child.Short != "" {
			opencliCmd.Description = &child.Short
		}

		// Set examples
		if child.Example != "" {
			opencliCmd.Examples = []string{child.Example}
		}

		// Extract arguments
		opencliCmd.Arguments = g.extractArguments(child)

		// Extract options (non-inherited flags)
		opencliCmd.Options = g.extractOptions(child)

		// Recursively extract subcommands
		opencliCmd.Commands = g.extractCommands(child)

		commands = append(commands, opencliCmd)
	}

	return commands
}

// extractExamples extracts examples from the command
func (g *opencliGenerator) extractExamples(cmd *cobra.Command) []string {
	if cmd.Example == "" {
		return nil
	}

	// Split by newlines and filter empty lines
	lines := strings.Split(cmd.Example, "\n")
	var examples []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			examples = append(examples, line)
		}
	}

	return examples
}
