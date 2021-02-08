package cobra

import (
	"bytes"
	"strings"
	"testing"
)

func validArgsFunc(cmd *Command, args []string, toComplete string) ([]string, ShellCompDirective) {
	if len(args) != 0 {
		return nil, ShellCompDirectiveNoFileComp
	}

	var completions []string
	for _, comp := range []string{"one\tThe first", "two\tThe second"} {
		if strings.HasPrefix(comp, toComplete) {
			completions = append(completions, comp)
		}
	}
	return completions, ShellCompDirectiveDefault
}

func validArgsFunc2(cmd *Command, args []string, toComplete string) ([]string, ShellCompDirective) {
	if len(args) != 0 {
		return nil, ShellCompDirectiveNoFileComp
	}

	var completions []string
	for _, comp := range []string{"three\tThe third", "four\tThe fourth"} {
		if strings.HasPrefix(comp, toComplete) {
			completions = append(completions, comp)
		}
	}
	return completions, ShellCompDirectiveDefault
}

func TestCmdNameCompletionInGo(t *testing.T) {
	rootCmd := &Command{
		Use: "root",
		Run: emptyRun,
	}
	childCmd1 := &Command{
		Use:   "firstChild",
		Short: "First command",
		Run:   emptyRun,
	}
	childCmd2 := &Command{
		Use: "secondChild",
		Run: emptyRun,
	}
	hiddenCmd := &Command{
		Use:    "testHidden",
		Hidden: true, // Not completed
		Run:    emptyRun,
	}
	deprecatedCmd := &Command{
		Use:        "testDeprecated",
		Deprecated: "deprecated", // Not completed
		Run:        emptyRun,
	}
	aliasedCmd := &Command{
		Use:     "aliased",
		Short:   "A command with aliases",
		Aliases: []string{"testAlias", "testSynonym"}, // Not completed
		Run:     emptyRun,
	}

	rootCmd.AddCommand(childCmd1, childCmd2, hiddenCmd, deprecatedCmd, aliasedCmd)

	// Test that sub-command names are completed
	output, err := executeCommand(rootCmd, ShellCompNoDescRequestCmd, "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := strings.Join([]string{
		"aliased",
		"firstChild",
		"help",
		"secondChild",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that sub-command names are completed with prefix
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "s")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"secondChild",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that even with no valid sub-command matches, hidden, deprecated and
	// aliases are not completed
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "test")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that sub-command names are completed with description
	output, err = executeCommand(rootCmd, ShellCompRequestCmd, "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"aliased\tA command with aliases",
		"firstChild\tFirst command",
		"help\tHelp about any command",
		"secondChild",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}
}

func TestNoCmdNameCompletionInGo(t *testing.T) {
	rootCmd := &Command{
		Use: "root",
		Run: emptyRun,
	}
	rootCmd.Flags().String("localroot", "", "local root flag")

	childCmd1 := &Command{
		Use:   "childCmd1",
		Short: "First command",
		Args:  MinimumNArgs(0),
		Run:   emptyRun,
	}
	rootCmd.AddCommand(childCmd1)
	childCmd1.PersistentFlags().StringP("persistent", "p", "", "persistent flag")
	persistentFlag := childCmd1.PersistentFlags().Lookup("persistent")
	childCmd1.Flags().StringP("nonPersistent", "n", "", "non-persistent flag")
	nonPersistentFlag := childCmd1.Flags().Lookup("nonPersistent")

	childCmd2 := &Command{
		Use: "childCmd2",
		Run: emptyRun,
	}
	childCmd1.AddCommand(childCmd2)

	// Test that sub-command names are not completed if there is an argument already
	output, err := executeCommand(rootCmd, ShellCompNoDescRequestCmd, "childCmd1", "arg1", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := strings.Join([]string{
		":0",
		"Completion ended with directive: ShellCompDirectiveDefault", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that sub-command names are not completed if a local non-persistent flag is present
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "childCmd1", "--nonPersistent", "value", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	// Reset the flag for the next command
	nonPersistentFlag.Changed = false

	expected = strings.Join([]string{
		":0",
		"Completion ended with directive: ShellCompDirectiveDefault", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that sub-command names are completed if a local non-persistent flag is present and TraverseChildren is set to true
	// set TraverseChildren to true on the root cmd
	rootCmd.TraverseChildren = true

	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "--localroot", "value", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	// Reset TraverseChildren for next command
	rootCmd.TraverseChildren = false

	expected = strings.Join([]string{
		"childCmd1",
		"help",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that sub-command names from a child cmd are completed if a local non-persistent flag is present
	// and TraverseChildren is set to true on the root cmd
	rootCmd.TraverseChildren = true

	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "--localroot", "value", "childCmd1", "--nonPersistent", "value", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	// Reset TraverseChildren for next command
	rootCmd.TraverseChildren = false
	// Reset the flag for the next command
	nonPersistentFlag.Changed = false

	expected = strings.Join([]string{
		"childCmd2",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that we don't use Traverse when we shouldn't.
	// This command should not return a completion since the command line is invalid without TraverseChildren.
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "--localroot", "value", "childCmd1", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		":0",
		"Completion ended with directive: ShellCompDirectiveDefault", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that sub-command names are not completed if a local non-persistent short flag is present
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "childCmd1", "-n", "value", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	// Reset the flag for the next command
	nonPersistentFlag.Changed = false

	expected = strings.Join([]string{
		":0",
		"Completion ended with directive: ShellCompDirectiveDefault", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that sub-command names are completed with a persistent flag
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "childCmd1", "--persistent", "value", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	// Reset the flag for the next command
	persistentFlag.Changed = false

	expected = strings.Join([]string{
		"childCmd2",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that sub-command names are completed with a persistent short flag
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "childCmd1", "-p", "value", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	// Reset the flag for the next command
	persistentFlag.Changed = false

	expected = strings.Join([]string{
		"childCmd2",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}
}

func TestValidArgsCompletionInGo(t *testing.T) {
	rootCmd := &Command{
		Use:       "root",
		ValidArgs: []string{"one", "two", "three"},
		Args:      MinimumNArgs(1),
	}

	// Test that validArgs are completed
	output, err := executeCommand(rootCmd, ShellCompNoDescRequestCmd, "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := strings.Join([]string{
		"one",
		"two",
		"three",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that validArgs are completed with prefix
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "o")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"one",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that validArgs don't repeat
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "one", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		":0",
		"Completion ended with directive: ShellCompDirectiveDefault", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}
}

func TestValidArgsAndCmdCompletionInGo(t *testing.T) {
	rootCmd := &Command{
		Use:       "root",
		ValidArgs: []string{"one", "two"},
		Run:       emptyRun,
	}

	childCmd := &Command{
		Use: "thechild",
		Run: emptyRun,
	}

	rootCmd.AddCommand(childCmd)

	// Test that both sub-commands and validArgs are completed
	output, err := executeCommand(rootCmd, ShellCompNoDescRequestCmd, "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := strings.Join([]string{
		"help",
		"thechild",
		"one",
		"two",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that both sub-commands and validArgs are completed with prefix
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "t")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"thechild",
		"two",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}
}

func TestValidArgsFuncAndCmdCompletionInGo(t *testing.T) {
	rootCmd := &Command{
		Use:               "root",
		ValidArgsFunction: validArgsFunc,
		Run:               emptyRun,
	}

	childCmd := &Command{
		Use:   "thechild",
		Short: "The child command",
		Run:   emptyRun,
	}

	rootCmd.AddCommand(childCmd)

	// Test that both sub-commands and validArgsFunction are completed
	output, err := executeCommand(rootCmd, ShellCompNoDescRequestCmd, "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := strings.Join([]string{
		"help",
		"thechild",
		"one",
		"two",
		":0",
		"Completion ended with directive: ShellCompDirectiveDefault", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that both sub-commands and validArgs are completed with prefix
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "t")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"thechild",
		"two",
		":0",
		"Completion ended with directive: ShellCompDirectiveDefault", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that both sub-commands and validArgs are completed with description
	output, err = executeCommand(rootCmd, ShellCompRequestCmd, "t")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"thechild\tThe child command",
		"two\tThe second",
		":0",
		"Completion ended with directive: ShellCompDirectiveDefault", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}
}

func TestFlagNameCompletionInGo(t *testing.T) {
	rootCmd := &Command{
		Use: "root",
		Run: emptyRun,
	}
	childCmd := &Command{
		Use: "childCmd",
		Run: emptyRun,
	}
	rootCmd.AddCommand(childCmd)

	rootCmd.Flags().IntP("first", "f", -1, "first flag")
	rootCmd.PersistentFlags().BoolP("second", "s", false, "second flag")
	childCmd.Flags().String("subFlag", "", "sub flag")

	// Test that flag names are not shown if the user has not given the '-' prefix
	output, err := executeCommand(rootCmd, ShellCompNoDescRequestCmd, "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := strings.Join([]string{
		"childCmd",
		"help",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that flag names are completed
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "-")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"--first",
		"-f",
		"--second",
		"-s",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that flag names are completed when a prefix is given
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "--f")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"--first",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that flag names are completed in a sub-cmd
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "childCmd", "-")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"--second",
		"-s",
		"--subFlag",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}
}

func TestFlagNameCompletionInGoWithDesc(t *testing.T) {
	rootCmd := &Command{
		Use: "root",
		Run: emptyRun,
	}
	childCmd := &Command{
		Use:   "childCmd",
		Short: "first command",
		Run:   emptyRun,
	}
	rootCmd.AddCommand(childCmd)

	rootCmd.Flags().IntP("first", "f", -1, "first flag\nlonger description for flag")
	rootCmd.PersistentFlags().BoolP("second", "s", false, "second flag")
	childCmd.Flags().String("subFlag", "", "sub flag")

	// Test that flag names are not shown if the user has not given the '-' prefix
	output, err := executeCommand(rootCmd, ShellCompRequestCmd, "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := strings.Join([]string{
		"childCmd\tfirst command",
		"help\tHelp about any command",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that flag names are completed
	output, err = executeCommand(rootCmd, ShellCompRequestCmd, "-")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"--first\tfirst flag",
		"-f\tfirst flag",
		"--second\tsecond flag",
		"-s\tsecond flag",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that flag names are completed when a prefix is given
	output, err = executeCommand(rootCmd, ShellCompRequestCmd, "--f")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"--first\tfirst flag",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that flag names are completed in a sub-cmd
	output, err = executeCommand(rootCmd, ShellCompRequestCmd, "childCmd", "-")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"--second\tsecond flag",
		"-s\tsecond flag",
		"--subFlag\tsub flag",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}
}

func TestFlagNameCompletionRepeat(t *testing.T) {
	rootCmd := &Command{
		Use: "root",
		Run: emptyRun,
	}
	childCmd := &Command{
		Use:   "childCmd",
		Short: "first command",
		Run:   emptyRun,
	}
	rootCmd.AddCommand(childCmd)

	rootCmd.Flags().IntP("first", "f", -1, "first flag")
	firstFlag := rootCmd.Flags().Lookup("first")
	rootCmd.Flags().BoolP("second", "s", false, "second flag")
	secondFlag := rootCmd.Flags().Lookup("second")
	rootCmd.Flags().StringArrayP("array", "a", nil, "array flag")
	arrayFlag := rootCmd.Flags().Lookup("array")
	rootCmd.Flags().IntSliceP("slice", "l", nil, "slice flag")
	sliceFlag := rootCmd.Flags().Lookup("slice")
	rootCmd.Flags().BoolSliceP("bslice", "b", nil, "bool slice flag")
	bsliceFlag := rootCmd.Flags().Lookup("bslice")

	// Test that flag names are not repeated unless they are an array or slice
	output, err := executeCommand(rootCmd, ShellCompNoDescRequestCmd, "--first", "1", "--")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	// Reset the flag for the next command
	firstFlag.Changed = false

	expected := strings.Join([]string{
		"--array",
		"--bslice",
		"--second",
		"--slice",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that flag names are not repeated unless they are an array or slice
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "--first", "1", "--second=false", "--")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	// Reset the flag for the next command
	firstFlag.Changed = false
	secondFlag.Changed = false

	expected = strings.Join([]string{
		"--array",
		"--bslice",
		"--slice",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that flag names are not repeated unless they are an array or slice
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "--slice", "1", "--slice=2", "--array", "val", "--bslice", "true", "--")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	// Reset the flag for the next command
	sliceFlag.Changed = false
	arrayFlag.Changed = false
	bsliceFlag.Changed = false

	expected = strings.Join([]string{
		"--array",
		"--bslice",
		"--first",
		"--second",
		"--slice",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that flag names are not repeated unless they are an array or slice, using shortname
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "-l", "1", "-l=2", "-a", "val", "-")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	// Reset the flag for the next command
	sliceFlag.Changed = false
	arrayFlag.Changed = false

	expected = strings.Join([]string{
		"--array",
		"-a",
		"--bslice",
		"-b",
		"--first",
		"-f",
		"--second",
		"-s",
		"--slice",
		"-l",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that flag names are not repeated unless they are an array or slice, using shortname with prefix
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "-l", "1", "-l=2", "-a", "val", "-a")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	// Reset the flag for the next command
	sliceFlag.Changed = false
	arrayFlag.Changed = false

	expected = strings.Join([]string{
		"-a",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}
}

func TestRequiredFlagNameCompletionInGo(t *testing.T) {
	rootCmd := &Command{
		Use:       "root",
		ValidArgs: []string{"realArg"},
		Run:       emptyRun,
	}
	childCmd := &Command{
		Use: "childCmd",
		ValidArgsFunction: func(cmd *Command, args []string, toComplete string) ([]string, ShellCompDirective) {
			return []string{"subArg"}, ShellCompDirectiveNoFileComp
		},
		Run: emptyRun,
	}
	rootCmd.AddCommand(childCmd)

	rootCmd.Flags().IntP("requiredFlag", "r", -1, "required flag")
	assertNoErr(t, rootCmd.MarkFlagRequired("requiredFlag"))
	requiredFlag := rootCmd.Flags().Lookup("requiredFlag")

	rootCmd.PersistentFlags().IntP("requiredPersistent", "p", -1, "required persistent")
	assertNoErr(t, rootCmd.MarkPersistentFlagRequired("requiredPersistent"))
	requiredPersistent := rootCmd.PersistentFlags().Lookup("requiredPersistent")

	rootCmd.Flags().StringP("release", "R", "", "Release name")

	childCmd.Flags().BoolP("subRequired", "s", false, "sub required flag")
	assertNoErr(t, childCmd.MarkFlagRequired("subRequired"))
	childCmd.Flags().BoolP("subNotRequired", "n", false, "sub not required flag")

	// Test that a required flag is suggested even without the - prefix
	output, err := executeCommand(rootCmd, ShellCompNoDescRequestCmd, "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := strings.Join([]string{
		"childCmd",
		"help",
		"--requiredFlag",
		"-r",
		"--requiredPersistent",
		"-p",
		"realArg",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that a required flag is suggested without other flags when using the '-' prefix
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "-")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"--requiredFlag",
		"-r",
		"--requiredPersistent",
		"-p",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that if no required flag matches, the normal flags are suggested
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "--relea")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"--release",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test required flags for sub-commands
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "childCmd", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"--requiredPersistent",
		"-p",
		"--subRequired",
		"-s",
		"subArg",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "childCmd", "-")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"--requiredPersistent",
		"-p",
		"--subRequired",
		"-s",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "childCmd", "--subNot")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"--subNotRequired",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that when a required flag is present, it is not suggested anymore
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "--requiredFlag", "1", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	// Reset the flag for the next command
	requiredFlag.Changed = false

	expected = strings.Join([]string{
		"--requiredPersistent",
		"-p",
		"realArg",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that when a persistent required flag is present, it is not suggested anymore
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "--requiredPersistent", "1", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	// Reset the flag for the next command
	requiredPersistent.Changed = false

	expected = strings.Join([]string{
		"childCmd",
		"help",
		"--requiredFlag",
		"-r",
		"realArg",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that when all required flags are present, normal completion is done
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "--requiredFlag", "1", "--requiredPersistent", "1", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	// Reset the flags for the next command
	requiredFlag.Changed = false
	requiredPersistent.Changed = false

	expected = strings.Join([]string{
		"realArg",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}
}

func TestFlagFileExtFilterCompletionInGo(t *testing.T) {
	rootCmd := &Command{
		Use: "root",
		Run: emptyRun,
	}

	// No extensions.  Should be ignored.
	rootCmd.Flags().StringP("file", "f", "", "file flag")
	assertNoErr(t, rootCmd.MarkFlagFilename("file"))

	// Single extension
	rootCmd.Flags().StringP("log", "l", "", "log flag")
	assertNoErr(t, rootCmd.MarkFlagFilename("log", "log"))

	// Multiple extensions
	rootCmd.Flags().StringP("yaml", "y", "", "yaml flag")
	assertNoErr(t, rootCmd.MarkFlagFilename("yaml", "yaml", "yml"))

	// Directly using annotation
	rootCmd.Flags().StringP("text", "t", "", "text flag")
	assertNoErr(t, rootCmd.Flags().SetAnnotation("text", BashCompFilenameExt, []string{"txt"}))

	// Test that the completion logic returns the proper info for the completion
	// script to handle the file filtering
	output, err := executeCommand(rootCmd, ShellCompNoDescRequestCmd, "--file", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := strings.Join([]string{
		":0",
		"Completion ended with directive: ShellCompDirectiveDefault", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "--log", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"log",
		":8",
		"Completion ended with directive: ShellCompDirectiveFilterFileExt", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "--yaml", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"yaml", "yml",
		":8",
		"Completion ended with directive: ShellCompDirectiveFilterFileExt", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "--yaml=")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"yaml", "yml",
		":8",
		"Completion ended with directive: ShellCompDirectiveFilterFileExt", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "-y", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"yaml", "yml",
		":8",
		"Completion ended with directive: ShellCompDirectiveFilterFileExt", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "-y=")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"yaml", "yml",
		":8",
		"Completion ended with directive: ShellCompDirectiveFilterFileExt", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "--text", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"txt",
		":8",
		"Completion ended with directive: ShellCompDirectiveFilterFileExt", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}
}

func TestFlagDirFilterCompletionInGo(t *testing.T) {
	rootCmd := &Command{
		Use: "root",
		Run: emptyRun,
	}

	// Filter directories
	rootCmd.Flags().StringP("dir", "d", "", "dir flag")
	assertNoErr(t, rootCmd.MarkFlagDirname("dir"))

	// Filter directories within a directory
	rootCmd.Flags().StringP("subdir", "s", "", "subdir")
	assertNoErr(t, rootCmd.Flags().SetAnnotation("subdir", BashCompSubdirsInDir, []string{"themes"}))

	// Multiple directory specification get ignored
	rootCmd.Flags().StringP("manydir", "m", "", "manydir")
	assertNoErr(t, rootCmd.Flags().SetAnnotation("manydir", BashCompSubdirsInDir, []string{"themes", "colors"}))

	// Test that the completion logic returns the proper info for the completion
	// script to handle the directory filtering
	output, err := executeCommand(rootCmd, ShellCompNoDescRequestCmd, "--dir", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := strings.Join([]string{
		":16",
		"Completion ended with directive: ShellCompDirectiveFilterDirs", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "-d", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		":16",
		"Completion ended with directive: ShellCompDirectiveFilterDirs", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "--subdir", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"themes",
		":16",
		"Completion ended with directive: ShellCompDirectiveFilterDirs", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "--subdir=")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"themes",
		":16",
		"Completion ended with directive: ShellCompDirectiveFilterDirs", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "-s", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"themes",
		":16",
		"Completion ended with directive: ShellCompDirectiveFilterDirs", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "-s=")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"themes",
		":16",
		"Completion ended with directive: ShellCompDirectiveFilterDirs", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "--manydir", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		":16",
		"Completion ended with directive: ShellCompDirectiveFilterDirs", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}
}

func TestValidArgsFuncSingleCmd(t *testing.T) {
	rootCmd := &Command{
		Use:               "root",
		ValidArgsFunction: validArgsFunc,
		Run:               emptyRun,
	}

	// Test completing an empty string
	output, err := executeCommand(rootCmd, ShellCompNoDescRequestCmd, "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := strings.Join([]string{
		"one",
		"two",
		":0",
		"Completion ended with directive: ShellCompDirectiveDefault", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Check completing with a prefix
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "t")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"two",
		":0",
		"Completion ended with directive: ShellCompDirectiveDefault", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}
}

func TestValidArgsFuncSingleCmdInvalidArg(t *testing.T) {
	rootCmd := &Command{
		Use: "root",
		// If we don't specify a value for Args, this test fails.
		// This is only true for a root command without any subcommands, and is caused
		// by the fact that the __complete command becomes a subcommand when there should not be one.
		// The problem is in the implementation of legacyArgs().
		Args:              MinimumNArgs(1),
		ValidArgsFunction: validArgsFunc,
		Run:               emptyRun,
	}

	// Check completing with wrong number of args
	output, err := executeCommand(rootCmd, ShellCompNoDescRequestCmd, "unexpectedArg", "t")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := strings.Join([]string{
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}
}

func TestValidArgsFuncChildCmds(t *testing.T) {
	rootCmd := &Command{Use: "root", Args: NoArgs, Run: emptyRun}
	child1Cmd := &Command{
		Use:               "child1",
		ValidArgsFunction: validArgsFunc,
		Run:               emptyRun,
	}
	child2Cmd := &Command{
		Use:               "child2",
		ValidArgsFunction: validArgsFunc2,
		Run:               emptyRun,
	}
	rootCmd.AddCommand(child1Cmd, child2Cmd)

	// Test completion of first sub-command with empty argument
	output, err := executeCommand(rootCmd, ShellCompNoDescRequestCmd, "child1", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := strings.Join([]string{
		"one",
		"two",
		":0",
		"Completion ended with directive: ShellCompDirectiveDefault", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test completion of first sub-command with a prefix to complete
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "child1", "t")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"two",
		":0",
		"Completion ended with directive: ShellCompDirectiveDefault", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Check completing with wrong number of args
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "child1", "unexpectedArg", "t")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test completion of second sub-command with empty argument
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "child2", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"three",
		"four",
		":0",
		"Completion ended with directive: ShellCompDirectiveDefault", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "child2", "t")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"three",
		":0",
		"Completion ended with directive: ShellCompDirectiveDefault", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Check completing with wrong number of args
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "child2", "unexpectedArg", "t")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}
}

func TestValidArgsFuncAliases(t *testing.T) {
	rootCmd := &Command{Use: "root", Args: NoArgs, Run: emptyRun}
	child := &Command{
		Use:               "child",
		Aliases:           []string{"son", "daughter"},
		ValidArgsFunction: validArgsFunc,
		Run:               emptyRun,
	}
	rootCmd.AddCommand(child)

	// Test completion of first sub-command with empty argument
	output, err := executeCommand(rootCmd, ShellCompNoDescRequestCmd, "son", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := strings.Join([]string{
		"one",
		"two",
		":0",
		"Completion ended with directive: ShellCompDirectiveDefault", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test completion of first sub-command with a prefix to complete
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "daughter", "t")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"two",
		":0",
		"Completion ended with directive: ShellCompDirectiveDefault", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Check completing with wrong number of args
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "son", "unexpectedArg", "t")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}
}

func TestValidArgsFuncInBashScript(t *testing.T) {
	rootCmd := &Command{Use: "root", Args: NoArgs, Run: emptyRun}
	child := &Command{
		Use:               "child",
		ValidArgsFunction: validArgsFunc,
		Run:               emptyRun,
	}
	rootCmd.AddCommand(child)

	buf := new(bytes.Buffer)
	assertNoErr(t, rootCmd.GenBashCompletion(buf))
	output := buf.String()

	check(t, output, "has_completion_function=1")
}

func TestNoValidArgsFuncInBashScript(t *testing.T) {
	rootCmd := &Command{Use: "root", Args: NoArgs, Run: emptyRun}
	child := &Command{
		Use: "child",
		Run: emptyRun,
	}
	rootCmd.AddCommand(child)

	buf := new(bytes.Buffer)
	assertNoErr(t, rootCmd.GenBashCompletion(buf))
	output := buf.String()

	checkOmit(t, output, "has_completion_function=1")
}

func TestCompleteCmdInBashScript(t *testing.T) {
	rootCmd := &Command{Use: "root", Args: NoArgs, Run: emptyRun}
	child := &Command{
		Use:               "child",
		ValidArgsFunction: validArgsFunc,
		Run:               emptyRun,
	}
	rootCmd.AddCommand(child)

	buf := new(bytes.Buffer)
	assertNoErr(t, rootCmd.GenBashCompletion(buf))
	output := buf.String()

	check(t, output, ShellCompNoDescRequestCmd)
}

func TestCompleteNoDesCmdInZshScript(t *testing.T) {
	rootCmd := &Command{Use: "root", Args: NoArgs, Run: emptyRun}
	child := &Command{
		Use:               "child",
		ValidArgsFunction: validArgsFunc,
		Run:               emptyRun,
	}
	rootCmd.AddCommand(child)

	buf := new(bytes.Buffer)
	assertNoErr(t, rootCmd.GenZshCompletionNoDesc(buf))
	output := buf.String()

	check(t, output, ShellCompNoDescRequestCmd)
}

func TestCompleteCmdInZshScript(t *testing.T) {
	rootCmd := &Command{Use: "root", Args: NoArgs, Run: emptyRun}
	child := &Command{
		Use:               "child",
		ValidArgsFunction: validArgsFunc,
		Run:               emptyRun,
	}
	rootCmd.AddCommand(child)

	buf := new(bytes.Buffer)
	assertNoErr(t, rootCmd.GenZshCompletion(buf))
	output := buf.String()

	check(t, output, ShellCompRequestCmd)
	checkOmit(t, output, ShellCompNoDescRequestCmd)
}

func TestFlagCompletionInGo(t *testing.T) {
	rootCmd := &Command{
		Use: "root",
		Run: emptyRun,
	}
	rootCmd.Flags().IntP("introot", "i", -1, "help message for flag introot")
	assertNoErr(t, rootCmd.RegisterFlagCompletionFunc("introot", func(cmd *Command, args []string, toComplete string) ([]string, ShellCompDirective) {
		completions := []string{}
		for _, comp := range []string{"1\tThe first", "2\tThe second", "10\tThe tenth"} {
			if strings.HasPrefix(comp, toComplete) {
				completions = append(completions, comp)
			}
		}
		return completions, ShellCompDirectiveDefault
	}))
	rootCmd.Flags().String("filename", "", "Enter a filename")
	assertNoErr(t, rootCmd.RegisterFlagCompletionFunc("filename", func(cmd *Command, args []string, toComplete string) ([]string, ShellCompDirective) {
		completions := []string{}
		for _, comp := range []string{"file.yaml\tYAML format", "myfile.json\tJSON format", "file.xml\tXML format"} {
			if strings.HasPrefix(comp, toComplete) {
				completions = append(completions, comp)
			}
		}
		return completions, ShellCompDirectiveNoSpace | ShellCompDirectiveNoFileComp
	}))

	// Test completing an empty string
	output, err := executeCommand(rootCmd, ShellCompNoDescRequestCmd, "--introot", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := strings.Join([]string{
		"1",
		"2",
		"10",
		":0",
		"Completion ended with directive: ShellCompDirectiveDefault", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Check completing with a prefix
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "--introot", "1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"1",
		"10",
		":0",
		"Completion ended with directive: ShellCompDirectiveDefault", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test completing an empty string
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "--filename", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"file.yaml",
		"myfile.json",
		"file.xml",
		":6",
		"Completion ended with directive: ShellCompDirectiveNoSpace, ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Check completing with a prefix
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "--filename", "f")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"file.yaml",
		"file.xml",
		":6",
		"Completion ended with directive: ShellCompDirectiveNoSpace, ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}
}

func TestValidArgsFuncChildCmdsWithDesc(t *testing.T) {
	rootCmd := &Command{Use: "root", Args: NoArgs, Run: emptyRun}
	child1Cmd := &Command{
		Use:               "child1",
		ValidArgsFunction: validArgsFunc,
		Run:               emptyRun,
	}
	child2Cmd := &Command{
		Use:               "child2",
		ValidArgsFunction: validArgsFunc2,
		Run:               emptyRun,
	}
	rootCmd.AddCommand(child1Cmd, child2Cmd)

	// Test completion of first sub-command with empty argument
	output, err := executeCommand(rootCmd, ShellCompRequestCmd, "child1", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := strings.Join([]string{
		"one\tThe first",
		"two\tThe second",
		":0",
		"Completion ended with directive: ShellCompDirectiveDefault", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test completion of first sub-command with a prefix to complete
	output, err = executeCommand(rootCmd, ShellCompRequestCmd, "child1", "t")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"two\tThe second",
		":0",
		"Completion ended with directive: ShellCompDirectiveDefault", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Check completing with wrong number of args
	output, err = executeCommand(rootCmd, ShellCompRequestCmd, "child1", "unexpectedArg", "t")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test completion of second sub-command with empty argument
	output, err = executeCommand(rootCmd, ShellCompRequestCmd, "child2", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"three\tThe third",
		"four\tThe fourth",
		":0",
		"Completion ended with directive: ShellCompDirectiveDefault", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	output, err = executeCommand(rootCmd, ShellCompRequestCmd, "child2", "t")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"three\tThe third",
		":0",
		"Completion ended with directive: ShellCompDirectiveDefault", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Check completing with wrong number of args
	output, err = executeCommand(rootCmd, ShellCompRequestCmd, "child2", "unexpectedArg", "t")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}
}

func TestFlagCompletionInGoWithDesc(t *testing.T) {
	rootCmd := &Command{
		Use: "root",
		Run: emptyRun,
	}
	rootCmd.Flags().IntP("introot", "i", -1, "help message for flag introot")
	assertNoErr(t, rootCmd.RegisterFlagCompletionFunc("introot", func(cmd *Command, args []string, toComplete string) ([]string, ShellCompDirective) {
		completions := []string{}
		for _, comp := range []string{"1\tThe first", "2\tThe second", "10\tThe tenth"} {
			if strings.HasPrefix(comp, toComplete) {
				completions = append(completions, comp)
			}
		}
		return completions, ShellCompDirectiveDefault
	}))
	rootCmd.Flags().String("filename", "", "Enter a filename")
	assertNoErr(t, rootCmd.RegisterFlagCompletionFunc("filename", func(cmd *Command, args []string, toComplete string) ([]string, ShellCompDirective) {
		completions := []string{}
		for _, comp := range []string{"file.yaml\tYAML format", "myfile.json\tJSON format", "file.xml\tXML format"} {
			if strings.HasPrefix(comp, toComplete) {
				completions = append(completions, comp)
			}
		}
		return completions, ShellCompDirectiveNoSpace | ShellCompDirectiveNoFileComp
	}))

	// Test completing an empty string
	output, err := executeCommand(rootCmd, ShellCompRequestCmd, "--introot", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := strings.Join([]string{
		"1\tThe first",
		"2\tThe second",
		"10\tThe tenth",
		":0",
		"Completion ended with directive: ShellCompDirectiveDefault", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Check completing with a prefix
	output, err = executeCommand(rootCmd, ShellCompRequestCmd, "--introot", "1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"1\tThe first",
		"10\tThe tenth",
		":0",
		"Completion ended with directive: ShellCompDirectiveDefault", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test completing an empty string
	output, err = executeCommand(rootCmd, ShellCompRequestCmd, "--filename", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"file.yaml\tYAML format",
		"myfile.json\tJSON format",
		"file.xml\tXML format",
		":6",
		"Completion ended with directive: ShellCompDirectiveNoSpace, ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Check completing with a prefix
	output, err = executeCommand(rootCmd, ShellCompRequestCmd, "--filename", "f")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"file.yaml\tYAML format",
		"file.xml\tXML format",
		":6",
		"Completion ended with directive: ShellCompDirectiveNoSpace, ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}
}

func TestValidArgsNotValidArgsFunc(t *testing.T) {
	rootCmd := &Command{
		Use:       "root",
		ValidArgs: []string{"one", "two"},
		ValidArgsFunction: func(cmd *Command, args []string, toComplete string) ([]string, ShellCompDirective) {
			return []string{"three", "four"}, ShellCompDirectiveNoFileComp
		},
		Run: emptyRun,
	}

	// Test that if both ValidArgs and ValidArgsFunction are present
	// only ValidArgs is considered
	output, err := executeCommand(rootCmd, ShellCompNoDescRequestCmd, "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := strings.Join([]string{
		"one",
		"two",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Check completing with a prefix
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "t")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"two",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}
}

func TestArgAliasesCompletionInGo(t *testing.T) {
	rootCmd := &Command{
		Use:        "root",
		Args:       OnlyValidArgs,
		ValidArgs:  []string{"one", "two", "three"},
		ArgAliases: []string{"un", "deux", "trois"},
		Run:        emptyRun,
	}

	// Test that argaliases are not completed when there are validargs that match
	output, err := executeCommand(rootCmd, ShellCompNoDescRequestCmd, "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := strings.Join([]string{
		"one",
		"two",
		"three",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that argaliases are not completed when there are validargs that match using a prefix
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "t")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"two",
		"three",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that argaliases are completed when there are no validargs that match
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "tr")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"trois",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}
}

func TestCompleteHelp(t *testing.T) {
	rootCmd := &Command{Use: "root", Args: NoArgs, Run: emptyRun}
	child1Cmd := &Command{
		Use: "child1",
		Run: emptyRun,
	}
	child2Cmd := &Command{
		Use: "child2",
		Run: emptyRun,
	}
	rootCmd.AddCommand(child1Cmd, child2Cmd)

	child3Cmd := &Command{
		Use: "child3",
		Run: emptyRun,
	}
	child1Cmd.AddCommand(child3Cmd)

	// Test that completion includes the help command
	output, err := executeCommand(rootCmd, ShellCompNoDescRequestCmd, "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := strings.Join([]string{
		"child1",
		"child2",
		"help",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test sub-commands are completed on first level of help command
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "help", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"child1",
		"child2",
		"help", // "<program> help help" is a valid command, so should be completed
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test sub-commands are completed on first level of help command
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "help", "child1", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"child3",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}
}
