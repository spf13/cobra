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
	"strings"
	"testing"
)

func getCommand(args PositionalArgs, withValid bool) *Command {
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
	OnlyValid_WithInvalidArgs ExpectedErrors = iota
	NoArgs_WithArgs
	MinimumNArgs_WithLessArgs
	MaximumNArgs_WithMoreArgs
	ExactArgs_WithInvalidCount
	RangeArgs_WithInvalidCount
)

func expectErrorWithArg(err error, t *testing.T, ex ExpectedErrors, arg string) {
	if err == nil {
		t.Fatal("Expected an error")
	}

	var expected string
	switch ex {
	case OnlyValid_WithInvalidArgs:
		expected = `invalid argument "a" for "c"`
	case NoArgs_WithArgs:
		expected = `unknown command "` + arg + `" for "c"`
	case MinimumNArgs_WithLessArgs:
		expected = "requires at least 2 arg(s), only received 1"
	case MaximumNArgs_WithMoreArgs:
		expected = "accepts at most 2 arg(s), received 3"
	case ExactArgs_WithInvalidCount:
		expected = "accepts 2 arg(s), received 3"
	case RangeArgs_WithInvalidCount:
		expected = "accepts between 2 and 4 arg(s), received 1"
	}

	if got := err.Error(); got != expected {
		t.Errorf("Expected: %q, got: %q", expected, got)
	}
}

func expectError(err error, t *testing.T, ex ExpectedErrors) {
	expectErrorWithArg(err, t, ex, "")
}

// NoArgs

func TestNoArgs(t *testing.T) {
	c := getCommand(NoArgs, false)
	output, err := executeCommand(c)
	expectSuccess(output, err, t)
}

func TestNoArgs_WithArgs(t *testing.T) {
	c := getCommand(NoArgs, false)
	_, err := executeCommand(c, "one")
	expectErrorWithArg(err, t, NoArgs_WithArgs, "one")
}

func TestNoArgs_WithValid_WithArgs(t *testing.T) {
	c := getCommand(NoArgs, true)
	_, err := executeCommand(c, "one")
	expectErrorWithArg(err, t, NoArgs_WithArgs, "one")
}

func TestNoArgs_WithValid_WithInvalidArgs(t *testing.T) {
	c := getCommand(NoArgs, true)
	_, err := executeCommand(c, "a")
	expectErrorWithArg(err, t, NoArgs_WithArgs, "a")
}

func TestNoArgs_WithValidOnly_WithInvalidArgs(t *testing.T) {
	c := getCommand(MatchAll(OnlyValidArgs, NoArgs), true)
	_, err := executeCommand(c, "a")
	expectError(err, t, OnlyValid_WithInvalidArgs)
}

// OnlyValidArgs

func TestOnlyValidArgs(t *testing.T) {
	c := getCommand(OnlyValidArgs, true)
	output, err := executeCommand(c, "one", "two")
	expectSuccess(output, err, t)
}

func TestOnlyValidArgs_WithInvalidArgs(t *testing.T) {
	c := getCommand(OnlyValidArgs, true)
	_, err := executeCommand(c, "a")
	expectError(err, t, OnlyValid_WithInvalidArgs)
}

// ArbitraryArgs

func TestArbitraryArgs(t *testing.T) {
	c := getCommand(ArbitraryArgs, false)
	output, err := executeCommand(c, "a", "b")
	expectSuccess(output, err, t)
}

func TestArbitraryArgs_WithValid(t *testing.T) {
	c := getCommand(ArbitraryArgs, true)
	output, err := executeCommand(c, "one", "two")
	expectSuccess(output, err, t)
}

func TestArbitraryArgs_WithValid_WithInvalidArgs(t *testing.T) {
	c := getCommand(ArbitraryArgs, true)
	output, err := executeCommand(c, "a")
	expectSuccess(output, err, t)
}

func TestArbitraryArgs_WithValidOnly_WithInvalidArgs(t *testing.T) {
	c := getCommand(MatchAll(OnlyValidArgs, ArbitraryArgs), true)
	_, err := executeCommand(c, "a")
	expectError(err, t, OnlyValid_WithInvalidArgs)
}

// MinimumNArgs

func TestMinimumNArgs(t *testing.T) {
	c := getCommand(MinimumNArgs(2), false)
	output, err := executeCommand(c, "a", "b", "c")
	expectSuccess(output, err, t)
}

func TestMinimumNArgs_WithValid(t *testing.T) {
	c := getCommand(MinimumNArgs(2), true)
	output, err := executeCommand(c, "one", "three")
	expectSuccess(output, err, t)
}

func TestMinimumNArgs_WithValid__WithInvalidArgs(t *testing.T) {
	c := getCommand(MinimumNArgs(2), true)
	output, err := executeCommand(c, "a", "b")
	expectSuccess(output, err, t)
}

func TestMinimumNArgs_WithValidOnly_WithInvalidArgs(t *testing.T) {
	c := getCommand(MatchAll(OnlyValidArgs, MinimumNArgs(2)), true)
	_, err := executeCommand(c, "a", "b")
	expectError(err, t, OnlyValid_WithInvalidArgs)
}

func TestMinimumNArgs_WithLessArgs(t *testing.T) {
	c := getCommand(MinimumNArgs(2), false)
	_, err := executeCommand(c, "a")
	expectError(err, t, MinimumNArgs_WithLessArgs)
}

func TestMinimumNArgs_WithLessArgs_WithValid(t *testing.T) {
	c := getCommand(MinimumNArgs(2), true)
	_, err := executeCommand(c, "one")
	expectError(err, t, MinimumNArgs_WithLessArgs)
}

func TestMinimumNArgs_WithLessArgs_WithValid_WithInvalidArgs(t *testing.T) {
	c := getCommand(MinimumNArgs(2), true)
	_, err := executeCommand(c, "a")
	expectError(err, t, MinimumNArgs_WithLessArgs)
}

func TestMinimumNArgs_WithLessArgs_WithValidOnly_WithInvalidArgs(t *testing.T) {
	c := getCommand(MatchAll(OnlyValidArgs, MinimumNArgs(2)), true)
	_, err := executeCommand(c, "a")
	expectError(err, t, OnlyValid_WithInvalidArgs)
}

// MaximumNArgs

func TestMaximumNArgs(t *testing.T) {
	c := getCommand(MaximumNArgs(3), false)
	output, err := executeCommand(c, "a", "b")
	expectSuccess(output, err, t)
}

func TestMaximumNArgs_WithValid(t *testing.T) {
	c := getCommand(MaximumNArgs(2), true)
	output, err := executeCommand(c, "one", "three")
	expectSuccess(output, err, t)
}

func TestMaximumNArgs_WithValid_WithInvalidArgs(t *testing.T) {
	c := getCommand(MaximumNArgs(2), true)
	output, err := executeCommand(c, "a", "b")
	expectSuccess(output, err, t)
}

func TestMaximumNArgs_WithValidOnly_WithInvalidArgs(t *testing.T) {
	c := getCommand(MatchAll(OnlyValidArgs, MaximumNArgs(2)), true)
	_, err := executeCommand(c, "a", "b")
	expectError(err, t, OnlyValid_WithInvalidArgs)
}

func TestMaximumNArgs_WithMoreArgs(t *testing.T) {
	c := getCommand(MaximumNArgs(2), false)
	_, err := executeCommand(c, "a", "b", "c")
	expectError(err, t, MaximumNArgs_WithMoreArgs)
}

func TestMaximumNArgs_WithMoreArgs_WithValid(t *testing.T) {
	c := getCommand(MaximumNArgs(2), true)
	_, err := executeCommand(c, "one", "three", "two")
	expectError(err, t, MaximumNArgs_WithMoreArgs)
}

func TestMaximumNArgs_WithMoreArgs_WithValid_WithInvalidArgs(t *testing.T) {
	c := getCommand(MaximumNArgs(2), true)
	_, err := executeCommand(c, "a", "b", "c")
	expectError(err, t, MaximumNArgs_WithMoreArgs)
}

func TestMaximumNArgs_WithMoreArgs_WithValidOnly_WithInvalidArgs(t *testing.T) {
	c := getCommand(MatchAll(OnlyValidArgs, MaximumNArgs(2)), true)
	_, err := executeCommand(c, "a", "b", "c")
	expectError(err, t, OnlyValid_WithInvalidArgs)
}

// ExactArgs

func TestExactArgs(t *testing.T) {
	c := getCommand(ExactArgs(3), false)
	output, err := executeCommand(c, "a", "b", "c")
	expectSuccess(output, err, t)
}

func TestExactArgs_WithValid(t *testing.T) {
	c := getCommand(ExactArgs(3), true)
	output, err := executeCommand(c, "three", "one", "two")
	expectSuccess(output, err, t)
}

func TestExactArgs_WithValid_WithInvalidArgs(t *testing.T) {
	c := getCommand(ExactArgs(3), true)
	output, err := executeCommand(c, "three", "a", "two")
	expectSuccess(output, err, t)
}

func TestExactArgs_WithValidOnly_WithInvalidArgs(t *testing.T) {
	c := getCommand(MatchAll(OnlyValidArgs, ExactArgs(3)), true)
	_, err := executeCommand(c, "three", "a", "two")
	expectError(err, t, OnlyValid_WithInvalidArgs)
}

func TestExactArgs_WithInvalidCount(t *testing.T) {
	c := getCommand(ExactArgs(2), false)
	_, err := executeCommand(c, "a", "b", "c")
	expectError(err, t, ExactArgs_WithInvalidCount)
}

func TestExactArgs_WithInvalidCount_WithValid(t *testing.T) {
	c := getCommand(ExactArgs(2), true)
	_, err := executeCommand(c, "three", "one", "two")
	expectError(err, t, ExactArgs_WithInvalidCount)
}

func TestExactArgs_WithInvalidCount_WithValid_WithInvalidArgs(t *testing.T) {
	c := getCommand(ExactArgs(2), true)
	_, err := executeCommand(c, "three", "a", "two")
	expectError(err, t, ExactArgs_WithInvalidCount)
}

func TestExactArgs_WithInvalidCount_WithValidOnly_WithInvalidArgs(t *testing.T) {
	c := getCommand(MatchAll(OnlyValidArgs, ExactArgs(2)), true)
	_, err := executeCommand(c, "three", "a", "two")
	expectError(err, t, OnlyValid_WithInvalidArgs)
}

// RangeArgs

func TestRangeArgs(t *testing.T) {
	c := getCommand(RangeArgs(2, 4), false)
	output, err := executeCommand(c, "a", "b", "c")
	expectSuccess(output, err, t)
}

func TestRangeArgs_WithValid(t *testing.T) {
	c := getCommand(RangeArgs(2, 4), true)
	output, err := executeCommand(c, "three", "one", "two")
	expectSuccess(output, err, t)
}

func TestRangeArgs_WithValid_WithInvalidArgs(t *testing.T) {
	c := getCommand(RangeArgs(2, 4), true)
	output, err := executeCommand(c, "three", "a", "two")
	expectSuccess(output, err, t)
}

func TestRangeArgs_WithValidOnly_WithInvalidArgs(t *testing.T) {
	c := getCommand(MatchAll(OnlyValidArgs, RangeArgs(2, 4)), true)
	_, err := executeCommand(c, "three", "a", "two")
	expectError(err, t, OnlyValid_WithInvalidArgs)
}

func TestRangeArgs_WithInvalidCount(t *testing.T) {
	c := getCommand(RangeArgs(2, 4), false)
	_, err := executeCommand(c, "a")
	expectError(err, t, RangeArgs_WithInvalidCount)
}

func TestRangeArgs_WithInvalidCount_WithValid(t *testing.T) {
	c := getCommand(RangeArgs(2, 4), true)
	_, err := executeCommand(c, "two")
	expectError(err, t, RangeArgs_WithInvalidCount)
}

func TestRangeArgs_WithInvalidCount_WithValid_WithInvalidArgs(t *testing.T) {
	c := getCommand(RangeArgs(2, 4), true)
	_, err := executeCommand(c, "a")
	expectError(err, t, RangeArgs_WithInvalidCount)
}

func TestRangeArgs_WithInvalidCount_WithValidOnly_WithInvalidArgs(t *testing.T) {
	c := getCommand(MatchAll(OnlyValidArgs, RangeArgs(2, 4)), true)
	_, err := executeCommand(c, "a")
	expectError(err, t, OnlyValid_WithInvalidArgs)
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
	c := getCommand(ExactValidArgs(3), true)
	output, err := executeCommand(c, "three", "one", "two")
	expectSuccess(output, err, t)
}

func TestExactValidArgs_WithInvalidCount(t *testing.T) {
	c := getCommand(ExactValidArgs(2), false)
	_, err := executeCommand(c, "three", "one", "two")
	expectError(err, t, ExactArgs_WithInvalidCount)
}

func TestExactValidArgs_WithInvalidCount_WithInvalidArgs(t *testing.T) {
	c := getCommand(ExactValidArgs(2), true)
	_, err := executeCommand(c, "three", "a", "two")
	expectError(err, t, ExactArgs_WithInvalidCount)
}

func TestExactValidArgs_WithInvalidArgs(t *testing.T) {
	c := getCommand(ExactValidArgs(2), true)
	_, err := executeCommand(c, "three", "a")
	expectError(err, t, OnlyValid_WithInvalidArgs)
}

// This test make sure we keep backwards-compatibility with respect
// to the legacyArgs() function.
// It makes sure the root command accepts arguments if it does not have
// sub-commands.
func TestLegacyArgsRootAcceptsArgs(t *testing.T) {
	rootCmd := &Command{Use: "root", Args: nil, Run: emptyRun}

	_, err := executeCommand(rootCmd, "somearg")
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
