// Copyright © 2022 Steve Francia <spf@spf13.com>.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cobra

import (
	"fmt"
	"sort"
	"strings"

	flag "github.com/spf13/pflag"
)

const (
	requiredAsGroup   = "cobra_annotation_required_if_others_set"
	mutuallyExclusive = "cobra_annotation_mutually_exclusive"
	dependsOn         = "cobra_annotation_depends_on"
	dependsOnAny      = "cobra_annotation_depends_on_any"
)

// MarkFlagsRequiredTogether marks the given flags with annotations so that Cobra errors
// if the command is invoked with a subset (but not all) of the given flags.
func (c *Command) MarkFlagsRequiredTogether(flagNames ...string) {
	c.mergePersistentFlags()
	for _, v := range flagNames {
		f := c.Flags().Lookup(v)
		if f == nil {
			panic(fmt.Sprintf("Failed to find flag %q and mark it as being required in a flag group", v))
		}
		if err := c.Flags().SetAnnotation(v, requiredAsGroup, append(f.Annotations[requiredAsGroup], strings.Join(flagNames, " "))); err != nil {
			// Only errs if the flag isn't found.
			panic(err)
		}
	}
}

// MarkFlagsMutuallyExclusive marks the given flags with annotations so that Cobra errors
// if the command is invoked with more than one flag from the given set of flags.
func (c *Command) MarkFlagsMutuallyExclusive(flagNames ...string) {
	c.mergePersistentFlags()
	for _, v := range flagNames {
		f := c.Flags().Lookup(v)
		if f == nil {
			panic(fmt.Sprintf("Failed to find flag %q and mark it as being in a mutually exclusive flag group", v))
		}
		// Each time this is called is a single new entry; this allows it to be a member of multiple groups if needed.
		if err := c.Flags().SetAnnotation(v, mutuallyExclusive, append(f.Annotations[mutuallyExclusive], strings.Join(flagNames, " "))); err != nil {
			panic(err)
		}
	}
}

// MarkFlagsDependsOn marks the given flags with annotations so that Cobra errors
// if the command is invoked with 1 or more flags that are dependent on a specified
// other.
func (c *Command) MarkFlagsDependsOn(flagNames ...string) {
	const format = "Failed to find flag %q and mark it as being part of depends on group"
	c.markAnnotation(dependsOn, format, flagNames...)
}

// MarkFlagDependsOnAny marks the given flags with annotations so that Cobra errors
// if the command is invoked with a flag that is dependent on any 1 of a group of others.
func (c *Command) MarkFlagDependsOnAny(flagNames ...string) {
	const format = "Failed to find flag %q and mark it as being part of depends on any group"
	c.markAnnotation(dependsOnAny, format, flagNames...)
}

// markAnnotation currently only used by MarkFlagsDependsOn and MarkFlagDependsOnAny,
// but is generic enough and should be used by MarkFlagsRequiredTogether and
// MarkFlagsMutuallyExclusive.
// - format must contain a single place holder
func (c *Command) markAnnotation(annotation, format string, flagNames ...string) {
	c.mergePersistentFlags()
	for _, name := range flagNames {
		c.setFlagAnnotation(name, annotation,
			fmt.Sprintf(format, name),
			flagNames...,
		)
	}
}

func (c *Command) setFlagAnnotation(flag string, annotation string, message string, flagNames ...string) {
	f := c.Flags().Lookup(flag)
	if f == nil {
		panic(message)
	}
	ordered := strings.Join(flagNames, " ")
	if err := c.Flags().SetAnnotation(
		flag, annotation,
		append(f.Annotations[annotation], ordered),
	); err != nil {
		panic(err)
	}
}

// The 'special-ness' of a group means that the first member of the group carries
// special meaning. In contrast to the other group types, where all members are equal.
type specialStatusInfoData map[string]bool

type specialGroupInfo struct {
	special string
	others  []string
	data    specialStatusInfoData // maps the flag name to special status info
}
type specialGroupInfoCollection map[string]*specialGroupInfo

func newSpecialGroup(specialName string, others []string) *specialGroupInfo {
	size := len(others) + 1
	result := specialGroupInfo{
		special: specialName,
		others:  others,
		data:    make(specialStatusInfoData, size),
	}

	return &result
}

// validateFlagGroups validates the mutuallyExclusive/requiredAsGroup logic and returns the
// first error encountered.
func (c *Command) validateFlagGroups() error {
	if c.DisableFlagParsing {
		return nil
	}

	flags := c.Flags()

	// groupStatus format is the list of flags as a unique ID,
	// then a map of each flag name and whether it is set or not.
	groupStatus := map[string]map[string]bool{}
	mutuallyExclusiveGroupStatus := map[string]map[string]bool{}
	dependsOnSpecialGroupStatus := specialGroupInfoCollection{}
	dependsOnAnySpecialGroupStatus := specialGroupInfoCollection{}
	flags.VisitAll(func(pflag *flag.Flag) {
		processFlagForGroupAnnotation(flags, pflag, requiredAsGroup, groupStatus)
		processFlagForGroupAnnotation(flags, pflag, mutuallyExclusive, mutuallyExclusiveGroupStatus)
		processFlagForSpecialGroupAnnotation(flags, pflag, dependsOn, dependsOnSpecialGroupStatus)
		processFlagForSpecialGroupAnnotation(flags, pflag, dependsOnAny, dependsOnAnySpecialGroupStatus)
	})

	if err := validateRequiredFlagGroups(groupStatus); err != nil {
		return err
	}
	if err := validateExclusiveFlagGroups(mutuallyExclusiveGroupStatus); err != nil {
		return err
	}
	if err := validateDependsOnFlagGroups(dependsOnSpecialGroupStatus); err != nil {
		return err
	}
	if err := validateDependsOnAnyFlagGroups(dependsOnAnySpecialGroupStatus); err != nil {
		return err
	}
	return nil
}

func hasAllFlags(fs *flag.FlagSet, flagnames ...string) bool {
	for _, fname := range flagnames {
		f := fs.Lookup(fname)
		if f == nil {
			return false
		}
	}
	return true
}

func hasAnyOfFlags(fs *flag.FlagSet, flagnames ...string) bool {
	for _, fname := range flagnames {
		f := fs.Lookup(fname)
		if f != nil {
			return true
		}
	}
	return false
}

func processFlagForGroupAnnotation(flags *flag.FlagSet, pflag *flag.Flag, annotation string, groupStatus map[string]map[string]bool) {
	groupInfo, found := pflag.Annotations[annotation]
	if found {
		for _, group := range groupInfo {
			if groupStatus[group] == nil {
				flagnames := strings.Split(group, " ")

				// Only consider this flag group at all if all the flags are defined.
				if !hasAllFlags(flags, flagnames...) {
					continue
				}

				groupStatus[group] = map[string]bool{}
				for _, name := range flagnames {
					groupStatus[group][name] = false
				}
			}

			groupStatus[group][pflag.Name] = pflag.Changed
		}
	}
}

func processFlagForSpecialGroupAnnotation(flags *flag.FlagSet, pflag *flag.Flag,
	annotation string, groupStatus specialGroupInfoCollection) {

	if groupInfo, found := pflag.Annotations[annotation]; found {
		for _, group := range groupInfo {
			if groupStatus[group] == nil {

				flagnames := strings.Split(group, " ")
				// it's important to know that the order of the flags is established
				// in setFlagAnnotation, which makes the assumption of the first
				// item being special, being valid
				special := flagnames[0]
				others := flagnames[1:]
				isFlagSpecial := pflag.Name == special

				// Only consider this flag group at all if the first flag (Special)
				// is set and at least 1 of the others is
				if isFlagSpecial && flags.Lookup(special) == nil {
					continue
				}

				if !isFlagSpecial && !hasAnyOfFlags(flags, others...) {
					continue
				}

				groupStatus[group] = newSpecialGroup(special, others)
				for _, name := range flagnames {
					groupStatus[group].data[name] = false

					if name == special {
						break // short circuit after finding special
					}
				}
			}

			// group exists, but we still need to check if the flag exists in the group,
			// because the previous loop is short circuited as soon as we find the special.
			if _, found := groupStatus[group].data[pflag.Name]; !found {
				groupStatus[group].data[pflag.Name] = false
			}
			groupStatus[group].data[pflag.Name] = pflag.Changed
		}
	}
}

func validateRequiredFlagGroups(data map[string]map[string]bool) error {
	keys := sortedKeys(data)
	for _, flagList := range keys {
		flagnameAndStatus := data[flagList]

		unset := []string{}
		for flagname, isSet := range flagnameAndStatus {
			if !isSet {
				unset = append(unset, flagname)
			}
		}
		if len(unset) == len(flagnameAndStatus) || len(unset) == 0 {
			continue
		}

		// Sort values, so they can be tested/scripted against consistently.
		sort.Strings(unset)
		return fmt.Errorf("if any flags in the group [%v] are set they must all be set; missing %v", flagList, unset)
	}

	return nil
}

func validateExclusiveFlagGroups(data map[string]map[string]bool) error {
	keys := sortedKeys(data)
	for _, flagList := range keys {
		flagnameAndStatus := data[flagList]
		var set []string
		for flagname, isSet := range flagnameAndStatus {
			if isSet {
				set = append(set, flagname)
			}
		}
		if len(set) == 0 || len(set) == 1 {
			continue
		}

		// Sort values, so they can be tested/scripted against consistently.
		sort.Strings(set)
		return fmt.Errorf("if any flags in the group [%v] are set none of the others can be; %v were all set", flagList, set)
	}
	return nil
}

func validateDependsOnFlagGroups(data specialGroupInfoCollection) error {
	keys := sortedKeysSpecial(data)

	for _, flagList := range keys {
		flagnameAndStatus := data[flagList]

		if flagnameAndStatus.data[flagnameAndStatus.special] {
			// rule is satisfied, because the special flag is present, regardless of
			// the presence of the other members in the group
			return nil
		}

		// we have a problem if at least one of present is set, because special is not set
		present := []string{}
		for _, o := range flagnameAndStatus.others {
			if flagnameAndStatus.data[o] {
				present = append(present, o)
			}
		}
		if len(present) == 0 {
			continue
		}
		sort.Strings(present)

		return fmt.Errorf(
			"if any flags in the group %v are set then [%v] must be present; only %v were set",
			flagnameAndStatus.others, flagnameAndStatus.special, present,
		)
	}
	return nil
}

func validateDependsOnAnyFlagGroups(data specialGroupInfoCollection) error {
	keys := sortedKeysSpecial(data)

	for _, flagList := range keys {
		flagnameAndStatus := data[flagList]
		if !flagnameAndStatus.data[flagnameAndStatus.special] {
			return nil
		}

		present := []string{}
		for _, o := range flagnameAndStatus.others {
			if flagnameAndStatus.data[o] {
				present = append(present, o)
			}
		}
		if len(present) > 0 {
			continue
		}

		return fmt.Errorf(
			"if [%v] is present, then at least one of the flags in %v must be; none were set",
			flagnameAndStatus.special, flagnameAndStatus.others,
		)
	}
	return nil
}

func sortedKeys(m map[string]map[string]bool) []string {
	keys := make([]string, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	return keys
}

// implemented as a duplicate of sortedKeys as generics can't be used yet
func sortedKeysSpecial(m specialGroupInfoCollection) []string {
	keys := make([]string, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	return keys
}

// enforceFlagGroupsForCompletion will do the following:
// - when a flag in a group is present, other flags in the group will be marked required
// - when a flag in a mutually exclusive group is present, other flags in the group will be marked as hidden
// This allows the standard completion logic to behave appropriately for flag groups
func (c *Command) enforceFlagGroupsForCompletion() {
	if c.DisableFlagParsing {
		return
	}

	flags := c.Flags()
	groupStatus := map[string]map[string]bool{}
	mutuallyExclusiveGroupStatus := map[string]map[string]bool{}
	dependsOnSpecialGroupStatus := specialGroupInfoCollection{}
	c.Flags().VisitAll(func(pflag *flag.Flag) {
		processFlagForGroupAnnotation(flags, pflag, requiredAsGroup, groupStatus)
		processFlagForGroupAnnotation(flags, pflag, mutuallyExclusive, mutuallyExclusiveGroupStatus)
		processFlagForSpecialGroupAnnotation(flags, pflag, dependsOn, dependsOnSpecialGroupStatus)
	})

	// If a flag that is part of a group is present, we make all the other flags
	// of that group required so that the shell completion suggests them automatically
	for flagList, flagnameAndStatus := range groupStatus {
		for _, isSet := range flagnameAndStatus {
			if isSet {
				// One of the flags of the group is set, mark the other ones as required
				for _, fName := range strings.Split(flagList, " ") {
					_ = c.MarkFlagRequired(fName)
				}
			}
		}
	}

	// If a flag that is mutually exclusive to others is present, we hide the other
	// flags of that group so the shell completion does not suggest them
	for flagList, flagnameAndStatus := range mutuallyExclusiveGroupStatus {
		for flagName, isSet := range flagnameAndStatus {
			if isSet {
				// One of the flags of the mutually exclusive group is set, mark the other ones as hidden
				// Don't mark the flag that is already set as hidden because it may be an
				// array or slice flag and therefore must continue being suggested
				for _, fName := range strings.Split(flagList, " ") {
					if fName != flagName {
						flag := c.Flags().Lookup(fName)
						flag.Hidden = true
					}
				}
			}
		}
	}

	// if any of others is set, then mark special as required
	for _, flagnameAndStatus := range dependsOnSpecialGroupStatus {
		for _, o := range flagnameAndStatus.others {
			if flagnameAndStatus.data[o] {
				c.MarkFlagRequired(flagnameAndStatus.special)
				break
			}
		}
	}
}
