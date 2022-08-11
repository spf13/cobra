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
	"strings"
	"testing"
)

func TestValidateFlagGroups(t *testing.T) {
	getCmd := func() *Command {
		cmd := &Command{
			Use: "testcmd",
			Run: func(cmd *Command, args []string) {},
		}

		cmd.Flags().String("a", "", "")
		cmd.Flags().String("b", "", "")
		cmd.Flags().String("c", "", "")
		cmd.Flags().String("d", "", "")
		cmd.PersistentFlags().String("p-a", "", "")
		cmd.PersistentFlags().String("p-b", "", "")
		cmd.PersistentFlags().String("p-c", "", "")

		subCmd := &Command{
			Use: "subcmd",
			Run: func(cmd *Command, args []string) {},
		}
		subCmd.Flags().String("sub-a", "", "")

		cmd.AddCommand(subCmd)

		return cmd
	}

	// Each test case uses a unique command from the function above.
	testcases := []struct {
		desc                 string
		requiredTogether     []string
		mutuallyExclusive    []string
		subRequiredTogether  []string
		subMutuallyExclusive []string
		args                 []string
		expectErr            string
	}{
		{
			desc: "No flags no problems",
		}, {
			desc:              "No flags no problems even with conflicting groups",
			requiredTogether:  []string{"a b"},
			mutuallyExclusive: []string{"a b"},
		}, {
			desc:             "Required together flag group validation fails",
			requiredTogether: []string{"a b c"},
			args:             []string{"--a=foo"},
			expectErr:        `flags [a b c] must be set together, but [b c] were not set`,
		}, {
			desc:             "Required together flag group validation passes",
			requiredTogether: []string{"a b c"},
			args:             []string{"--c=bar", "--a=foo", "--b=baz"},
		}, {
			desc:              "Mutually exclusive flag group validation fails",
			mutuallyExclusive: []string{"a b c"},
			args:              []string{"--b=foo", "--c=bar"},
			expectErr:         `exactly one of the flags [a b c] can be set, but [b c] were set`,
		}, {
			desc:              "Mutually exclusive flag group validation passes",
			mutuallyExclusive: []string{"a b c"},
			args:              []string{"--b=foo"},
		}, {
			desc:             "Multiple required together flag groups failed validation returns first error",
			requiredTogether: []string{"a b c", "a d"},
			args:             []string{"--d=foo", "--c=foo"},
			expectErr:        `flags [a b c] must be set together, but [a b] were not set`,
		}, {
			desc:              "Multiple mutually exclusive flag groups failed validation returns first error",
			mutuallyExclusive: []string{"a b c", "a d"},
			args:              []string{"--a=foo", "--c=foo", "--d=foo"},
			expectErr:         `exactly one of the flags [a b c] can be set, but [a c] were set`,
		}, {
			desc:              "Flag and persistent flags being in multiple groups fail required together group",
			requiredTogether:  []string{"a p-a", "p-a p-b"},
			mutuallyExclusive: []string{"p-b p-c"},
			args:              []string{"--a=foo", "--p-b=foo", "--p-c=foo"},
			expectErr:         `flags [a p-a] must be set together, but [p-a] were not set`,
		}, {
			desc:              "Flag and persistent flags being in multiple groups fail mutually exclusive group",
			requiredTogether:  []string{"a p-a", "p-a p-b"},
			mutuallyExclusive: []string{"p-b p-c"},
			args:              []string{"--a=foo", "--p-a=foo", "--p-b=foo", "--p-c=foo"},
			expectErr:         `exactly one of the flags [p-b p-c] can be set, but [p-b p-c] were set`,
		}, {
			desc:              "Flag and persistent flags pass required together and mutually exclusive groups",
			requiredTogether:  []string{"a p-a", "p-a p-b"},
			mutuallyExclusive: []string{"p-b p-c"},
			args:              []string{"--a=foo", "--p-a=foo", "--p-b=foo"},
		}, {
			desc:                "Required together flag group validation fails on subcommand with inherited flag",
			subRequiredTogether: []string{"p-a sub-a"},
			args:                []string{"subcmd", "--sub-a=foo"},
			expectErr:           `flags [p-a sub-a] must be set together, but [p-a] were not set`,
		}, {
			desc:                "Required together flag group validation passes on subcommand with inherited flag",
			subRequiredTogether: []string{"p-a sub-a"},
			args:                []string{"subcmd", "--p-a=foo", "--sub-a=foo"},
		}, {
			desc:                 "Mutually exclusive flag group validation fails on subcommand with inherited flag",
			subMutuallyExclusive: []string{"p-a sub-a"},
			args:                 []string{"subcmd", "--p-a=foo", "--sub-a=foo"},
			expectErr:            `exactly one of the flags [p-a sub-a] can be set, but [p-a sub-a] were set`,
		}, {
			desc:                 "Mutually exclusive flag group validation passes on subcommand with inherited flag",
			subMutuallyExclusive: []string{"p-a sub-a"},
			args:                 []string{"subcmd", "--p-a=foo"},
		}, {
			desc:                "Required together flag group validation is not applied on other command",
			subRequiredTogether: []string{"p-a sub-a"},
			args:                []string{"--p-a=foo"},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.desc, func(t *testing.T) {
			cmd := getCmd()
			subCmd := cmd.Commands()[0]

			for _, group := range tc.requiredTogether {
				cmd.MarkFlagsRequiredTogether(strings.Split(group, " ")...)
			}
			for _, group := range tc.mutuallyExclusive {
				cmd.MarkFlagsMutuallyExclusive(strings.Split(group, " ")...)
			}
			for _, group := range tc.subRequiredTogether {
				subCmd.MarkFlagsRequiredTogether(strings.Split(group, " ")...)
			}
			for _, group := range tc.subMutuallyExclusive {
				subCmd.MarkFlagsMutuallyExclusive(strings.Split(group, " ")...)
			}

			cmd.SetArgs(tc.args)
			err := cmd.Execute()

			switch {
			case err == nil && len(tc.expectErr) > 0:
				t.Errorf("Expected error %q but got nil", tc.expectErr)
			case err != nil && err.Error() != tc.expectErr:
				t.Errorf("Expected error %q but got %q", tc.expectErr, err)
			}
		})
	}
}
