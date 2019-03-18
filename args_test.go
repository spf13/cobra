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

func validWithInvalidArgs(err error, t *testing.T) {
	if err == nil {
		t.Fatal("Expected an error")
	}

	got := err.Error()
	expected := `invalid argument "a" for "c"`
	if got != expected {
		t.Errorf("Expected: %q, got: %q", expected, got)
	}
}

func noArgsWithArgs(err error, t *testing.T) {
	if err == nil {
		t.Fatal("Expected an error")
	}
	got := err.Error()
	expected := `unknown command "one" for "c"`
	if got != expected {
		t.Errorf("Expected: %q, got: %q", expected, got)
	}
}

func minimumNArgsWithLessArgs(err error, t *testing.T) {
	if err == nil {
		t.Fatal("Expected an error")
	}
	got := err.Error()
	expected := "requires at least 2 arg(s), only received 1"
	if got != expected {
		t.Fatalf("Expected %q, got %q", expected, got)
	}
}

func maximumNArgsWithMoreArgs(err error, t *testing.T) {
	if err == nil {
		t.Fatal("Expected an error")
	}

	got := err.Error()
	expected := "accepts at most 2 arg(s), received 3"
	if got != expected {
		t.Fatalf("Expected %q, got %q", expected, got)
	}
}

func exactArgsWithInvalidCount(err error, t *testing.T) {
	if err == nil {
		t.Fatal("Expected an error")
	}
	got := err.Error()
	expected := "accepts 2 arg(s), received 3"
	if got != expected {
		t.Fatalf("Expected %q, got %q", expected, got)
	}
}

func rangeArgsWithInvalidCount(err error, t *testing.T) {
	if err == nil {
		t.Fatal("Expected an error")
	}
	got := err.Error()
	expected := "accepts between 2 and 4 arg(s), received 1"
	if got != expected {
		t.Fatalf("Expected %q, got %q", expected, got)
	}
}

// NoArgs

func TestNoArgs(t *testing.T) {
	c := getCommand(NoArgs, false)
	output, err := executeCommand(c)
	expectSuccess(output, err, t)
}

func TestNoArgsWithArgs(t *testing.T) {
	c := getCommand(NoArgs, false)
	_, err := executeCommand(c, "one")
	noArgsWithArgs(err, t)
}

func TestNoArgsWithArgsWithValid(t *testing.T) {
	c := getCommand(NoArgs, true)
	_, err := executeCommand(c, "one")
	noArgsWithArgs(err, t)
}

// ArbitraryArgs

func TestArbitraryArgs(t *testing.T) {
	c := getCommand(ArbitraryArgs, false)
	output, err := executeCommand(c, "a", "b")
	expectSuccess(output, err, t)
}

func TestArbitraryArgsWithValid(t *testing.T) {
	c := getCommand(ArbitraryArgs, true)
	output, err := executeCommand(c, "one", "two")
	expectSuccess(output, err, t)
}

func TestArbitraryArgsWithValidWithInvalidArgs(t *testing.T) {
	c := getCommand(ArbitraryArgs, true)
	_, err := executeCommand(c, "a")
	validWithInvalidArgs(err, t)
}

// MinimumNArgs

func TestMinimumNArgs(t *testing.T) {
	c := getCommand(MinimumNArgs(2), false)
	output, err := executeCommand(c, "a", "b", "c")
	expectSuccess(output, err, t)
}

func TestMinimumNArgsWithValid(t *testing.T) {
	c := getCommand(MinimumNArgs(2), true)
	output, err := executeCommand(c, "one", "three")
	expectSuccess(output, err, t)
}

func TestMinimumNArgsWithValidWithInvalidArgs(t *testing.T) {
	c := getCommand(MinimumNArgs(2), true)
	_, err := executeCommand(c, "a", "b")
	validWithInvalidArgs(err, t)
}

func TestMinimumNArgsWithLessArgs(t *testing.T) {
	c := getCommand(MinimumNArgs(2), false)
	_, err := executeCommand(c, "a")
	minimumNArgsWithLessArgs(err, t)
}

func TestMinimumNArgsWithLessArgsWithValid(t *testing.T) {
	c := getCommand(MinimumNArgs(2), true)
	_, err := executeCommand(c, "one")
	minimumNArgsWithLessArgs(err, t)
}

func TestMinimumNArgsWithLessArgsWithValidWithInvalidArgs(t *testing.T) {
	c := getCommand(MinimumNArgs(2), true)
	_, err := executeCommand(c, "a")
	validWithInvalidArgs(err, t)
}

// MaximumNArgs

func TestMaximumNArgs(t *testing.T) {
	c := getCommand(MaximumNArgs(3), false)
	output, err := executeCommand(c, "a", "b")
	expectSuccess(output, err, t)
}

func TestMaximumNArgsWithValid(t *testing.T) {
	c := getCommand(MaximumNArgs(2), true)
	output, err := executeCommand(c, "one", "three")
	expectSuccess(output, err, t)
}

func TestMaximumNArgsWithValidWithInvalidArgs(t *testing.T) {
	c := getCommand(MaximumNArgs(2), true)
	_, err := executeCommand(c, "a", "b")
	validWithInvalidArgs(err, t)
}

func TestMaximumNArgsWithMoreArgs(t *testing.T) {
	c := getCommand(MaximumNArgs(2), false)
	_, err := executeCommand(c, "a", "b", "c")
	maximumNArgsWithMoreArgs(err, t)
}

func TestMaximumNArgsWithMoreArgsWithValid(t *testing.T) {
	c := getCommand(MaximumNArgs(2), true)
	_, err := executeCommand(c, "one", "three", "two")
	maximumNArgsWithMoreArgs(err, t)
}

func TestMaximumNArgsWithMoreArgsWithValidWithInvalidArgs(t *testing.T) {
	c := getCommand(MaximumNArgs(2), true)
	_, err := executeCommand(c, "a", "b", "c")
	validWithInvalidArgs(err, t)
}

// ExactArgs

func TestExactArgs(t *testing.T) {
	c := getCommand(ExactArgs(3), false)
	output, err := executeCommand(c, "a", "b", "c")
	expectSuccess(output, err, t)
}

func TestExactArgsWithValid(t *testing.T) {
	c := getCommand(ExactArgs(3), true)
	output, err := executeCommand(c, "three", "one", "two")
	expectSuccess(output, err, t)
}

func TestExactArgsWithValidWithInvalidArgs(t *testing.T) {
	c := getCommand(ExactArgs(3), true)
	_, err := executeCommand(c, "three", "a", "two")
	validWithInvalidArgs(err, t)
}

func TestExactArgsWithInvalidCount(t *testing.T) {
	c := getCommand(ExactArgs(2), false)
	_, err := executeCommand(c, "a", "b", "c")
	exactArgsWithInvalidCount(err, t)
}

func TestExactArgsWithInvalidCountWithValid(t *testing.T) {
	c := getCommand(ExactArgs(2), true)
	_, err := executeCommand(c, "three", "one", "two")
	exactArgsWithInvalidCount(err, t)
}

func TestExactArgsWithInvalidCountWithValidWithInvalidArgs(t *testing.T) {
	c := getCommand(ExactArgs(2), true)
	_, err := executeCommand(c, "three", "a", "two")
	validWithInvalidArgs(err, t)
}

// RangeArgs

func TestRangeArgs(t *testing.T) {
	c := getCommand(RangeArgs(2, 4), false)
	output, err := executeCommand(c, "a", "b", "c")
	expectSuccess(output, err, t)
}

func TestRangeArgsWithValid(t *testing.T) {
	c := getCommand(RangeArgs(2, 4), true)
	output, err := executeCommand(c, "three", "one", "two")
	expectSuccess(output, err, t)
}

func TestRangeArgsWithValidWithInvalidArgs(t *testing.T) {
	c := getCommand(RangeArgs(2, 4), true)
	_, err := executeCommand(c, "three", "a", "two")
	validWithInvalidArgs(err, t)
}

func TestRangeArgsWithInvalidCount(t *testing.T) {
	c := getCommand(RangeArgs(2, 4), false)
	_, err := executeCommand(c, "a")
	rangeArgsWithInvalidCount(err, t)
}

func TestRangeArgsWithInvalidCountWithValid(t *testing.T) {
	c := getCommand(RangeArgs(2, 4), true)
	_, err := executeCommand(c, "two")
	rangeArgsWithInvalidCount(err, t)
}

func TestRangeArgsWithInvalidCountWithValidWithInvalidArgs(t *testing.T) {
	c := getCommand(RangeArgs(2, 4), true)
	_, err := executeCommand(c, "a")
	validWithInvalidArgs(err, t)
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

func TestOnlyValidArgs(t *testing.T) {
	c := &Command{
		Use:       "c",
		Args:      OnlyValidArgs,
		ValidArgs: []string{"one", "two"},
		Run:       emptyRun,
	}

	output, err := executeCommand(c, "one", "two")
	if output != "" {
		t.Errorf("Unexpected output: %v", output)
	}
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestOnlyValidArgsWithInvalidArgs(t *testing.T) {
	c := &Command{
		Use:       "c",
		Args:      OnlyValidArgs,
		ValidArgs: []string{"one", "two"},
		Run:       emptyRun,
	}

	_, err := executeCommand(c, "three")
	if err == nil {
		t.Fatal("Expected an error")
	}

	got := err.Error()
	expected := `invalid argument "three" for "c"`
	if got != expected {
		t.Errorf("Expected: %q, got: %q", expected, got)
	}
}

func TestExactValidArgs(t *testing.T) {
	c := &Command{Use: "c", Args: ExactValidArgs(3), ValidArgs: []string{"a", "b", "c"}, Run: emptyRun}
	output, err := executeCommand(c, "a", "b", "c")
	if output != "" {
		t.Errorf("Unexpected output: %v", output)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestExactValidArgsWithInvalidCount(t *testing.T) {
	c := &Command{Use: "c", Args: ExactValidArgs(2), Run: emptyRun}
	_, err := executeCommand(c, "a", "b", "c")

	if err == nil {
		t.Fatal("Expected an error")
	}

	got := err.Error()
	expected := "accepts 2 arg(s), received 3"
	if got != expected {
		t.Fatalf("Expected %q, got %q", expected, got)
	}
}

func TestExactValidArgsWithInvalidArgs(t *testing.T) {
	c := &Command{
		Use:       "c",
		Args:      ExactValidArgs(1),
		ValidArgs: []string{"one", "two"},
		Run:       emptyRun,
	}

	_, err := executeCommand(c, "three")
	if err == nil {
		t.Fatal("Expected an error")
	}

	got := err.Error()
	expected := `invalid argument "three" for "c"`
	if got != expected {
		t.Errorf("Expected: %q, got: %q", expected, got)
	}
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
