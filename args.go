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

type PositionalArgs func(cmd *Command, args []string) error

// legacyArgs validates the arguments for a given command based on whether it has subcommands or not.
//
// Behavior:
// - Root commands without subcommands can take arbitrary arguments.
// - Root commands with subcommands perform subcommand-specific validity checks.
// - Subcommands always accept arbitrary arguments.
//
// Parameters:
//   - cmd: A pointer to the Command struct representing the command being checked.
//   - args: A slice of strings representing the arguments passed to the command.
//
// Returns:
//   - error: If the arguments are invalid for the given command, an error is returned. Otherwise, nil.
func legacyArgs(cmd *Command, args []string) error {
	// no subcommand, always take args
	if !cmd.HasSubCommands() {
		return nil
	}

	// root command with subcommands, do subcommand checking.
	if !cmd.HasParent() && len(args) > 0 {
		return fmt.Errorf("unknown command %q for %q%s", args[0], cmd.CommandPath(), cmd.findSuggestions(args[0]))
	}
	return nil
}

// NoArgs returns an error if any arguments are provided.
// It checks whether the given slice of arguments `args` is empty.
// If `args` contains elements, it indicates that unexpected arguments were provided,
// and thus returns a formatted error indicating the unknown command and its path.
// If no arguments are present, it returns nil, indicating successful validation.
func NoArgs(cmd *Command, args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("unknown command %q for %q", args[0], cmd.CommandPath())
	}
	return nil
}

// OnlyValidArgs checks if the provided arguments are valid based on the `ValidArgs` field of the given command.
// It returns an error if any argument is not a valid option. If no validation is required, it returns nil.
// The ValidArgs field can contain descriptions for each valid argument, which should be ignored during validation.
func OnlyValidArgs(cmd *Command, args []string) error {
	if len(cmd.ValidArgs) > 0 {
		// Remove any description that may be included in ValidArgs.
		// A description is following a tab character.
		validArgs := make([]string, 0, len(cmd.ValidArgs))
		for _, v := range cmd.ValidArgs {
			validArgs = append(validArgs, strings.SplitN(v, "\t", 2)[0])
		}
		for _, v := range args {
			if !stringInSlice(v, validArgs) {
				return fmt.Errorf("invalid argument %q for %q%s", v, cmd.CommandPath(), cmd.findSuggestions(args[0]))
			}
		}
	}
	return nil
}

// ArbitraryArgs executes a command with arbitrary arguments and does not return an error.
func ArbitraryArgs(cmd *Command, args []string) error {
	return nil
}

// MinimumNArgs returns a PositionalArgs function that checks if there are at least N arguments provided.
// If fewer than N arguments are given, it returns an error with a message indicating the required number of arguments and the actual count received.
func MinimumNArgs(n int) PositionalArgs {
	return func(cmd *Command, args []string) error {
		if len(args) < n {
			return fmt.Errorf("requires at least %d arg(s), only received %d", n, len(args))
		}
		return nil
	}
}

// MaximumNArgs returns a PositionalArgs function that ensures no more than N arguments are provided.
// If the number of arguments exceeds N, it returns an error indicating the maximum allowed arguments and the actual count received.
func MaximumNArgs(n int) PositionalArgs {
	return func(cmd *Command, args []string) error {
		if len(args) > n {
			return fmt.Errorf("accepts at most %d arg(s), received %d", n, len(args))
		}
		return nil
	}
}

// ExactArgs returns a PositionalArgs function that checks if the command has exactly `n` arguments.
// If the number of arguments is not exactly `n`, it returns an error with a descriptive message.
// Otherwise, it returns nil indicating no error.
func ExactArgs(n int) PositionalArgs {
	return func(cmd *Command, args []string) error {
		if len(args) != n {
			return fmt.Errorf("accepts %d arg(s), received %d", n, len(args))
		}
		return nil
	}
}

// RangeArgs returns an error if the number of args is not within the expected range.
// It takes two integers, min and max, representing the minimum and maximum number of arguments allowed.
// The function returns a PositionalArgs closure that can be used as a validator for command arguments.
// If the number of arguments does not fall within the specified range (inclusive), it returns an error with details about the expected and received argument count.
func RangeArgs(min int, max int) PositionalArgs {
	return func(cmd *Command, args []string) error {
		if len(args) < min || len(args) > max {
			return fmt.Errorf("accepts between %d and %d arg(s), received %d", min, max, len(args))
		}
		return nil
	}
}

// MatchAll combines multiple PositionalArgs to work as a single PositionalArg.
// It applies each provided PositionalArg in sequence to the command and arguments.
// If any of the PositionalArgs return an error, that error is returned immediately.
// If all PositionalArgs are successfully applied, nil is returned.
func MatchAll(pargs ...PositionalArgs) PositionalArgs {
	return func(cmd *Command, args []string) error {
		for _, parg := range pargs {
			if err := parg(cmd, args); err != nil {
				return err
			}
		}
		return nil
	}
}

// ExactValidArgs returns an error if there are not exactly N positional args OR
// there are any positional args that are not in the `ValidArgs` field of `Command`.
//
// Deprecated: use MatchAll(ExactArgs(n), OnlyValidArgs) instead.
func ExactValidArgs(n int) PositionalArgs {
	return MatchAll(ExactArgs(n), OnlyValidArgs)
}
