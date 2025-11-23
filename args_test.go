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
	"testing"
)

func newCommand(args PositionalArgs, withValid bool) *Command {
	c := &Command{
		Use:  "c",
		Args: args,
		Run:  emptyRun,
	}
	if withValid {
		c.ValidArgs = []string{"one", "two", "three"}
	}
	return c
}

func expectSuccess(output string, err error, t *testing.T) {
	if output != "" {
		t.Errorf("Unexpected output: %v", output)
	}
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

type ExpectedErrors int64

const (
	OnlyValidWithInvalidArgs ExpectedErrors = iota
	NoArgsWithArgs
	MinimumNArgsWithLessArgs
	MaximumNArgsWithMoreArgs
	ExactArgsWithInvalidCount
	RangeArgsWithInvalidCount
)

func expectErrorWithArg(err error, t *testing.T, ex ExpectedErrors, arg string) {
	if err == nil {
		t.Fatal("Expected an error")
	}
	expected := map[ExpectedErrors]string{
		OnlyValidWithInvalidArgs:  `invalid argument "a" for "c"`,
		NoArgsWithArgs:            `unknown command "` + arg + `" for "c"`,
		MinimumNArgsWithLessArgs:  "requires at least 2 arg(s), only received 1",
		MaximumNArgsWithMoreArgs:  "accepts at most 2 arg(s), received 3",
		ExactArgsWithInvalidCount: "accepts 2 arg(s), received 3",
		RangeArgsWithInvalidCount: "accepts between 2 and 4 arg(s), received 1",
	}[ex]
	if got := err.Error(); got != expected {
		t.Errorf("Expected: %q, got: %q", expected, got)
	}
}

func expectError(err error, t *testing.T, ex ExpectedErrors) {
	expectErrorWithArg(err, t, ex, "")
}

// NoArgs

func TestNoArgs(t *testing.T) {
	output, err := executeCommand(newCommand(NoArgs, false))
	expectSuccess(output, err, t)
}

func TestNoArgs_WithArgs(t *testing.T) {
	_, err := executeCommand(newCommand(NoArgs, false), "one")
	expectErrorWithArg(err, t, NoArgsWithArgs, "one")
}

func TestNoArgs_WithValid_WithArgs(t *testing.T) {
	_, err := executeCommand(newCommand(NoArgs, true), "one")
	expectErrorWithArg(err, t, NoArgsWithArgs, "one")
}

func TestNoArgs_WithValid_WithInvalidArgs(t *testing.T) {
	_, err := executeCommand(newCommand(NoArgs, true), "a")
	expectErrorWithArg(err, t, NoArgsWithArgs, "a")
}

func TestNoArgs_WithValidOnly_WithInvalidArgs(t *testing.T) {
	_, err := executeCommand(newCommand(MatchAll(OnlyValidArgs, NoArgs), true), "a")
	expectError(err, t, OnlyValidWithInvalidArgs)
}

// OnlyValidArgs

func TestOnlyValidArgs(t *testing.T) {
	output, err := executeCommand(newCommand(OnlyValidArgs, true), "one", "two")
	expectSuccess(output, err, t)
}

func TestOnlyValidArgs_WithInvalidArgs(t *testing.T) {
	_, err := executeCommand(newCommand(OnlyValidArgs, true), "a")
	expectError(err, t, OnlyValidWithInvalidArgs)
}

// ArbitraryArgs

func TestArbitraryArgs(t *testing.T) {
	output, err := executeCommand(newCommand(ArbitraryArgs, false), "a", "b")
	expectSuccess(output, err, t)
}

func TestArbitraryArgs_WithValid(t *testing.T) {
	output, err := executeCommand(newCommand(ArbitraryArgs, true), "one", "two")
	expectSuccess(output, err, t)
}

func TestArbitraryArgs_WithValid_WithInvalidArgs(t *testing.T) {
	output, err := executeCommand(newCommand(ArbitraryArgs, true), "a")
	expectSuccess(output, err, t)
}

func TestArbitraryArgs_WithValidOnly_WithInvalidArgs(t *testing.T) {
	_, err := executeCommand(newCommand(MatchAll(OnlyValidArgs, ArbitraryArgs), true), "a")
	expectError(err, t, OnlyValidWithInvalidArgs)
}

// MinimumNArgs

func TestMinimumNArgs(t *testing.T) {
	output, err := executeCommand(newCommand(MinimumNArgs(2), false), "a", "b", "c")
	expectSuccess(output, err, t)
}

func TestMinimumNArgs_WithValid(t *testing.T) {
	output, err := executeCommand(newCommand(MinimumNArgs(2), true), "one", "three")
	expectSuccess(output, err, t)
}

func TestMinimumNArgs_WithValid__WithInvalidArgs(t *testing.T) {
	output, err := executeCommand(newCommand(MinimumNArgs(2), true), "a", "b")
	expectSuccess(output, err, t)
}

func TestMinimumNArgs_WithValidOnly_WithInvalidArgs(t *testing.T) {
	_, err := executeCommand(newCommand(MatchAll(OnlyValidArgs, MinimumNArgs(2)), true), "a", "b")
	expectError(err, t, OnlyValidWithInvalidArgs)
}

func TestMinimumNArgs_WithLessArgs(t *testing.T) {
	_, err := executeCommand(newCommand(MinimumNArgs(2), false), "a")
	expectError(err, t, MinimumNArgsWithLessArgs)
}

func TestMinimumNArgs_WithLessArgs_WithValid(t *testing.T) {
	_, err := executeCommand(newCommand(MinimumNArgs(2), true), "one")
	expectError(err, t, MinimumNArgsWithLessArgs)
}

func TestMinimumNArgs_WithLessArgs_WithValid_WithInvalidArgs(t *testing.T) {
	_, err := executeCommand(newCommand(MinimumNArgs(2), true), "a")
	expectError(err, t, MinimumNArgsWithLessArgs)
}

func TestMinimumNArgs_WithLessArgs_WithValidOnly_WithInvalidArgs(t *testing.T) {
	_, err := executeCommand(newCommand(MatchAll(OnlyValidArgs, MinimumNArgs(2)), true), "a")
	expectError(err, t, OnlyValidWithInvalidArgs)
}

// MaximumNArgs

func TestMaximumNArgs(t *testing.T) {
	output, err := executeCommand(newCommand(MaximumNArgs(3), false), "a", "b")
	expectSuccess(output, err, t)
}

func TestMaximumNArgs_WithValid(t *testing.T) {
	output, err := executeCommand(newCommand(MaximumNArgs(2), true), "one", "three")
	expectSuccess(output, err, t)
}

func TestMaximumNArgs_WithValid_WithInvalidArgs(t *testing.T) {
	output, err := executeCommand(newCommand(MaximumNArgs(2), true), "a", "b")
	expectSuccess(output, err, t)
}

func TestMaximumNArgs_WithValidOnly_WithInvalidArgs(t *testing.T) {
	_, err := executeCommand(newCommand(MatchAll(OnlyValidArgs, MaximumNArgs(2)), true), "a", "b")
	expectError(err, t, OnlyValidWithInvalidArgs)
}

func TestMaximumNArgs_WithMoreArgs(t *testing.T) {
	_, err := executeCommand(newCommand(MaximumNArgs(2), false), "a", "b", "c")
	expectError(err, t, MaximumNArgsWithMoreArgs)
}

func TestMaximumNArgs_WithMoreArgs_WithValid(t *testing.T) {
	_, err := executeCommand(newCommand(MaximumNArgs(2), true), "one", "three", "two")
	expectError(err, t, MaximumNArgsWithMoreArgs)
}

func TestMaximumNArgs_WithMoreArgs_WithValid_WithInvalidArgs(t *testing.T) {
	_, err := executeCommand(newCommand(MaximumNArgs(2), true), "a", "b", "c")
	expectError(err, t, MaximumNArgsWithMoreArgs)
}

func TestMaximumNArgs_WithMoreArgs_WithValidOnly_WithInvalidArgs(t *testing.T) {
	_, err := executeCommand(newCommand(MatchAll(OnlyValidArgs, MaximumNArgs(2)), true), "a", "b", "c")
	expectError(err, t, OnlyValidWithInvalidArgs)
}

// ExactArgs

func TestExactArgs(t *testing.T) {
	output, err := executeCommand(newCommand(ExactArgs(3), false), "a", "b", "c")
	expectSuccess(output, err, t)
}

func TestExactArgs_WithValid(t *testing.T) {
	output, err := executeCommand(newCommand(ExactArgs(3), true), "three", "one", "two")
	expectSuccess(output, err, t)
}

func TestExactArgs_WithValid_WithInvalidArgs(t *testing.T) {
	output, err := executeCommand(newCommand(ExactArgs(3), true), "three", "a", "two")
	expectSuccess(output, err, t)
}

func TestExactArgs_WithValidOnly_WithInvalidArgs(t *testing.T) {
	_, err := executeCommand(newCommand(MatchAll(OnlyValidArgs, ExactArgs(3)), true), "three", "a", "two")
	expectError(err, t, OnlyValidWithInvalidArgs)
}

func TestExactArgs_WithInvalidCount(t *testing.T) {
	_, err := executeCommand(newCommand(ExactArgs(2), false), "a", "b", "c")
	expectError(err, t, ExactArgsWithInvalidCount)
}

func TestExactArgs_WithInvalidCount_WithValid(t *testing.T) {
	_, err := executeCommand(newCommand(ExactArgs(2), true), "three", "one", "two")
	expectError(err, t, ExactArgsWithInvalidCount)
}

func TestExactArgs_WithInvalidCount_WithValid_WithInvalidArgs(t *testing.T) {
	_, err := executeCommand(newCommand(ExactArgs(2), true), "three", "a", "two")
	expectError(err, t, ExactArgsWithInvalidCount)
}

func TestExactArgs_WithInvalidCount_WithValidOnly_WithInvalidArgs(t *testing.T) {
	_, err := executeCommand(newCommand(MatchAll(OnlyValidArgs, ExactArgs(2)), true), "three", "a", "two")
	expectError(err, t, OnlyValidWithInvalidArgs)
}

// RangeArgs

func TestRangeArgs(t *testing.T) {
	output, err := executeCommand(newCommand(RangeArgs(2, 4), false), "a", "b", "c")
	expectSuccess(output, err, t)
}

func TestRangeArgs_WithValid(t *testing.T) {
	output, err := executeCommand(newCommand(RangeArgs(2, 4), true), "three", "one", "two")
	expectSuccess(output, err, t)
}

func TestRangeArgs_WithValid_WithInvalidArgs(t *testing.T) {
	output, err := executeCommand(newCommand(RangeArgs(2, 4), true), "three", "a", "two")
	expectSuccess(output, err, t)
}

func TestRangeArgs_WithValidOnly_WithInvalidArgs(t *testing.T) {
	_, err := executeCommand(newCommand(MatchAll(OnlyValidArgs, RangeArgs(2, 4)), true), "three", "a", "two")
	expectError(err, t, OnlyValidWithInvalidArgs)
}

func TestRangeArgs_WithInvalidCount(t *testing.T) {
	_, err := executeCommand(newCommand(RangeArgs(2, 4), false), "a")
	expectError(err, t, RangeArgsWithInvalidCount)
}

func TestRangeArgs_WithInvalidCount_WithValid(t *testing.T) {
	_, err := executeCommand(newCommand(RangeArgs(2, 4), true), "two")
	expectError(err, t, RangeArgsWithInvalidCount)
}

func TestRangeArgs_WithInvalidCount_WithValid_WithInvalidArgs(t *testing.T) {
	_, err := executeCommand(newCommand(RangeArgs(2, 4), true), "a")
	expectError(err, t, RangeArgsWithInvalidCount)
}

func TestRangeArgs_WithInvalidCount_WithValidOnly_WithInvalidArgs(t *testing.T) {
	_, err := executeCommand(newCommand(MatchAll(OnlyValidArgs, RangeArgs(2, 4)), true), "a")
	expectError(err, t, OnlyValidWithInvalidArgs)
}

// Takes(No)Args

func TestRootTakesNoArgs(t *testing.T) {
	rootCmd := &Command{Use: "root", Run: emptyRun}
	childCmd := &Command{Use: "child", Run: emptyRun}
	rootCmd.AddCommand(childCmd)

	_, err := executeCommand(rootCmd, "illegal", "args")
	if err == nil {
		t.Fatal("Expected an error")
	}

	got := err.Error()
	expected := `unknown command "illegal" for "root"`
	if !strings.Contains(got, expected) {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestRootTakesArgs(t *testing.T) {
	rootCmd := &Command{Use: "root", Args: ArbitraryArgs, Run: emptyRun}
	childCmd := &Command{Use: "child", Run: emptyRun}
	rootCmd.AddCommand(childCmd)

	_, err := executeCommand(rootCmd, "legal", "args")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestChildTakesNoArgs(t *testing.T) {
	rootCmd := &Command{Use: "root", Run: emptyRun}
	childCmd := &Command{Use: "child", Args: NoArgs, Run: emptyRun}
	rootCmd.AddCommand(childCmd)

	_, err := executeCommand(rootCmd, "child", "illegal", "args")
	if err == nil {
		t.Fatal("Expected an error")
	}

	got := err.Error()
	expected := `unknown command "illegal" for "root child"`
	if !strings.Contains(got, expected) {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestChildTakesArgs(t *testing.T) {
	rootCmd := &Command{Use: "root", Run: emptyRun}
	childCmd := &Command{Use: "child", Args: ArbitraryArgs, Run: emptyRun}
	rootCmd.AddCommand(childCmd)

	_, err := executeCommand(rootCmd, "child", "legal", "args")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestMatchAll(t *testing.T) {
	// Somewhat contrived example check that ensures there are exactly 3
	// arguments, and each argument is exactly 2 bytes long.
	pargs := MatchAll(
		ExactArgs(3),
		func(cmd *Command, args []string) error {
			for _, arg := range args {
				if len([]byte(arg)) != 2 {
					return fmt.Errorf("expected to be exactly 2 bytes long")
				}
			}
			return nil
		},
	)

	testCases := map[string]struct {
		args []string
		fail bool
	}{
		"happy path": {
			[]string{"aa", "bb", "cc"},
			false,
		},
		"incorrect number of args": {
			[]string{"aa", "bb", "cc", "dd"},
			true,
		},
		"incorrect number of bytes in one arg": {
			[]string{"aa", "bb", "abc"},
			true,
		},
	}

	rootCmd := &Command{Use: "root", Args: pargs, Run: emptyRun}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			_, err := executeCommand(rootCmd, tc.args...)
			if err != nil && !tc.fail {
				t.Errorf("unexpected: %v\n", err)
			}
			if err == nil && tc.fail {
				t.Errorf("expected error")
			}
		})
	}
}

// DEPRECATED

func TestExactValidArgs(t *testing.T) {
	output, err := executeCommand(newCommand(ExactValidArgs(3), true), "three", "one", "two")
	expectSuccess(output, err, t)
}

func TestExactValidArgs_WithInvalidCount(t *testing.T) {
	_, err := executeCommand(newCommand(ExactValidArgs(2), false), "three", "one", "two")
	expectError(err, t, ExactArgsWithInvalidCount)
}

func TestExactValidArgs_WithInvalidCount_WithInvalidArgs(t *testing.T) {
	_, err := executeCommand(newCommand(ExactValidArgs(2), true), "three", "a", "two")
	expectError(err, t, ExactArgsWithInvalidCount)
}

func TestExactValidArgs_WithInvalidArgs(t *testing.T) {
	_, err := executeCommand(newCommand(ExactValidArgs(2), true), "three", "a")
	expectError(err, t, OnlyValidWithInvalidArgs)
}

// This test make sure we keep backwards-compatibility with respect
// to the legacyArgs() function.
// It makes sure the root command accepts arguments if it does not have
// sub-commands.
func TestLegacyArgsRootAcceptsArgs(t *testing.T) {
	_, err := executeCommand(&Command{Use: "root", Args: nil, Run: emptyRun}, "somearg")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

// This test make sure we keep backwards-compatibility with respect
// to the legacyArgs() function.
// It makes sure a sub-command accepts arguments and further sub-commands
func TestLegacyArgsSubcmdAcceptsArgs(t *testing.T) {
	rootCmd := &Command{Use: "root", Args: nil, Run: emptyRun}
	childCmd := &Command{Use: "child", Args: nil, Run: emptyRun}
	grandchildCmd := &Command{Use: "grandchild", Args: nil, Run: emptyRun}
	rootCmd.AddCommand(childCmd)
	childCmd.AddCommand(grandchildCmd)

	_, err := executeCommand(rootCmd, "child", "somearg")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}
