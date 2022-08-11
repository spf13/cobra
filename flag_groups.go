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
	"fmt"

	flag "github.com/spf13/pflag"
)

// MarkFlagsRequiredTogether creates a relationship between flags, which ensures
// that if any of flags with names from flagNames is set, other flags must be set too.
func (c *Command) MarkFlagsRequiredTogether(flagNames ...string) {
	c.addFlagGroup(&requiredTogetherFlagGroup{
		flagNames: flagNames,
	})
}

// MarkFlagsMutuallyExclusive creates a relationship between flags, which ensures
// that if any of flags with names from flagNames is set, other flags must not be set.
func (c *Command) MarkFlagsMutuallyExclusive(flagNames ...string) {
	c.addFlagGroup(&mutuallyExclusiveFlagGroup{
		flagNames: flagNames,
	})
}

// addFlagGroup merges persistent flags of the command and adds flagGroup into command's flagGroups list.
// Panics, if flagGroup g contains the name of the flag, which is not defined in the Command c.
func (c *Command) addFlagGroup(g flagGroup) {
	c.mergePersistentFlags()

	for _, flagName := range g.AssignedFlagNames() {
		if c.Flags().Lookup(flagName) == nil {
			panic(fmt.Sprintf("flag %q is not defined", flagName))
		}
	}

	c.flagGroups = append(c.flagGroups, g)
}

// ValidateFlagGroups runs validation for each group from command's flagGroups list,
// and returns the first error encountered, or nil, if there were no validation errors.
func (c *Command) ValidateFlagGroups() error {
	setFlags := makeSetFlagsSet(c.Flags())
	for _, group := range c.flagGroups {
		if err := group.ValidateSetFlags(setFlags); err != nil {
			return err
		}
	}
	return nil
}

// adjustByFlagGroupsForCompletions changes the command by each flagGroup from command's flagGroups list
// to make the further command completions generation more convenient.
// Does nothing, if Command.DisableFlagParsing is true.
func (c *Command) adjustByFlagGroupsForCompletions() {
	if c.DisableFlagParsing {
		return
	}

	for _, group := range c.flagGroups {
		group.AdjustCommandForCompletions(c)
	}
}

type flagGroup interface {
	// ValidateSetFlags checks whether the combination of flags that have been set is valid.
	// If not, an error is returned.
	ValidateSetFlags(setFlags setFlagsSet) error

	// AssignedFlagNames returns a full list of flag names that have been assigned to the group.
	AssignedFlagNames() []string

	// AdjustCommandForCompletions updates the command to generate more convenient for this group completions.
	AdjustCommandForCompletions(c *Command)
}

// requiredTogetherFlagGroup groups flags that are required together and
// must all be set, if any of flags from this group is set.
type requiredTogetherFlagGroup struct {
	flagNames []string
}

func (g *requiredTogetherFlagGroup) AssignedFlagNames() []string {
	return g.flagNames
}
func (g *requiredTogetherFlagGroup) ValidateSetFlags(setFlags setFlagsSet) error {
	unset := setFlags.selectUnsetFlagNamesFrom(g.flagNames)

	if unsetCount := len(unset); unsetCount != 0 && unsetCount != len(g.flagNames) {
		return fmt.Errorf("flags %v must be set together, but %v were not set", g.flagNames, unset)
	}

	return nil
}
func (g *requiredTogetherFlagGroup) AdjustCommandForCompletions(c *Command) {
	setFlags := makeSetFlagsSet(c.Flags())
	if setFlags.hasAnyFrom(g.flagNames) {
		for _, requiredFlagName := range g.flagNames {
			_ = c.MarkFlagRequired(requiredFlagName)
		}
	}
}

// mutuallyExclusiveFlagGroup groups flags that are mutually exclusive
// and must not be set together, if any of flags from this group is set.
type mutuallyExclusiveFlagGroup struct {
	flagNames []string
}

func (g *mutuallyExclusiveFlagGroup) AssignedFlagNames() []string {
	return g.flagNames
}
func (g *mutuallyExclusiveFlagGroup) ValidateSetFlags(setFlags setFlagsSet) error {
	set := setFlags.selectSetFlagNamesFrom(g.flagNames)

	if len(set) > 1 {
		return fmt.Errorf("exactly one of the flags %v can be set, but %v were set", g.flagNames, set)
	}
	return nil
}
func (g *mutuallyExclusiveFlagGroup) AdjustCommandForCompletions(c *Command) {
	setFlags := makeSetFlagsSet(c.Flags())
	firstSetFlagName, hasAny := setFlags.selectFirstSetFlagNameFrom(g.flagNames)
	if hasAny {
		for _, exclusiveFlagName := range g.flagNames {
			if exclusiveFlagName != firstSetFlagName {
				c.Flags().Lookup(exclusiveFlagName).Hidden = true
			}
		}
	}
}

// setFlagsSet is a helper set type that is intended to be used to store names of the flags
// that have been set in flag.FlagSet and to perform some lookups and checks on those flags.
type setFlagsSet map[string]struct{}

// makeSetFlagsSet creates setFlagsSet of names of the flags that have been set in the given flag.FlagSet.
func makeSetFlagsSet(fs *flag.FlagSet) setFlagsSet {
	s := make(setFlagsSet)

	// Visit flags that have been set and add them to the set
	fs.Visit(func(f *flag.Flag) {
		s[f.Name] = struct{}{}
	})

	return s
}
func (s setFlagsSet) has(flagName string) bool {
	_, ok := s[flagName]
	return ok
}
func (s setFlagsSet) hasAnyFrom(flagNames []string) bool {
	for _, flagName := range flagNames {
		if s.has(flagName) {
			return true
		}
	}
	return false
}
func (s setFlagsSet) selectFirstSetFlagNameFrom(flagNames []string) (string, bool) {
	for _, flagName := range flagNames {
		if s.has(flagName) {
			return flagName, true
		}
	}
	return "", false
}
func (s setFlagsSet) selectSetFlagNamesFrom(flagNames []string) (setFlagNames []string) {
	for _, flagName := range flagNames {
		if s.has(flagName) {
			setFlagNames = append(setFlagNames, flagName)
		}
	}
	return
}
func (s setFlagsSet) selectUnsetFlagNamesFrom(flagNames []string) (unsetFlagNames []string) {
	for _, flagName := range flagNames {
		if !s.has(flagName) {
			unsetFlagNames = append(unsetFlagNames, flagName)
		}
	}
	return
}
