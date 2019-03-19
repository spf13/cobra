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

func expectError(err error, t *testing.T, ex string) {
	if err == nil {
		t.Fatal("Expected an error")
	}
	expected := map[string]string{
		"valid": `invalid argument "a" for "c"`,
		"no":    `unknown command "one" for "c"`,
		"min":   "requires at least 2 arg(s), only received 1",
		"max":   "accepts at most 2 arg(s), received 3",
		"exact": "accepts 2 arg(s), received 3",
		"range": "accepts between 2 and 4 arg(s), received 1",
	}[ex]
	if got := err.Error(); got != expected {
		t.Errorf("Expected: %q, got: %q", expected, got)
	}
}

// NoArgs

func TestNoArgs(t *testing.T) {
	o, e := executeCommand(newCommand(NoArgs, false))
	expectSuccess(o, e, t)
}

func TestNoArgs_WithArgs(t *testing.T) {
	_, e := executeCommand(newCommand(NoArgs, false), "one")
	expectError(e, t, "no")
}

func TestNoArgs_WithArgs_WithValid(t *testing.T) {
	_, e := executeCommand(newCommand(NoArgs, true), "one")
	expectError(e, t, "no")
}

// ArbitraryArgs

func TestArbitraryArgs(t *testing.T) {
	o, e := executeCommand(newCommand(ArbitraryArgs, false), "a", "b")
	expectSuccess(o, e, t)
}

func TestArbitraryArgs_WithValid(t *testing.T) {
	o, e := executeCommand(newCommand(ArbitraryArgs, true), "one", "two")
	expectSuccess(o, e, t)
}

func TestArbitraryArgs_WithValid_WithInvalidArgs(t *testing.T) {
	_, e := executeCommand(newCommand(ArbitraryArgs, true), "a")
	expectError(e, t, "valid")
}

// MinimumNArgs

func TestMinimumNArgs(t *testing.T) {
	o, e := executeCommand(newCommand(MinimumNArgs(2), false), "a", "b", "c")
	expectSuccess(o, e, t)
}

func TestMinimumNArgs_WithValid(t *testing.T) {
	o, e := executeCommand(newCommand(MinimumNArgs(2), true), "one", "three")
	expectSuccess(o, e, t)
}

func TestMinimumNArgs_WithValid_WithInvalidArgs(t *testing.T) {
	_, e := executeCommand(newCommand(MinimumNArgs(2), true), "a", "b")
	expectError(e, t, "valid")
}

func TestMinimumNArgs_WithLessArgs(t *testing.T) {
	_, e := executeCommand(newCommand(MinimumNArgs(2), false), "a")
	expectError(e, t, "min")
}

func TestMinimumNArgs_WithValid_WithLessArgs(t *testing.T) {
	_, e := executeCommand(newCommand(MinimumNArgs(2), true), "one")
	expectError(e, t, "min")
}

func TestMinimumNArgs_WithValid_WithLessArgsWithInvalidArgs(t *testing.T) {
	_, e := executeCommand(newCommand(MinimumNArgs(2), true), "a")
	expectError(e, t, "valid")
}

// MaximumNArgs

func TestMaximumNArgs(t *testing.T) {
	o, e := executeCommand(newCommand(MaximumNArgs(3), false), "a", "b")
	expectSuccess(o, e, t)
}

func TestMaximumNArgs_WithValid(t *testing.T) {
	o, e := executeCommand(newCommand(MaximumNArgs(2), true), "one", "three")
	expectSuccess(o, e, t)
}

func TestMaximumNArgs_WithValid_WithInvalidArgs(t *testing.T) {
	_, e := executeCommand(newCommand(MaximumNArgs(2), true), "a", "b")
	expectError(e, t, "valid")
}

func TestMaximumNArgs_WithMoreArgs(t *testing.T) {
	_, e := executeCommand(newCommand(MaximumNArgs(2), false), "a", "b", "c")
	expectError(e, t, "max")
}

func TestMaximumNArgs_WithValid_WithMoreArgs(t *testing.T) {
	_, e := executeCommand(newCommand(MaximumNArgs(2), true), "one", "three", "two")
	expectError(e, t, "max")
}

func TestMaximumNArgs_WithValid_WithMoreArgsWithInvalidArgs(t *testing.T) {
	_, e := executeCommand(newCommand(MaximumNArgs(2), true), "a", "b", "c")
	expectError(e, t, "valid")
}

// ExactArgs

func TestExactArgs(t *testing.T) {
	o, e := executeCommand(newCommand(ExactArgs(3), false), "a", "b", "c")
	expectSuccess(o, e, t)
}

func TestExactArgs_WithValid(t *testing.T) {
	o, e := executeCommand(newCommand(ExactArgs(3), true), "three", "one", "two")
	expectSuccess(o, e, t)
}

func TestExactArgs_WithValid_WithInvalidArgs(t *testing.T) {
	_, e := executeCommand(newCommand(ExactArgs(3), true), "three", "a", "two")
	expectError(e, t, "valid")
}

func TestExactArgs_WithInvalidCount(t *testing.T) {
	_, e := executeCommand(newCommand(ExactArgs(2), false), "a", "b", "c")
	expectError(e, t, "exact")
}

func TestExactArgs_WithValid_WithInvalidCount(t *testing.T) {
	_, e := executeCommand(newCommand(ExactArgs(2), true), "three", "one", "two")
	expectError(e, t, "exact")
}

func TestExactArgs_WithValid_WithInvalidCountWithInvalidArgs(t *testing.T) {
	_, e := executeCommand(newCommand(ExactArgs(2), true), "three", "a", "two")
	expectError(e, t, "valid")
}

// RangeArgs

func TestRangeArgs(t *testing.T) {
	o, e := executeCommand(newCommand(RangeArgs(2, 4), false), "a", "b", "c")
	expectSuccess(o, e, t)
}

func TestRangeArgs_WithValid(t *testing.T) {
	o, e := executeCommand(newCommand(RangeArgs(2, 4), true), "three", "one", "two")
	expectSuccess(o, e, t)
}

func TestRangeArgs_WithValid_WithInvalidArgs(t *testing.T) {
	_, e := executeCommand(newCommand(RangeArgs(2, 4), true), "three", "a", "two")
	expectError(e, t, "valid")
}

func TestRangeArgs_WithInvalidCount(t *testing.T) {
	_, e := executeCommand(newCommand(RangeArgs(2, 4), false), "a")
	expectError(e, t, "range")
}

func TestRangeArgs_WithValid_WithInvalidCount(t *testing.T) {
	_, e := executeCommand(newCommand(RangeArgs(2, 4), true), "two")
	expectError(e, t, "range")
}

func TestRangeArgs_WithValid_WithInvalidCountWithInvalidArgs(t *testing.T) {
	_, e := executeCommand(newCommand(RangeArgs(2, 4), true), "a")
	expectError(e, t, "valid")
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

func TestDEPRECATED_OnlyValidArgs(t *testing.T) {
	o, e := executeCommand(newCommand(OnlyValidArgs, true), "one", "two")
	expectSuccess(o, e, t)
}

func TestDEPRECATED_OnlyValidArgs_WithInvalidArgs(t *testing.T) {
	_, e := executeCommand(newCommand(OnlyValidArgs, true), "a")
	expectError(e, t, "valid")
}

func TestDEPRECATED_ExactValidArgs(t *testing.T) {
	// Note that the order is not required to be the same:
	// Definition: "one", "two", "three"
	// Execution: "two", "three", "one"
	o, e := executeCommand(newCommand(ExactValidArgs(3), true), "two", "three", "one")
	expectSuccess(o, e, t)
}

func TestDEPRECATED_ExactValidArgs_WithInvalidCount(t *testing.T) {
	_, e := executeCommand(newCommand(ExactValidArgs(2), true), "two", "three", "one")
	expectError(e, t, "exact")
}

func TestDEPRECATED_ExactValidArgs_WithInvalidArgs(t *testing.T) {
	_, e := executeCommand(newCommand(ExactValidArgs(2), true), "two", "a")
	expectError(e, t, "valid")
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
