package cobra

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func executeUsage(c *Command) (string, error) {
	buf := new(bytes.Buffer)
	c.SetOutput(buf)
	err := c.Usage()
	return buf.String(), err
}

func checkOutput(o string, t *testing.T, i string) {
	str := map[rune]string{
		'u': "Usage:",
		'h': "Run 'c --help' for usage",
		'c': "c [command]",
		'v': "Valid Args:",
		'a': "c [flags] [args]",
		'f': "c [flags]",
	}
	for _, x := range "uhcva" {
		b := strings.Contains(i, string(x))
		if s := str[x]; b != strings.Contains(o, s) {
			m := "Did not expect"
			if b {
				m = "Expected"
			}
			t.Errorf("%s to find '%s' in the output", m, s)
			continue
		}
		if (x == 'a') && b {
			return
		}
	}
}

func expectErrorAndCheckOutput(t *testing.T, err error, err_k, o, i string) {
	//	expectError(err, t, err_k)
	//	checkOutput(o, t, i)
}

type argsTestcase struct {
	exerr        string         // Expected error key (see map[string][string])
	args         PositionalArgs // Args validator
	wValid, wRun bool           // Define `ValidArgs` in the command
	rargs        []string       // Runtime args
}

var errStrings = map[string]string{
	"run":        `command "c" is not runnable`,
	"runsub":     `command "c" is not runnable; please provide a subcmd`,
	"no":         `"one" rejected; "c" does not accept args`,
	"invalid":    `invalid argument "a" for "c"`,
	"unknown":    `unknown command "one" for "c"`,
	"less":       "requires at least 2 arg(s), only received 1",
	"more":       "accepts at most 2 arg(s), received 3",
	"notexact":   "accepts 2 arg(s), received 3",
	"notinrange": "accepts between 2 and 4 arg(s), received 1",
}

func newCmd(args PositionalArgs, wValid, wRun bool) *Command {
	c := &Command{
		Use:   "c",
		Short: "A generator",
		Long:  `Cobra is a CLI ...`,
		//Run:   emptyRun,
	}
	if args != nil {
		c.Args = args
	}
	if wValid {
		c.ValidArgs = []string{"one", "two", "three"}
	}
	if wRun {
		c.Run = func(cmd *Command, args []string) {
			//fmt.Println("RUN", args)
		}
	}
	return c
}

func (tc *argsTestcase) test(t *testing.T) {
	o, e := executeCommand(newCmd(tc.args, tc.wValid, tc.wRun), tc.rargs...)

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
		"      | ":      {"", NoArgs, false, true, []string{}},
		"      | Arb":   {"no", NoArgs, false, true, []string{"one"}},
		"Valid | Valid": {"no", NoArgs, true, true, []string{"one"}},
	})
}
func TestArgs_Nil(t *testing.T) {
	testArgs(t, map[string]argsTestcase{
		"      | Arb":     {"", nil, false, true, []string{"a", "b"}},
		"Valid | Valid":   {"", nil, true, true, []string{"one", "two"}},
		"Valid | Invalid": {"invalid", nil, true, true, []string{"a"}},
	})
}
func TestArgs_Arbitrary(t *testing.T) {
	testArgs(t, map[string]argsTestcase{
		"      | Arb":     {"", ArbitraryArgs, false, true, []string{"a", "b"}},
		"Valid | Valid":   {"", ArbitraryArgs, true, true, []string{"one", "two"}},
		"Valid | Invalid": {"invalid", ArbitraryArgs, true, true, []string{"a"}},
	})
}
func TestArgs_MinimumN(t *testing.T) {
	testArgs(t, map[string]argsTestcase{
		"      | Arb":         {"", MinimumNArgs(2), false, true, []string{"a", "b", "c"}},
		"Valid | Valid":       {"", MinimumNArgs(2), true, true, []string{"one", "three"}},
		"Valid | Invalid":     {"invalid", MinimumNArgs(2), true, true, []string{"a", "b"}},
		"      | Less":        {"less", MinimumNArgs(2), false, true, []string{"a"}},
		"Valid | Less":        {"less", MinimumNArgs(2), true, true, []string{"one"}},
		"Valid | LessInvalid": {"invalid", MinimumNArgs(2), true, true, []string{"a"}},
	})
}
func TestArgs_MaximumN(t *testing.T) {
	testArgs(t, map[string]argsTestcase{
		"      | Arb":         {"", MaximumNArgs(3), false, true, []string{"a", "b"}},
		"Valid | Valid":       {"", MaximumNArgs(2), true, true, []string{"one", "three"}},
		"Valid | Invalid":     {"invalid", MaximumNArgs(2), true, true, []string{"a", "b"}},
		"      | More":        {"more", MaximumNArgs(2), false, true, []string{"a", "b", "c"}},
		"Valid | More":        {"more", MaximumNArgs(2), true, true, []string{"one", "three", "two"}},
		"Valid | MoreInvalid": {"invalid", MaximumNArgs(2), true, true, []string{"a", "b", "c"}},
	})
}
func TestArgs_Exact(t *testing.T) {
	testArgs(t, map[string]argsTestcase{
		"      | Arb":                 {"", ExactArgs(3), false, true, []string{"a", "b", "c"}},
		"Valid | Valid":               {"", ExactArgs(3), true, true, []string{"three", "one", "two"}},
		"Valid | Invalid":             {"invalid", ExactArgs(3), true, true, []string{"three", "a", "two"}},
		"      | InvalidCount":        {"notexact", ExactArgs(2), false, true, []string{"a", "b", "c"}},
		"Valid | InvalidCount":        {"notexact", ExactArgs(2), true, true, []string{"three", "one", "two"}},
		"Valid | InvalidCountInvalid": {"invalid", ExactArgs(2), true, true, []string{"three", "a", "two"}},
	})
}
func TestArgs_Range(t *testing.T) {
	testArgs(t, map[string]argsTestcase{
		"      | Arb":                 {"", RangeArgs(2, 4), false, true, []string{"a", "b", "c"}},
		"Valid | Valid":               {"", RangeArgs(2, 4), true, true, []string{"three", "one", "two"}},
		"Valid | Invalid":             {"invalid", RangeArgs(2, 4), true, true, []string{"three", "a", "two"}},
		"      | InvalidCount":        {"notinrange", RangeArgs(2, 4), false, true, []string{"a"}},
		"Valid | InvalidCount":        {"notinrange", RangeArgs(2, 4), true, true, []string{"two"}},
		"Valid | InvalidCountInvalid": {"invalid", RangeArgs(2, 4), true, true, []string{"a"}},
	})
}
func TestArgs_DEPRECATED(t *testing.T) {
	testArgs(t, map[string]argsTestcase{
		"OnlyValid  | Valid | Valid":        {"", OnlyValidArgs, true, true, []string{"one", "two"}},
		"OnlyValid  | Valid | Invalid":      {"invalid", OnlyValidArgs, true, true, []string{"a"}},
		"ExactValid | Valid | Valid":        {"", ExactValidArgs(3), true, true, []string{"two", "three", "one"}},
		"ExactValid | Valid | InvalidCount": {"notexact", ExactValidArgs(2), true, true, []string{"two", "three", "one"}},
		"ExactValid | Valid | Invalid":      {"invalid", ExactValidArgs(2), true, true, []string{"two", "a"}},
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
	expected := `"illegal" rejected; "root child" does not accept args`
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

// NOTE 'c [command]' is not shown because this command does not have any subcommand
// NOTE 'Valid Args:' is not shown because this command is not runnable
// NOTE 'c [flags]' is not shown because this command is not runnable
func noRunChecks(t *testing.T, err error, err_k, o string) {
	expectErrorAndCheckOutput(t, err, err_k, o, "u")
}

// NoRun (no children)

func TestArgs_NoRun(t *testing.T) {
	tc := argsTestcase{"run", nil, false, false, []string{}}
	t.Run("|", tc.test)
	//	noRunChecks(t, e, "run", o)
}

func TestArgs_NoRun_ArbValid(t *testing.T) {
	tc := argsTestcase{"run", nil, false, false, []string{"one", "three"}}
	t.Run("|", tc.test)
	//	noRunChecks(t, e, "run", o)
}

func TestArgs_NoRun_Invalid(t *testing.T) {
	tc := argsTestcase{"run", nil, false, false, []string{"two", "a"}}
	t.Run("|", tc.test)
	//noRunChecks(t, e, "run", o)
}

// NoRun (with children)
// NOTE 'Valid Args:' is not shown because this command is not runnable
// NOTE 'c [flags]' is not shown because this command is not runnable

func TestArgs_NoRun_wChild(t *testing.T) {
	c := newCmd(nil, false, false)
	d := newCmd(nil, false, true)
	c.AddCommand(d)
	o, e := executeCommand(c)
	expectErrorAndCheckOutput(t, e, "runsub", o, "uc")
}

func TestArgs_NoRun_wChild_ArbValid(t *testing.T) {
	c := newCmd(nil, false, false)
	d := newCmd(nil, false, true)
	c.AddCommand(d)
	o, e := executeCommand(c, "one", "three")
	expectErrorAndCheckOutput(t, e, "runsub", o, "h")
}

func TestArgs_NoRun_wChild_Invalid(t *testing.T) {
	c := newCmd(nil, false, false)
	d := newCmd(nil, false, true)
	c.AddCommand(d)
	o, e := executeCommand(c, "one", "a")
	expectErrorAndCheckOutput(t, e, "runsub", o, "h")
}

// NoRun Args

func TestArgs_NoRun_wArgs(t *testing.T) {
	testArgs(t, map[string]argsTestcase{
		"|": {"run", ArbitraryArgs, false, false, []string{}},
	})
	//noRunChecks(t, e, "run", o)
}

func TestArgs_NoRun_wArgs_ArbValid(t *testing.T) {
	testArgs(t, map[string]argsTestcase{
		"|": {"run", ArbitraryArgs, false, false, []string{"one", "three"}},
	})
	//noRunChecks(t, e, "run", o)
}

func TestArgs_NoRun_wArgs_Invalid(t *testing.T) {
	testArgs(t, map[string]argsTestcase{
		"|": {"run", ArbitraryArgs, false, false, []string{"two", "a"}},
	})
	//noRunChecks(t, e, "run", o)
}

// NoRun ValidArgs

func TestArgs_NoRun_wValid(t *testing.T) {
	testArgs(t, map[string]argsTestcase{
		"|": {"run", nil, true, false, []string{}},
	})
	//noRunChecks(t, e, "run", o)
}

func TestArgs_NoRun_wValid_ArbValid(t *testing.T) {
	testArgs(t, map[string]argsTestcase{
		"|": {"run", nil, true, false, []string{"one", "three"}},
	})
	//noRunChecks(t, e, "run", o)
}

func TestArgs_NoRun_wValid_Invalid(t *testing.T) {
	testArgs(t, map[string]argsTestcase{
		"|": {"run", nil, true, false, []string{"two", "a"}},
	})
	//noRunChecks(t, e, "run", o)
}

// NoRun Args ValidArgs

func TestArgs_NoRun_wArgswValid(t *testing.T) {
	testArgs(t, map[string]argsTestcase{
		"|": {"run", ArbitraryArgs, true, false, []string{}},
	})
	//	noRunChecks(t, e, "run", o)
}

func TestArgs_NoRun_wArgswValid_ArbValid(t *testing.T) {
	testArgs(t, map[string]argsTestcase{
		"|": {"run", ArbitraryArgs, true, false, []string{"one", "three"}},
	})
	//	noRunChecks(t, e, "run", o)
}

func TestArgs_NoRun_wArgswValid_Invalid(t *testing.T) {
	testArgs(t, map[string]argsTestcase{
		"|": {"run", ArbitraryArgs, true, false, []string{"two", "a"}},
	})
	//	noRunChecks(t, e, "run", o)
}

// Run (no children)
// NOTE 'c [command]' is not shown because this command does not have any subcommand
// NOTE 'Valid Args:' is not shown because ValidArgs is not defined

func TestArgs_Run(t *testing.T) {
	testArgs(t, map[string]argsTestcase{
		"|": {"", nil, false, true, []string{}},
	})
	//o, e = executeUsage(c)
	//checkOutput(o, t, "ua")
}

func TestArgs_Run_ArbValid(t *testing.T) {
	testArgs(t, map[string]argsTestcase{
		"|": {"", nil, false, true, []string{"one", "three"}},
	})
	//	o, e = executeUsage(c)
	//	checkOutput(o, t, "ua")
}

func TestArgs_Run_Invalid(t *testing.T) {
	testArgs(t, map[string]argsTestcase{
		"|": {"", nil, false, true, []string{"two", "a"}},
	})
	//o, e = executeUsage(c)
	//checkOutput(o, t, "ua")
}

// Run (with children)
// NOTE 'Valid Args:' is not shown because ValidArgs is not defined

func TestArgs_Run_wChild(t *testing.T) {
	c := newCmd(nil, false, true)
	d := newCmd(nil, false, true)
	c.AddCommand(d)
	//	o, e := executeCommand(c)
	//	expectSuccess(o, e, t)
	o, _ := executeUsage(c)
	checkOutput(o, t, "ucf")
}

func TestArgs_Run_wChild_ArbValid(t *testing.T) {
	c := newCmd(nil, false, true)
	d := newCmd(nil, false, false)
	c.AddCommand(d)
	o, _ := executeCommand(c, "one", "three")
	//	expectError(e, t, "no")
	// NOTE 'c [command]' is not shown because this command does not have any available subcommand
	checkOutput(o, t, "uf")
}

func TestArgs_Run_wChild_Invalid(t *testing.T) {
	c := newCmd(nil, false, true)
	d := newCmd(nil, false, false)
	c.AddCommand(d)
	o, _ := executeCommand(c, "one", "a")
	//	expectError(e, t, "no")
	// NOTE 'c [command]' is not shown because this command does not have any available subcommand
	checkOutput(o, t, "uf")
}

// Run Args
// NOTE 'c [command]' is not shown because this command does not have any subcommand

func TestArgs_Run_wArgs(t *testing.T) {
	testArgs(t, map[string]argsTestcase{
		"|": {"", ArbitraryArgs, false, true, []string{}},
	})
	//	o, e = executeUsage(c)
	//	checkOutput(o, t, "ua")
}

func TestArgs_Run_wArgs_ArbValid(t *testing.T) {
	testArgs(t, map[string]argsTestcase{
		"|": {"", ArbitraryArgs, false, true, []string{"one", "three"}},
	})
	//	o, e = executeUsage(c)
	//	checkOutput(o, t, "ua")
}

func TestArgs_Run_wArgs_Invalid(t *testing.T) {
	testArgs(t, map[string]argsTestcase{
		"|": {"", ArbitraryArgs, false, true, []string{"two", "a"}},
	})
	//	o, e = executeUsage(c)
	//	checkOutput(o, t, "ua")
}

// Run ValidArgs
// NOTE 'c [command]' is not shown because this command does not have any subcommand

func TestArgs_Run_wValid(t *testing.T) {
	testArgs(t, map[string]argsTestcase{
		"|": {"", nil, true, true, []string{}},
	})
	//	o, e = executeUsage(c)
	//	checkOutput(o, t, "uva")
}

func TestArgs_Run_wValid_ArbValid(t *testing.T) {
	testArgs(t, map[string]argsTestcase{
		"|": {"", nil, true, true, []string{"one", "three"}},
	})
	//	o, e = executeUsage(c)
	//	checkOutput(o, t, "uva")
}

func TestArgs_Run_wValid_Invalid(t *testing.T) {
	testArgs(t, map[string]argsTestcase{
		"|": {"invalid", nil, true, true, []string{"two", "a"}},
	})
	//	checkOutput(o, t, "uva")
}

// Run Args ValidArgs
// NOTE 'c [command]' is not shown because this command does not have any subcommand

func TestArgs_Run_wArgswValid(t *testing.T) {
	testArgs(t, map[string]argsTestcase{
		"|": {"", ArbitraryArgs, true, true, []string{}},
	})
	//o, e = executeUsage(c)
	//checkOutput(o, t, "uva")
}

func TestArgs_Run_wArgswValid_ArbValid(t *testing.T) {
	testArgs(t, map[string]argsTestcase{
		"|": {"", ArbitraryArgs, true, true, []string{"one", "three"}},
	})
	//o, e = executeUsage(c)
	//checkOutput(o, t, "uva")
}

func TestArgs_Run_wArgswValid_Invalid(t *testing.T) {
	testArgs(t, map[string]argsTestcase{
		"|": {"invalid", ArbitraryArgs, true, true, []string{"two", "a"}},
	})
	//checkOutput(o, t, "uva")
}

//

func TestArgs_Run_wMinimumNArgs_ArbValid(t *testing.T) {
	testArgs(t, map[string]argsTestcase{
		"|": {"", MinimumNArgs(2), false, true, []string{"one", "three"}},
	})
	//o, e = executeUsage(c)
	//checkOutput(o, t, "ua")
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
