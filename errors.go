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
	"errors"
	"fmt"
	"strings"
)

// InvalidArgCountError is the error returned when the wrong number of arguments
// are supplied to a command.
type InvalidArgCountError struct {
	cmd     *Command
	args    []string
	atLeast int
	atMost  int
}

// Error implements error.
func (e InvalidArgCountError) Error() string {
	if e.atLeast == e.atMost && e.atLeast >= 0 { // ExactArgs
		return fmt.Sprintf("accepts %d arg(s), received %d", e.atLeast, len(e.args))
	}

	if e.atMost >= 0 && e.atLeast >= 0 { // RangeArgs
		return fmt.Sprintf("accepts between %d and %d arg(s), received %d", e.atLeast, e.atMost, len(e.args))
	}

	if e.atLeast >= 0 { // MinimumNArgs
		return fmt.Sprintf("requires at least %d arg(s), only received %d", e.atLeast, len(e.args))
	}

	// MaximumNArgs
	return fmt.Sprintf("accepts at most %d arg(s), received %d", e.atMost, len(e.args))
}

// InvalidArgCountError is the error returned when an invalid argument is present.
type InvalidArgValueError struct {
	cmd         *Command
	arg         string
	suggestions []string
}

// Error implements error.
func (e InvalidArgValueError) Error() string {
	return fmt.Sprintf("invalid argument %q for %q%s", e.arg, e.cmd.CommandPath(), helpTextForSuggestions(e.suggestions))
}

// UnknownSubcommandError is the error returned when a subcommand can not be
// found.
type UnknownSubcommandError struct {
	cmd         *Command
	subcmd      string
	suggestions []string
}

// Error implements error.
func (e UnknownSubcommandError) Error() string {
	return fmt.Sprintf("unknown command %q for %q%s", e.subcmd, e.cmd.CommandPath(), helpTextForSuggestions(e.suggestions))
}

// RequiredFlagError is the error returned when a required flag is not set.
type RequiredFlagError struct {
	missingFlagNames []string
}

// Error implements error.
func (e RequiredFlagError) Error() string {
	return fmt.Sprintf(`required flag(s) "%s" not set`, strings.Join(e.missingFlagNames, `", "`))
}

// FlagGroupError is the error returned when mutually-required or
// mutually-exclusive flags are not properly specified.
type FlagGroupError struct {
	err          error
	flagList     string
	problemFlags []string
}

var (
	// errFlagsAreMutuallyExclusive indicates that more than one flag marked by MarkFlagsMutuallyExclusive was provided.
	errFlagsAreMutuallyExclusive = errors.New("if any is set, none of the others can be")

	// errFlagsAreRequiredTogether indicates that only one of the flags marked by MarkFlagsRequiredTogether were provided.
	errFlagsAreRequiredTogether = errors.New("if any is set, they must all be set")

	// errFlagsAreOneRequired indicates that none of the flags marked by MarkFlagsOneRequired flags were provided.
	errFlagsAreOneRequired = errors.New("at least one of the flags is required")
)

// Error implements error.
func (e FlagGroupError) Error() string {
	switch {
	case errors.Is(e.err, errFlagsAreRequiredTogether):
		return fmt.Sprintf("if any flags in the group [%v] are set they must all be set; missing %v", e.flagList, e.problemFlags)
	case errors.Is(e.err, errFlagsAreOneRequired):
		return fmt.Sprintf("at least one of the flags in the group [%v] is required", e.flagList)
	case errors.Is(e.err, errFlagsAreMutuallyExclusive):
		return fmt.Sprintf("if any flags in the group [%v] are set none of the others can be; %v were all set", e.flagList, e.problemFlags)
	}

	// If the error struct is empty (i.e. wasn't created by Cobra), e.err will be nil.
	// We don't have a message to print, so instead just print the struct contents.
	return fmt.Sprintf("%#v", e)
}
