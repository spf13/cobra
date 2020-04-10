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
	rootCmd.GenBashCompletion(buf)
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
	rootCmd.GenBashCompletion(buf)
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
	rootCmd.GenBashCompletion(buf)
	output := buf.String()

	check(t, output, ShellCompNoDescRequestCmd)
}

func TestCompleteNoDesCmdInFishScript(t *testing.T) {
	rootCmd := &Command{Use: "root", Args: NoArgs, Run: emptyRun}
	child := &Command{
		Use:               "child",
		ValidArgsFunction: validArgsFunc,
		Run:               emptyRun,
	}
	rootCmd.AddCommand(child)

	buf := new(bytes.Buffer)
	rootCmd.GenFishCompletion(buf, false)
	output := buf.String()

	check(t, output, ShellCompNoDescRequestCmd)
}

func TestCompleteCmdInFishScript(t *testing.T) {
	rootCmd := &Command{Use: "root", Args: NoArgs, Run: emptyRun}
	child := &Command{
		Use:               "child",
		ValidArgsFunction: validArgsFunc,
		Run:               emptyRun,
	}
	rootCmd.AddCommand(child)

	buf := new(bytes.Buffer)
	rootCmd.GenFishCompletion(buf, true)
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
	rootCmd.RegisterFlagCompletionFunc("introot", func(cmd *Command, args []string, toComplete string) ([]string, ShellCompDirective) {
		completions := []string{}
		for _, comp := range []string{"1\tThe first", "2\tThe second", "10\tThe tenth"} {
			if strings.HasPrefix(comp, toComplete) {
				completions = append(completions, comp)
			}
		}
		return completions, ShellCompDirectiveDefault
	})
	rootCmd.Flags().String("filename", "", "Enter a filename")
	rootCmd.RegisterFlagCompletionFunc("filename", func(cmd *Command, args []string, toComplete string) ([]string, ShellCompDirective) {
		completions := []string{}
		for _, comp := range []string{"file.yaml\tYAML format", "myfile.json\tJSON format", "file.xml\tXML format"} {
			if strings.HasPrefix(comp, toComplete) {
				completions = append(completions, comp)
			}
		}
		return completions, ShellCompDirectiveNoSpace | ShellCompDirectiveNoFileComp
	})

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
	rootCmd.RegisterFlagCompletionFunc("introot", func(cmd *Command, args []string, toComplete string) ([]string, ShellCompDirective) {
		completions := []string{}
		for _, comp := range []string{"1\tThe first", "2\tThe second", "10\tThe tenth"} {
			if strings.HasPrefix(comp, toComplete) {
				completions = append(completions, comp)
			}
		}
		return completions, ShellCompDirectiveDefault
	})
	rootCmd.Flags().String("filename", "", "Enter a filename")
	rootCmd.RegisterFlagCompletionFunc("filename", func(cmd *Command, args []string, toComplete string) ([]string, ShellCompDirective) {
		completions := []string{}
		for _, comp := range []string{"file.yaml\tYAML format", "myfile.json\tJSON format", "file.xml\tXML format"} {
			if strings.HasPrefix(comp, toComplete) {
				completions = append(completions, comp)
			}
		}
		return completions, ShellCompDirectiveNoSpace | ShellCompDirectiveNoFileComp
	})

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
