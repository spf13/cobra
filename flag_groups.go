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
	flags.VisitAll(func(pflag *flag.Flag) {
		processFlagForGroupAnnotation(flags, pflag, requiredAsGroup, groupStatus)
		processFlagForGroupAnnotation(flags, pflag, mutuallyExclusive, mutuallyExclusiveGroupStatus)
	})

	if err := validateRequiredFlagGroups(groupStatus); err != nil {
		return err
	}
	if err := validateExclusiveFlagGroups(mutuallyExclusiveGroupStatus); err != nil {
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
