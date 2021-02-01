package cobra

import (
	"fmt"
	"strings"
	"testing"
)

type argsTestcase struct {
	exerr  string         // Expected error key (see map[string][string])
	args   PositionalArgs // Args validator
	wValid bool           // Define `ValidArgs` in the command
	rargs  []string       // Runtime args
}

var errStrings = map[string]string{
	"invalid":    `invalid argument "a" for "c"`,
	"unknown":    `unknown command "one" for "c"`,
	"less":       "requires at least 2 arg(s), only received 1",
	"more":       "accepts at most 2 arg(s), received 3",
	"notexact":   "accepts 2 arg(s), received 3",
	"notinrange": "accepts between 2 and 4 arg(s), received 1",
}

func (tc *argsTestcase) test(t *testing.T) {
	c := &Command{
		Use:  "c",
		Args: tc.args,
		Run:  emptyRun,
	}
	if tc.wValid {
		c.ValidArgs = []string{"one", "two", "three"}
	}

	o, e := executeCommand(c, tc.rargs...)

	if len(tc.exerr) > 0 {
		// Expect error
		if e == nil {
			t.Fatal("Expected an error")
		}
		expected, ok := errStrings[tc.exerr]
		if !ok {
			t.Errorf(`key "%s" is not found in map "errStrings"`, tc.exerr)
			return
		}
		if got := e.Error(); got != expected {
			t.Errorf("Expected: %q, got: %q", expected, got)
		}
	} else {
		// Expect success
		if o != "" {
			t.Errorf("Unexpected output: %v", o)
		}
		if e != nil {
			t.Fatalf("Unexpected error: %v", e)
		}
	}
}

func testArgs(t *testing.T, tests map[string]argsTestcase) {
	for name, tc := range tests {
		t.Run(name, tc.test)
	}
}

func TestArgs_No(t *testing.T) {
	testArgs(t, map[string]argsTestcase{
		"      | ":      {"", NoArgs, false, []string{}},
		"      | Arb":   {"unknown", NoArgs, false, []string{"one"}},
		"Valid | Valid": {"unknown", NoArgs, true, []string{"one"}},
	})
}
func TestArgs_Nil(t *testing.T) {
	testArgs(t, map[string]argsTestcase{
		"      | Arb":     {"", nil, false, []string{"a", "b"}},
		"Valid | Valid":   {"", nil, true, []string{"one", "two"}},
		"Valid | Invalid": {"invalid", nil, true, []string{"a"}},
	})
}
func TestArgs_Arbitrary(t *testing.T) {
	testArgs(t, map[string]argsTestcase{
		"      | Arb":     {"", ArbitraryArgs, false, []string{"a", "b"}},
		"Valid | Valid":   {"", ArbitraryArgs, true, []string{"one", "two"}},
		"Valid | Invalid": {"invalid", ArbitraryArgs, true, []string{"a"}},
	})
}
func TestArgs_MinimumN(t *testing.T) {
	testArgs(t, map[string]argsTestcase{
		"      | Arb":         {"", MinimumNArgs(2), false, []string{"a", "b", "c"}},
		"Valid | Valid":       {"", MinimumNArgs(2), true, []string{"one", "three"}},
		"Valid | Invalid":     {"invalid", MinimumNArgs(2), true, []string{"a", "b"}},
		"      | Less":        {"less", MinimumNArgs(2), false, []string{"a"}},
		"Valid | Less":        {"less", MinimumNArgs(2), true, []string{"one"}},
		"Valid | LessInvalid": {"invalid", MinimumNArgs(2), true, []string{"a"}},
	})
}
func TestArgs_MaximumN(t *testing.T) {
	testArgs(t, map[string]argsTestcase{
		"      | Arb":         {"", MaximumNArgs(3), false, []string{"a", "b"}},
		"Valid | Valid":       {"", MaximumNArgs(2), true, []string{"one", "three"}},
		"Valid | Invalid":     {"invalid", MaximumNArgs(2), true, []string{"a", "b"}},
		"      | More":        {"more", MaximumNArgs(2), false, []string{"a", "b", "c"}},
		"Valid | More":        {"more", MaximumNArgs(2), true, []string{"one", "three", "two"}},
		"Valid | MoreInvalid": {"invalid", MaximumNArgs(2), true, []string{"a", "b", "c"}},
	})
}
func TestArgs_Exact(t *testing.T) {
	testArgs(t, map[string]argsTestcase{
		"      | Arb":                 {"", ExactArgs(3), false, []string{"a", "b", "c"}},
		"Valid | Valid":               {"", ExactArgs(3), true, []string{"three", "one", "two"}},
		"Valid | Invalid":             {"invalid", ExactArgs(3), true, []string{"three", "a", "two"}},
		"      | InvalidCount":        {"notexact", ExactArgs(2), false, []string{"a", "b", "c"}},
		"Valid | InvalidCount":        {"notexact", ExactArgs(2), true, []string{"three", "one", "two"}},
		"Valid | InvalidCountInvalid": {"invalid", ExactArgs(2), true, []string{"three", "a", "two"}},
	})
}
func TestArgs_Range(t *testing.T) {
	testArgs(t, map[string]argsTestcase{
		"      | Arb":                 {"", RangeArgs(2, 4), false, []string{"a", "b", "c"}},
		"Valid | Valid":               {"", RangeArgs(2, 4), true, []string{"three", "one", "two"}},
		"Valid | Invalid":             {"invalid", RangeArgs(2, 4), true, []string{"three", "a", "two"}},
		"      | InvalidCount":        {"notinrange", RangeArgs(2, 4), false, []string{"a"}},
		"Valid | InvalidCount":        {"notinrange", RangeArgs(2, 4), true, []string{"two"}},
		"Valid | InvalidCountInvalid": {"invalid", RangeArgs(2, 4), true, []string{"a"}},
	})
}
func TestArgs_DEPRECATED(t *testing.T) {
	testArgs(t, map[string]argsTestcase{
		"OnlyValid  | Valid | Valid":        {"", OnlyValidArgs, true, []string{"one", "two"}},
		"OnlyValid  | Valid | Invalid":      {"invalid", OnlyValidArgs, true, []string{"a"}},
		"ExactValid | Valid | Valid":        {"", ExactValidArgs(3), true, []string{"two", "three", "one"}},
		"ExactValid | Valid | InvalidCount": {"notexact", ExactValidArgs(2), true, []string{"two", "three", "one"}},
		"ExactValid | Valid | Invalid":      {"invalid", ExactValidArgs(2), true, []string{"two", "a"}},
	})
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
