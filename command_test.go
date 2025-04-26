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
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/pflag"
)

// emptyRun does nothing and is used as a placeholder or default function. It takes a pointer to a Command struct and a slice of strings as parameters but does not return any value.
func emptyRun(*Command, []string) {}

// ExecuteCommand executes a command with the given arguments and returns its output.
//
// Parameters:
// - root: The root command to execute.
// - args: Additional arguments to pass to the command.
//
// Returns:
// - output: The output of the executed command.
// - err: An error if the command execution fails.
func executeCommand(root *Command, args ...string) (output string, err error) {
	_, output, err = executeCommandC(root, args...)
	return output, err
}

// executeCommandWithContext executes the given command in the context of a specific context.
// It sets up the command's output and error buffers, applies the provided arguments,
// and runs the command within the given context. The function returns the combined output
// from stdout and stderr as a string, along with any errors that occur during execution.
// The command is expected to be properly configured beforehand, with its out, err, and args set appropriately.
func executeCommandWithContext(ctx context.Context, root *Command, args ...string) (output string, err error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)

	err = root.ExecuteContext(ctx)

	return buf.String(), err
}

// executeCommandC executes the given command with the provided arguments and returns the resulting command object, output, and any error that occurs.
// The function sets the standard output and error to a buffer and captures the output during execution. It then calls ExecuteC on the root command with the specified arguments.
//
// Parameters:
// - root: A pointer to the Command object to be executed.
// - args: A variadic slice of strings representing the arguments to pass to the command.
//
// Returns:
// - c: A pointer to the Command object that was executed.
// - output: A string containing the captured standard output and error from the execution.
// - err: An error if an error occurred during execution, otherwise nil.
func executeCommandC(root *Command, args ...string) (c *Command, output string, err error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)

	c, err = root.ExecuteC()

	return c, buf.String(), err
}

// executeCommandWithContextC executes the given command with the specified context and arguments.
// It captures both the output and any errors produced by the command execution.
//
// Parameters:
//   - ctx: The context to use for executing the command, allowing for cancellation or timeouts.
//   - root: The root command to be executed.
//   - args: Additional arguments to pass to the command.
//
// Returns:
//   - c: The command that was executed.
//   - output: The captured standard output and error of the command execution.
//   - err: Any error encountered during the execution of the command.
func executeCommandWithContextC(ctx context.Context, root *Command, args ...string) (c *Command, output string, err error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)

	c, err = root.ExecuteContextC(ctx)

	return c, buf.String(), err
}

// ResetCommandLineFlagSet resets the command line flag set to a new one with os.Args[0] as the program name and ExitOnError as the error handling policy. This is useful for reinitializing flags in tests or when flags need to be reset between runs.
func resetCommandLineFlagSet() {
	pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)
}

// checkStringContains checks if the 'got' string contains the 'expected' substring.
// If not, it logs an error message indicating the mismatch between the expected and actual strings.
func checkStringContains(t *testing.T, got, expected string) {
	if !strings.Contains(got, expected) {
		t.Errorf("Expected to contain: \n %v\nGot:\n %v\n", expected, got)
	}
}

// checkStringOmits checks if the `got` string contains the `expected` substring.
// If it does, the test fails with an error message indicating what was expected and what was got.
func checkStringOmits(t *testing.T, got, expected string) {
	if strings.Contains(got, expected) {
		t.Errorf("Expected to not contain: \n %v\nGot: %v", expected, got)
	}
}

const onetwo = "one two"

// TestSingleCommand tests the execution of a single command within a root command.
//
// It sets up a root command with two subcommands "a" and "b". The root command expects exactly two arguments.
// After executing the command with arguments "one" and "two", it checks if there is no unexpected output, error,
// and if the arguments passed to the root command are as expected.
func TestSingleCommand(t *testing.T) {
	var rootCmdArgs []string
	rootCmd := &Command{
		Use:  "root",
		Args: ExactArgs(2),
		Run:  func(_ *Command, args []string) { rootCmdArgs = args },
	}
	aCmd := &Command{Use: "a", Args: NoArgs, Run: emptyRun}
	bCmd := &Command{Use: "b", Args: NoArgs, Run: emptyRun}
	rootCmd.AddCommand(aCmd, bCmd)

	output, err := executeCommand(rootCmd, "one", "two")
	if output != "" {
		t.Errorf("Unexpected output: %v", output)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	got := strings.Join(rootCmdArgs, " ")
	if got != onetwo {
		t.Errorf("rootCmdArgs expected: %q, got: %q", onetwo, got)
	}
}

// TestChildCommand tests the behavior of the root command with a child command that takes exactly two arguments.
// It verifies that the child command receives the correct arguments and no unexpected output or error is returned.
func TestChildCommand(t *testing.T) {
	var child1CmdArgs []string
	rootCmd := &Command{Use: "root", Args: NoArgs, Run: emptyRun}
	child1Cmd := &Command{
		Use:  "child1",
		Args: ExactArgs(2),
		Run:  func(_ *Command, args []string) { child1CmdArgs = args },
	}
	child2Cmd := &Command{Use: "child2", Args: NoArgs, Run: emptyRun}
	rootCmd.AddCommand(child1Cmd, child2Cmd)

	output, err := executeCommand(rootCmd, "child1", "one", "two")
	if output != "" {
		t.Errorf("Unexpected output: %v", output)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	got := strings.Join(child1CmdArgs, " ")
	if got != onetwo {
		t.Errorf("child1CmdArgs expected: %q, got: %q", onetwo, got)
	}
}

// TestCallCommandWithoutSubcommands tests the scenario where a command is called without any subcommands. It asserts that there should be no errors when calling such a command.
func TestCallCommandWithoutSubcommands(t *testing.T) {
	rootCmd := &Command{Use: "root", Args: NoArgs, Run: emptyRun}
	_, err := executeCommand(rootCmd)
	if err != nil {
		t.Errorf("Calling command without subcommands should not have error: %v", err)
	}
}

// TestRootExecuteUnknownCommand tests the execution of an unknown command in a root command hierarchy.
// It creates a root command with a child command and attempts to execute an unknown command.
// The function checks if the output contains the expected error message for an unknown command.
func TestRootExecuteUnknownCommand(t *testing.T) {
	rootCmd := &Command{Use: "root", Run: emptyRun}
	rootCmd.AddCommand(&Command{Use: "child", Run: emptyRun})

	output, _ := executeCommand(rootCmd, "unknown")

	expected := "Error: unknown command \"unknown\" for \"root\"\nRun 'root --help' for usage.\n"

	if output != expected {
		t.Errorf("Expected:\n %q\nGot:\n %q\n", expected, output)
	}
}

// TestSubcommandExecuteC tests the execution of a subcommand using the Command type.
// It sets up a root command with a child command and asserts that executing the child command
// does not produce any output, no error is returned, and the correct command name is returned.
func TestSubcommandExecuteC(t *testing.T) {
	rootCmd := &Command{Use: "root", Run: emptyRun}
	childCmd := &Command{Use: "child", Run: emptyRun}
	rootCmd.AddCommand(childCmd)

	c, output, err := executeCommandC(rootCmd, "child")
	if output != "" {
		t.Errorf("Unexpected output: %v", output)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if c.Name() != "child" {
		t.Errorf(`invalid command returned from ExecuteC: expected "child"', got: %q`, c.Name())
	}
}

// TestExecuteContext tests the execution of commands with a context.
func TestExecuteContext(t *testing.T) {
	ctx := context.TODO()

	ctxRun := func(cmd *Command, args []string) {
		if cmd.Context() != ctx {
			t.Errorf("Command %q must have context when called with ExecuteContext", cmd.Use)
		}
	}

	rootCmd := &Command{Use: "root", Run: ctxRun, PreRun: ctxRun}
	childCmd := &Command{Use: "child", Run: ctxRun, PreRun: ctxRun}
	granchildCmd := &Command{Use: "grandchild", Run: ctxRun, PreRun: ctxRun}

	childCmd.AddCommand(granchildCmd)
	rootCmd.AddCommand(childCmd)

	if _, err := executeCommandWithContext(ctx, rootCmd, ""); err != nil {
		t.Errorf("Root command must not fail: %+v", err)
	}

	if _, err := executeCommandWithContext(ctx, rootCmd, "child"); err != nil {
		t.Errorf("Subcommand must not fail: %+v", err)
	}

	if _, err := executeCommandWithContext(ctx, rootCmd, "child", "grandchild"); err != nil {
		t.Errorf("Command child must not fail: %+v", err)
	}
}

// TestExecuteContextC tests the ExecuteContext method with a context and verifies that all commands have the correct context.
func TestExecuteContextC(t *testing.T) {
	ctx := context.TODO()

	ctxRun := func(cmd *Command, args []string) {
		if cmd.Context() != ctx {
			t.Errorf("Command %q must have context when called with ExecuteContext", cmd.Use)
		}
	}

	rootCmd := &Command{Use: "root", Run: ctxRun, PreRun: ctxRun}
	childCmd := &Command{Use: "child", Run: ctxRun, PreRun: ctxRun}
	granchildCmd := &Command{Use: "grandchild", Run: ctxRun, PreRun: ctxRun}

	childCmd.AddCommand(granchildCmd)
	rootCmd.AddCommand(childCmd)

	if _, _, err := executeCommandWithContextC(ctx, rootCmd, ""); err != nil {
		t.Errorf("Root command must not fail: %+v", err)
	}

	if _, _, err := executeCommandWithContextC(ctx, rootCmd, "child"); err != nil {
		t.Errorf("Subcommand must not fail: %+v", err)
	}

	if _, _, err := executeCommandWithContextC(ctx, rootCmd, "child", "grandchild"); err != nil {
		t.Errorf("Command child must not fail: %+v", err)
	}
}

// TestExecute_NoContext tests the Execute function without a context.
func TestExecute_NoContext(t *testing.T) {
	run := func(cmd *Command, args []string) {
		if cmd.Context() != context.Background() {
			t.Errorf("Command %s must have background context", cmd.Use)
		}
	}

	rootCmd := &Command{Use: "root", Run: run, PreRun: run}
	childCmd := &Command{Use: "child", Run: run, PreRun: run}
	granchildCmd := &Command{Use: "grandchild", Run: run, PreRun: run}

	childCmd.AddCommand(granchildCmd)
	rootCmd.AddCommand(childCmd)

	if _, err := executeCommand(rootCmd, ""); err != nil {
		t.Errorf("Root command must not fail: %+v", err)
	}

	if _, err := executeCommand(rootCmd, "child"); err != nil {
		t.Errorf("Subcommand must not fail: %+v", err)
	}

	if _, err := executeCommand(rootCmd, "child", "grandchild"); err != nil {
		t.Errorf("Command child must not fail: %+v", err)
	}
}

// TestRootUnknownCommandSilenced tests that unknown commands are handled correctly when errors and usage are silenced.
func TestRootUnknownCommandSilenced(t *testing.T) {
	rootCmd := &Command{Use: "root", Run: emptyRun}
	rootCmd.SilenceErrors = true
	rootCmd.SilenceUsage = true
	rootCmd.AddCommand(&Command{Use: "child", Run: emptyRun})

	output, _ := executeCommand(rootCmd, "unknown")
	if output != "" {
		t.Errorf("Expected blank output, because of silenced usage.\nGot:\n %q\n", output)
	}
}

// TestCommandAlias tests the functionality of adding a command with aliases to a root command.
func TestCommandAlias(t *testing.T) {
	var timesCmdArgs []string
	rootCmd := &Command{Use: "root", Args: NoArgs, Run: emptyRun}
	echoCmd := &Command{
		Use:     "echo",
		Aliases: []string{"say", "tell"},
		Args:    NoArgs,
		Run:     emptyRun,
	}
	timesCmd := &Command{
		Use:  "times",
		Args: ExactArgs(2),
		Run:  func(_ *Command, args []string) { timesCmdArgs = args },
	}
	echoCmd.AddCommand(timesCmd)
	rootCmd.AddCommand(echoCmd)

	output, err := executeCommand(rootCmd, "tell", "times", "one", "two")
	if output != "" {
		t.Errorf("Unexpected output: %v", output)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	got := strings.Join(timesCmdArgs, " ")
	if got != onetwo {
		t.Errorf("timesCmdArgs expected: %v, got: %v", onetwo, got)
	}
}

// TestEnablePrefixMatching tests the functionality of enabling prefix matching for commands.
// It checks if the command arguments are correctly captured when prefix matching is enabled.
// The test asserts that there is no unexpected output or error, and that the captured arguments match the expected values.
// After the test, it resets the `EnablePrefixMatching` to its default value.
func TestEnablePrefixMatching(t *testing.T) {
	EnablePrefixMatching = true

	var aCmdArgs []string
	rootCmd := &Command{Use: "root", Args: NoArgs, Run: emptyRun}
	aCmd := &Command{
		Use:  "aCmd",
		Args: ExactArgs(2),
		Run:  func(_ *Command, args []string) { aCmdArgs = args },
	}
	bCmd := &Command{Use: "bCmd", Args: NoArgs, Run: emptyRun}
	rootCmd.AddCommand(aCmd, bCmd)

	output, err := executeCommand(rootCmd, "a", "one", "two")
	if output != "" {
		t.Errorf("Unexpected output: %v", output)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	got := strings.Join(aCmdArgs, " ")
	if got != onetwo {
		t.Errorf("aCmdArgs expected: %q, got: %q", onetwo, got)
	}

	EnablePrefixMatching = defaultPrefixMatching
}

// TestAliasPrefixMatching tests the alias prefix matching feature of commands.
func TestAliasPrefixMatching(t *testing.T) {
	EnablePrefixMatching = true

	var timesCmdArgs []string
	rootCmd := &Command{Use: "root", Args: NoArgs, Run: emptyRun}
	echoCmd := &Command{
		Use:     "echo",
		Aliases: []string{"say", "tell"},
		Args:    NoArgs,
		Run:     emptyRun,
	}
	timesCmd := &Command{
		Use:  "times",
		Args: ExactArgs(2),
		Run:  func(_ *Command, args []string) { timesCmdArgs = args },
	}
	echoCmd.AddCommand(timesCmd)
	rootCmd.AddCommand(echoCmd)

	output, err := executeCommand(rootCmd, "sa", "times", "one", "two")
	if output != "" {
		t.Errorf("Unexpected output: %v", output)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	got := strings.Join(timesCmdArgs, " ")
	if got != onetwo {
		t.Errorf("timesCmdArgs expected: %v, got: %v", onetwo, got)
	}

	EnablePrefixMatching = defaultPrefixMatching
}

// TestPlugin checks the usage of a plugin command with another command like `kubectl`. The
// executable name is `kubectl-plugin`, but it's executed as `kubectl plugin`. The help text should
// reflect how the command is invoked. It verifies that the output contains specific strings indicating
// the correct usage and annotations.
func TestPlugin(t *testing.T) {
	cmd := &Command{
		Use:     "kubectl-plugin",
		Version: "1.0.0",
		Args:    NoArgs,
		Annotations: map[string]string{
			CommandDisplayNameAnnotation: "kubectl plugin",
		},
		Run: emptyRun,
	}

	cmdHelp, err := executeCommand(cmd, "-h")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringContains(t, cmdHelp, "kubectl plugin [flags]")
	checkStringContains(t, cmdHelp, "help for kubectl plugin")
	checkStringContains(t, cmdHelp, "version for kubectl plugin")
}

// TestPluginWithSubCommands checks usage as a plugin with subcommands.
func TestPluginWithSubCommands(t *testing.T) {
	rootCmd := &Command{
		Use:     "kubectl-plugin",
		Version: "1.0.0",
		Args:    NoArgs,
		Annotations: map[string]string{
			CommandDisplayNameAnnotation: "kubectl plugin",
		},
	}

	subCmd := &Command{Use: "sub [flags]", Args: NoArgs, Run: emptyRun}
	rootCmd.AddCommand(subCmd)

	rootHelp, err := executeCommand(rootCmd, "-h")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringContains(t, rootHelp, "kubectl plugin [command]")
	checkStringContains(t, rootHelp, "help for kubectl plugin")
	checkStringContains(t, rootHelp, "version for kubectl plugin")
	checkStringContains(t, rootHelp, "kubectl plugin [command] --help")

	childHelp, err := executeCommand(rootCmd, "sub", "-h")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringContains(t, childHelp, "kubectl plugin sub [flags]")
	checkStringContains(t, childHelp, "help for sub")

	helpHelp, err := executeCommand(rootCmd, "help", "-h")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringContains(t, helpHelp, "kubectl plugin help [path to command]")
	checkStringContains(t, helpHelp, "kubectl plugin help [command]")
}

// TestChildSameName checks the correct behaviour of cobra in cases,
// when an application with name "foo" and with subcommand "foo"
// is executed with args "foo foo".
func TestChildSameName(t *testing.T) {
	var fooCmdArgs []string
	rootCmd := &Command{Use: "foo", Args: NoArgs, Run: emptyRun}
	fooCmd := &Command{
		Use:  "foo",
		Args: ExactArgs(2),
		Run:  func(_ *Command, args []string) { fooCmdArgs = args },
	}
	barCmd := &Command{Use: "bar", Args: NoArgs, Run: emptyRun}
	rootCmd.AddCommand(fooCmd, barCmd)

	output, err := executeCommand(rootCmd, "foo", "one", "two")
	if output != "" {
		t.Errorf("Unexpected output: %v", output)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	got := strings.Join(fooCmdArgs, " ")
	if got != onetwo {
		t.Errorf("fooCmdArgs expected: %v, got: %v", onetwo, got)
	}
}

// TestGrandChildSameName checks the correct behavior of cobra in cases where
// a user has a root command and a grandchild with the same name. It ensures that
// the execution is correctly routed to the intended child command.
func TestGrandChildSameName(t *testing.T) {
	var fooCmdArgs []string
	rootCmd := &Command{Use: "foo", Args: NoArgs, Run: emptyRun}
	barCmd := &Command{Use: "bar", Args: NoArgs, Run: emptyRun}
	fooCmd := &Command{
		Use:  "foo",
		Args: ExactArgs(2),
		Run:  func(_ *Command, args []string) { fooCmdArgs = args },
	}
	barCmd.AddCommand(fooCmd)
	rootCmd.AddCommand(barCmd)

	output, err := executeCommand(rootCmd, "bar", "foo", "one", "two")
	if output != "" {
		t.Errorf("Unexpected output: %v", output)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	got := strings.Join(fooCmdArgs, " ")
	if got != onetwo {
		t.Errorf("fooCmdArgs expected: %v, got: %v", onetwo, got)
	}
}

// TestFlagLong tests the handling of long flags in a command.
func TestFlagLong(t *testing.T) {
	var cArgs []string
	c := &Command{
		Use:  "c",
		Args: ArbitraryArgs,
		Run:  func(_ *Command, args []string) { cArgs = args },
	}

	var intFlagValue int
	var stringFlagValue string
	c.Flags().IntVar(&intFlagValue, "intf", -1, "")
	c.Flags().StringVar(&stringFlagValue, "sf", "", "")

	output, err := executeCommand(c, "--intf=7", "--sf=abc", "one", "--", "two")
	if output != "" {
		t.Errorf("Unexpected output: %v", output)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if c.ArgsLenAtDash() != 1 {
		t.Errorf("Expected ArgsLenAtDash: %v but got %v", 1, c.ArgsLenAtDash())
	}
	if intFlagValue != 7 {
		t.Errorf("Expected intFlagValue: %v, got %v", 7, intFlagValue)
	}
	if stringFlagValue != "abc" {
		t.Errorf("Expected stringFlagValue: %q, got %q", "abc", stringFlagValue)
	}

	got := strings.Join(cArgs, " ")
	if got != onetwo {
		t.Errorf("rootCmdArgs expected: %q, got: %q", onetwo, got)
	}
}

// TestFlagShort tests the functionality of short flag parsing in a Command.
// It sets up a command with integer and string flags, executes it with specific arguments,
// and verifies that the flags are parsed correctly and no unexpected output or error is returned.
func TestFlagShort(t *testing.T) {
	var cArgs []string
	c := &Command{
		Use:  "c",
		Args: ArbitraryArgs,
		Run:  func(_ *Command, args []string) { cArgs = args },
	}

	var intFlagValue int
	var stringFlagValue string
	c.Flags().IntVarP(&intFlagValue, "intf", "i", -1, "")
	c.Flags().StringVarP(&stringFlagValue, "sf", "s", "", "")

	output, err := executeCommand(c, "-i", "7", "-sabc", "one", "two")
	if output != "" {
		t.Errorf("Unexpected output: %v", output)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if intFlagValue != 7 {
		t.Errorf("Expected flag value: %v, got %v", 7, intFlagValue)
	}
	if stringFlagValue != "abc" {
		t.Errorf("Expected stringFlagValue: %q, got %q", "abc", stringFlagValue)
	}

	got := strings.Join(cArgs, " ")
	if got != onetwo {
		t.Errorf("rootCmdArgs expected: %q, got: %q", onetwo, got)
	}
}

// TestChildFlag tests the functionality of a child command with an integer flag.
func TestChildFlag(t *testing.T) {
	rootCmd := &Command{Use: "root", Run: emptyRun}
	childCmd := &Command{Use: "child", Run: emptyRun}
	rootCmd.AddCommand(childCmd)

	var intFlagValue int
	childCmd.Flags().IntVarP(&intFlagValue, "intf", "i", -1, "")

	output, err := executeCommand(rootCmd, "child", "-i7")
	if output != "" {
		t.Errorf("Unexpected output: %v", output)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if intFlagValue != 7 {
		t.Errorf("Expected flag value: %v, got %v", 7, intFlagValue)
	}
}

// TestChildFlagWithParentLocalFlag tests the behavior of child command flags when parent has a local flag with the same shorthand.
//
// It creates a root command and a child command, adds the child to the root, sets up string and integer flags on both commands,
// executes the command with specific arguments, and verifies that an error is returned due to a flag conflict. Additionally,
// it checks that the integer flag value from the child command is set correctly.
func TestChildFlagWithParentLocalFlag(t *testing.T) {
	rootCmd := &Command{Use: "root", Run: emptyRun}
	childCmd := &Command{Use: "child", Run: emptyRun}
	rootCmd.AddCommand(childCmd)

	var intFlagValue int
	rootCmd.Flags().StringP("sf", "s", "", "")
	childCmd.Flags().IntVarP(&intFlagValue, "intf", "i", -1, "")

	_, err := executeCommand(rootCmd, "child", "-i7", "-sabc")
	if err == nil {
		t.Errorf("Invalid flag should generate error")
	}

	checkStringContains(t, err.Error(), "unknown shorthand")

	if intFlagValue != 7 {
		t.Errorf("Expected flag value: %v, got %v", 7, intFlagValue)
	}
}

// TestFlagInvalidInput tests the behavior of the command when provided with an invalid integer flag value.
// It ensures that the function correctly identifies and returns an error for invalid input.
func TestFlagInvalidInput(t *testing.T) {
	rootCmd := &Command{Use: "root", Run: emptyRun}
	rootCmd.Flags().IntP("intf", "i", -1, "")

	_, err := executeCommand(rootCmd, "-iabc")
	if err == nil {
		t.Errorf("Invalid flag value should generate error")
	}

	checkStringContains(t, err.Error(), "invalid syntax")
}

// TestFlagBeforeCommand tests the behavior of flags when specified before the command.
func TestFlagBeforeCommand(t *testing.T) {
	rootCmd := &Command{Use: "root", Run: emptyRun}
	childCmd := &Command{Use: "child", Run: emptyRun}
	rootCmd.AddCommand(childCmd)

	var flagValue int
	childCmd.Flags().IntVarP(&flagValue, "intf", "i", -1, "")

	// With short flag.
	_, err := executeCommand(rootCmd, "-i7", "child")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if flagValue != 7 {
		t.Errorf("Expected flag value: %v, got %v", 7, flagValue)
	}

	// With long flag.
	_, err = executeCommand(rootCmd, "--intf=8", "child")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if flagValue != 8 {
		t.Errorf("Expected flag value: %v, got %v", 9, flagValue)
	}
}

// TestStripFlags runs a series of tests to verify that the stripFlags function correctly removes flags from an input slice of strings.
//
// It creates a Command instance with persistent and local flags. Each test case provides an input slice of strings representing command-line arguments and expected output after flag removal.
//
// The function iterates over each test case, calls stripFlags with the input, and compares the result to the expected output. If the results do not match, it logs an error indicating the test case number, expected output, and actual output.
//
// This test ensures that the stripFlags function is correctly identifying and removing flags based on their prefixes and names.
func TestStripFlags(t *testing.T) {
	tests := []struct {
		input  []string
		output []string
	}{
		{
			[]string{"foo", "bar"},
			[]string{"foo", "bar"},
		},
		{
			[]string{"foo", "--str", "-s"},
			[]string{"foo"},
		},
		{
			[]string{"-s", "foo", "--str", "bar"},
			[]string{},
		},
		{
			[]string{"-i10", "echo"},
			[]string{"echo"},
		},
		{
			[]string{"-i=10", "echo"},
			[]string{"echo"},
		},
		{
			[]string{"--int=100", "echo"},
			[]string{"echo"},
		},
		{
			[]string{"-ib", "echo", "-sfoo", "baz"},
			[]string{"echo", "baz"},
		},
		{
			[]string{"-i=baz", "bar", "-i", "foo", "blah"},
			[]string{"bar", "blah"},
		},
		{
			[]string{"--int=baz", "-sbar", "-i", "foo", "blah"},
			[]string{"blah"},
		},
		{
			[]string{"--bool", "bar", "-i", "foo", "blah"},
			[]string{"bar", "blah"},
		},
		{
			[]string{"-b", "bar", "-i", "foo", "blah"},
			[]string{"bar", "blah"},
		},
		{
			[]string{"--persist", "bar"},
			[]string{"bar"},
		},
		{
			[]string{"-p", "bar"},
			[]string{"bar"},
		},
	}

	c := &Command{Use: "c", Run: emptyRun}
	c.PersistentFlags().BoolP("persist", "p", false, "")
	c.Flags().IntP("int", "i", -1, "")
	c.Flags().StringP("str", "s", "", "")
	c.Flags().BoolP("bool", "b", false, "")

	for i, test := range tests {
		got := stripFlags(test.input, c)
		if !reflect.DeepEqual(test.output, got) {
			t.Errorf("(%v) Expected: %v, got: %v", i, test.output, got)
		}
	}
}

// TestDisableFlagParsing tests the functionality of disabling flag parsing in a command.
// It creates a new Command with DisableFlagParsing set to true and verifies that the Run function receives all arguments without flag parsing.
func TestDisableFlagParsing(t *testing.T) {
	var cArgs []string
	c := &Command{
		Use:                "c",
		DisableFlagParsing: true,
		Run: func(_ *Command, args []string) {
			cArgs = args
		},
	}

	args := []string{"cmd", "-v", "-race", "-file", "foo.go"}
	output, err := executeCommand(c, args...)
	if output != "" {
		t.Errorf("Unexpected output: %v", output)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !reflect.DeepEqual(args, cArgs) {
		t.Errorf("Expected: %v, got: %v", args, cArgs)
	}
}

// TestPersistentFlagsOnSameCommand tests the behavior of persistent flags when used on the same command.
//
// It creates a root command with a persistent integer flag and runs it with specific arguments.
// The function verifies that the flag value is correctly set and that the command arguments are as expected.
func TestPersistentFlagsOnSameCommand(t *testing.T) {
	var rootCmdArgs []string
	rootCmd := &Command{
		Use:  "root",
		Args: ArbitraryArgs,
		Run:  func(_ *Command, args []string) { rootCmdArgs = args },
	}

	var flagValue int
	rootCmd.PersistentFlags().IntVarP(&flagValue, "intf", "i", -1, "")

	output, err := executeCommand(rootCmd, "-i7", "one", "two")
	if output != "" {
		t.Errorf("Unexpected output: %v", output)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	got := strings.Join(rootCmdArgs, " ")
	if got != onetwo {
		t.Errorf("rootCmdArgs expected: %q, got %q", onetwo, got)
	}
	if flagValue != 7 {
		t.Errorf("flagValue expected: %v, got %v", 7, flagValue)
	}
}

// TestEmptyInputs checks if flags are correctly parsed with blank strings in args.
func TestEmptyInputs(t *testing.T) {
	c := &Command{Use: "c", Run: emptyRun}

	var flagValue int
	c.Flags().IntVarP(&flagValue, "intf", "i", -1, "")

	output, err := executeCommand(c, "", "-i7", "")
	if output != "" {
		t.Errorf("Unexpected output: %v", output)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if flagValue != 7 {
		t.Errorf("flagValue expected: %v, got %v", 7, flagValue)
	}
}

// TestChildFlagShadowsParentPersistentFlag tests if a child command's flags shadow parent persistent flags.
func TestChildFlagShadowsParentPersistentFlag(t *testing.T) {
	parent := &Command{Use: "parent", Run: emptyRun}
	child := &Command{Use: "child", Run: emptyRun}

	parent.PersistentFlags().Bool("boolf", false, "")
	parent.PersistentFlags().Int("intf", -1, "")
	child.Flags().String("strf", "", "")
	child.Flags().Int("intf", -1, "")

	parent.AddCommand(child)

	childInherited := child.InheritedFlags()
	childLocal := child.LocalFlags()

	if childLocal.Lookup("strf") == nil {
		t.Error(`LocalFlags expected to contain "strf", got "nil"`)
	}
	if childInherited.Lookup("boolf") == nil {
		t.Error(`InheritedFlags expected to contain "boolf", got "nil"`)
	}

	if childInherited.Lookup("intf") != nil {
		t.Errorf(`InheritedFlags should not contain shadowed flag "intf"`)
	}
	if childLocal.Lookup("intf") == nil {
		t.Error(`LocalFlags expected to contain "intf", got "nil"`)
	}
}

// TestPersistentFlagsOnChild tests that persistent flags set on the root command are available to its child commands.
func TestPersistentFlagsOnChild(t *testing.T) {
	var childCmdArgs []string
	rootCmd := &Command{Use: "root", Run: emptyRun}
	childCmd := &Command{
		Use:  "child",
		Args: ArbitraryArgs,
		Run:  func(_ *Command, args []string) { childCmdArgs = args },
	}
	rootCmd.AddCommand(childCmd)

	var parentFlagValue int
	var childFlagValue int
	rootCmd.PersistentFlags().IntVarP(&parentFlagValue, "parentf", "p", -1, "")
	childCmd.Flags().IntVarP(&childFlagValue, "childf", "c", -1, "")

	output, err := executeCommand(rootCmd, "child", "-c7", "-p8", "one", "two")
	if output != "" {
		t.Errorf("Unexpected output: %v", output)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	got := strings.Join(childCmdArgs, " ")
	if got != onetwo {
		t.Errorf("rootCmdArgs expected: %q, got: %q", onetwo, got)
	}
	if parentFlagValue != 8 {
		t.Errorf("parentFlagValue expected: %v, got %v", 8, parentFlagValue)
	}
	if childFlagValue != 7 {
		t.Errorf("childFlagValue expected: %v, got %v", 7, childFlagValue)
	}
}

// TestRequiredFlags checks that required flags are enforced when running a command.
func TestRequiredFlags(t *testing.T) {
	c := &Command{Use: "c", Run: emptyRun}
	c.Flags().String("foo1", "", "")
	assertNoErr(t, c.MarkFlagRequired("foo1"))
	c.Flags().String("foo2", "", "")
	assertNoErr(t, c.MarkFlagRequired("foo2"))
	c.Flags().String("bar", "", "")

	expected := fmt.Sprintf("required flag(s) %q, %q not set", "foo1", "foo2")

	_, err := executeCommand(c)
	got := err.Error()

	if got != expected {
		t.Errorf("Expected error: %q, got: %q", expected, got)
	}
}

// TestPersistentRequiredFlags tests the marking of persistent and local flags as required in a command hierarchy.
func TestPersistentRequiredFlags(t *testing.T) {
	parent := &Command{Use: "parent", Run: emptyRun}
	parent.PersistentFlags().String("foo1", "", "")
	assertNoErr(t, parent.MarkPersistentFlagRequired("foo1"))
	parent.PersistentFlags().String("foo2", "", "")
	assertNoErr(t, parent.MarkPersistentFlagRequired("foo2"))
	parent.Flags().String("foo3", "", "")

	child := &Command{Use: "child", Run: emptyRun}
	child.Flags().String("bar1", "", "")
	assertNoErr(t, child.MarkFlagRequired("bar1"))
	child.Flags().String("bar2", "", "")
	assertNoErr(t, child.MarkFlagRequired("bar2"))
	child.Flags().String("bar3", "", "")

	parent.AddCommand(child)

	expected := fmt.Sprintf("required flag(s) %q, %q, %q, %q not set", "bar1", "bar2", "foo1", "foo2")

	_, err := executeCommand(parent, "child")
	if err.Error() != expected {
		t.Errorf("Expected %q, got %q", expected, err.Error())
	}
}

// TestPersistentRequiredFlagsWithDisableFlagParsing tests that a required persistent flag does not break commands that disable flag parsing.
func TestPersistentRequiredFlagsWithDisableFlagParsing(t *testing.T) {
	// Make sure a required persistent flag does not break
	// commands that disable flag parsing

	parent := &Command{Use: "parent", Run: emptyRun}
	parent.PersistentFlags().Bool("foo", false, "")
	flag := parent.PersistentFlags().Lookup("foo")
	assertNoErr(t, parent.MarkPersistentFlagRequired("foo"))

	child := &Command{Use: "child", Run: emptyRun}
	child.DisableFlagParsing = true

	parent.AddCommand(child)

	if _, err := executeCommand(parent, "--foo", "child"); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Reset the flag or else it will remember the state from the previous command
	flag.Changed = false
	if _, err := executeCommand(parent, "child", "--foo"); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Reset the flag or else it will remember the state from the previous command
	flag.Changed = false
	if _, err := executeCommand(parent, "child"); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

// TestInitHelpFlagMergesFlags verifies that when InitDefaultHelpFlag is called on a child command, it merges the help flag from its parent command.
//
// Parameters:
//   - t: A pointer to testing.T for test assertions.
//
// Returns:
//   None. The function uses t.Errorf to report failures.
func TestInitHelpFlagMergesFlags(t *testing.T) {
	usage := "custom flag"
	rootCmd := &Command{Use: "root"}
	rootCmd.PersistentFlags().Bool("help", false, "custom flag")
	childCmd := &Command{Use: "child"}
	rootCmd.AddCommand(childCmd)

	childCmd.InitDefaultHelpFlag()
	got := childCmd.Flags().Lookup("help").Usage
	if got != usage {
		t.Errorf("Expected the help flag from the root command with usage: %v\nGot the default with usage: %v", usage, got)
	}
}

// TestHelpCommandExecuted tests that the 'help' command is executed correctly.
// It verifies that the output contains the long description of the root command.
func TestHelpCommandExecuted(t *testing.T) {
	rootCmd := &Command{Use: "root", Long: "Long description", Run: emptyRun}
	rootCmd.AddCommand(&Command{Use: "child", Run: emptyRun})

	output, err := executeCommand(rootCmd, "help")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringContains(t, output, rootCmd.Long)
}

// TestHelpCommandExecutedOnChild tests that the help command is executed on a child command.
func TestHelpCommandExecutedOnChild(t *testing.T) {
	rootCmd := &Command{Use: "root", Run: emptyRun}
	childCmd := &Command{Use: "child", Long: "Long description", Run: emptyRun}
	rootCmd.AddCommand(childCmd)

	output, err := executeCommand(rootCmd, "help", "child")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringContains(t, output, childCmd.Long)
}

// TestHelpCommandExecutedOnChildWithFlagThatShadowsParentFlag tests that the help command executed on a child command shows the child's flags and not the parent's shadowed flags.
func TestHelpCommandExecutedOnChildWithFlagThatShadowsParentFlag(t *testing.T) {
	parent := &Command{Use: "parent", Run: emptyRun}
	child := &Command{Use: "child", Run: emptyRun}
	parent.AddCommand(child)

	parent.PersistentFlags().Bool("foo", false, "parent foo usage")
	parent.PersistentFlags().Bool("bar", false, "parent bar usage")
	child.Flags().Bool("foo", false, "child foo usage") // This shadows parent's foo flag
	child.Flags().Bool("baz", false, "child baz usage")

	got, err := executeCommand(parent, "help", "child")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := `Usage:
  parent child [flags]

Flags:
      --baz    child baz usage
      --foo    child foo usage
  -h, --help   help for child

Global Flags:
      --bar   parent bar usage
`

	if got != expected {
		t.Errorf("Help text mismatch.\nExpected:\n%s\n\nGot:\n%s\n", expected, got)
	}
}

// TestSetHelpCommand tests the SetHelpCommand method of the Command struct.
//
// It creates a new command and sets its help command. The help command runs a specific function that prints an expected string.
// The test then executes the help command and checks if the output contains the expected string, asserting no errors occur during execution.
func TestSetHelpCommand(t *testing.T) {
	c := &Command{Use: "c", Run: emptyRun}
	c.AddCommand(&Command{Use: "empty", Run: emptyRun})

	expected := "WORKS"
	c.SetHelpCommand(&Command{
		Use:   "help [command]",
		Short: "Help about any command",
		Long: `Help provides help for any command in the application.
	Simply type ` + c.Name() + ` help [path to command] for full details.`,
		Run: func(c *Command, _ []string) { c.Print(expected) },
	})

	got, err := executeCommand(c, "help")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if got != expected {
		t.Errorf("Expected to contain %q, got %q", expected, got)
	}
}

// TestSetHelpTemplate tests the SetHelpTemplate method of the Command struct.
//
// It verifies that setting a custom help template on a root command and its child commands works as expected.
// It also checks that the default help template is used when no custom template is set.
func TestSetHelpTemplate(t *testing.T) {
	rootCmd := &Command{Use: "root", Run: emptyRun}
	childCmd := &Command{Use: "child", Run: emptyRun}
	rootCmd.AddCommand(childCmd)

	rootCmd.SetHelpTemplate("WORKS {{.UseLine}}")

	// Call the help on the root command and check the new template is used
	got, err := executeCommand(rootCmd, "--help")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := "WORKS " + rootCmd.UseLine()
	if got != expected {
		t.Errorf("Expected %q, got %q", expected, got)
	}

	// Call the help on the child command and check
	// the new template is inherited from the parent
	got, err = executeCommand(rootCmd, "child", "--help")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = "WORKS " + childCmd.UseLine()
	if got != expected {
		t.Errorf("Expected %q, got %q", expected, got)
	}

	// Reset the root command help template and make sure
	// it falls back to the default
	rootCmd.SetHelpTemplate("")
	got, err = executeCommand(rootCmd, "--help")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !strings.Contains(got, "Usage:") {
		t.Errorf("Expected to contain %q, got %q", "Usage:", got)
	}
}

// TestHelpFlagExecuted tests if the help flag is executed correctly.
// It creates a root command with a long description and an empty run function.
// The test then executes the command with the "--help" flag and checks if the output contains the long description.
// If any error occurs during execution, it reports the error using the test's Errorf method.
// Additionally, it uses the checkStringContains helper function to verify that the output string contains the expected substring.
func TestHelpFlagExecuted(t *testing.T) {
	rootCmd := &Command{Use: "root", Long: "Long description", Run: emptyRun}

	output, err := executeCommand(rootCmd, "--help")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringContains(t, output, rootCmd.Long)
}

// TestHelpFlagExecutedOnChild tests whether the help flag is executed on a child command.
// It sets up a root command with a child command and executes the child command with the "--help" flag.
// The function then checks if the output contains the long description of the child command.
func TestHelpFlagExecutedOnChild(t *testing.T) {
	rootCmd := &Command{Use: "root", Run: emptyRun}
	childCmd := &Command{Use: "child", Long: "Long description", Run: emptyRun}
	rootCmd.AddCommand(childCmd)

	output, err := executeCommand(rootCmd, "child", "--help")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringContains(t, output, childCmd.Long)
}

// TestHelpFlagInHelp verifies that the '--help' flag is displayed in the help output for a child command when using `parent help child`.
//
// The test ensures that the Cobra library correctly handles the display of the '--help' flag in nested command structures.
// This addresses an issue reported in https://github.com/spf13/cobra/issues/302, ensuring proper functionality and user experience.
func TestHelpFlagInHelp(t *testing.T) {
	parentCmd := &Command{Use: "parent", Run: func(*Command, []string) {}}

	childCmd := &Command{Use: "child", Run: func(*Command, []string) {}}
	parentCmd.AddCommand(childCmd)

	output, err := executeCommand(parentCmd, "help", "child")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringContains(t, output, "[flags]")
}

// TestFlagsInUsage tests if the command usage includes flags section.
func TestFlagsInUsage(t *testing.T) {
	rootCmd := &Command{Use: "root", Args: NoArgs, Run: func(*Command, []string) {}}
	output, err := executeCommand(rootCmd, "--help")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringContains(t, output, "[flags]")
}

// TestHelpExecutedOnNonRunnableChild tests the behavior of help when executed on a non-runnable child command.
func TestHelpExecutedOnNonRunnableChild(t *testing.T) {
	rootCmd := &Command{Use: "root", Run: emptyRun}
	childCmd := &Command{Use: "child", Long: "Long description"}
	rootCmd.AddCommand(childCmd)

	output, err := executeCommand(rootCmd, "child")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringContains(t, output, childCmd.Long)
}

// TestSetUsageTemplate tests the functionality of SetUsageTemplate method.
// It verifies that setting a custom usage template on a command and its child commands works as expected.
// It also checks that resetting the template to empty falls back to the default usage format.
func TestSetUsageTemplate(t *testing.T) {
	rootCmd := &Command{Use: "root", Run: emptyRun}
	childCmd := &Command{Use: "child", Run: emptyRun}
	rootCmd.AddCommand(childCmd)

	rootCmd.SetUsageTemplate("WORKS {{.UseLine}}")

	// Trigger the usage on the root command and check the new template is used
	got, err := executeCommand(rootCmd, "--invalid")
	if err == nil {
		t.Errorf("Expected error but did not get one")
	}

	expected := "WORKS " + rootCmd.UseLine()
	checkStringContains(t, got, expected)

	// Trigger the usage on the child command and check
	// the new template is inherited from the parent
	got, err = executeCommand(rootCmd, "child", "--invalid")
	if err == nil {
		t.Errorf("Expected error but did not get one")
	}

	expected = "WORKS " + childCmd.UseLine()
	checkStringContains(t, got, expected)

	// Reset the root command usage template and make sure
	// it falls back to the default
	rootCmd.SetUsageTemplate("")
	got, err = executeCommand(rootCmd, "--invalid")
	if err == nil {
		t.Errorf("Expected error but did not get one")
	}

	if !strings.Contains(got, "Usage:") {
		t.Errorf("Expected to contain %q, got %q", "Usage:", got)
	}
}

// TestVersionFlagExecuted tests that the --version flag is executed correctly.
//
// It creates a root command with a specific version and an empty run function.
// The function then executes the command with the --version flag and checks if the output contains the expected version string.
func TestVersionFlagExecuted(t *testing.T) {
	rootCmd := &Command{Use: "root", Version: "1.0.0", Run: emptyRun}

	output, err := executeCommand(rootCmd, "--version", "arg1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringContains(t, output, "root version 1.0.0")
}

// TestVersionFlagExecutedDiplayName tests if the version flag executed with a display name.
func TestVersionFlagExecutedDiplayName(t *testing.T) {
	rootCmd := &Command{
		Use:     "kubectl-plugin",
		Version: "1.0.0",
		Annotations: map[string]string{
			CommandDisplayNameAnnotation: "kubectl plugin",
		},
		Run: emptyRun,
	}

	output, err := executeCommand(rootCmd, "--version", "arg1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringContains(t, output, "kubectl plugin version 1.0.0")
}

// TestVersionFlagExecutedWithNoName tests that the version flag is executed when no name is provided.
//
// It creates a root command with a version and an empty run function. It then executes the command with the --version flag and "arg1".
//
// The function asserts that there is no error returned and that the output contains the string "version 1.0.0".
func TestVersionFlagExecutedWithNoName(t *testing.T) {
	rootCmd := &Command{Version: "1.0.0", Run: emptyRun}

	output, err := executeCommand(rootCmd, "--version", "arg1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringContains(t, output, "version 1.0.0")
}

// TestShortAndLongVersionFlagInHelp tests that both the short and long version flags are present in the help output.
//
// Parameters:
//   - t: A testing.T instance to which assertions can be made.
//
// This function creates a root command with a specified use case, version, and empty run function.
// It then executes the command with the "--help" flag and checks if the "-v, --version" string is present in the output.
// If an error occurs during the execution, it asserts an error with a message indicating the unexpected error.
func TestShortAndLongVersionFlagInHelp(t *testing.T) {
	rootCmd := &Command{Use: "root", Version: "1.0.0", Run: emptyRun}

	output, err := executeCommand(rootCmd, "--help")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringContains(t, output, "-v, --version")
}

// TestLongVersionFlagOnlyInHelpWhenShortPredefined tests that the long version flag is only included in help when a short version flag is predefined.
//
// Parameters:
//   - t: The testing.T instance for running the test.
//
// Returns:
//   None. This function uses t.Errorf to report any unexpected errors encountered during the execution of the test.
func TestLongVersionFlagOnlyInHelpWhenShortPredefined(t *testing.T) {
	rootCmd := &Command{Use: "root", Version: "1.0.0", Run: emptyRun}
	rootCmd.Flags().StringP("foo", "v", "", "not a version flag")

	output, err := executeCommand(rootCmd, "--help")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringOmits(t, output, "-v, --version")
	checkStringContains(t, output, "--version")
}

// TestShorthandVersionFlagExecuted tests if the shorthand version flag is executed correctly.
// It sets up a root command with a specific use and version, executes it with a version flag,
// and checks if the output contains the expected version information.
func TestShorthandVersionFlagExecuted(t *testing.T) {
	rootCmd := &Command{Use: "root", Version: "1.0.0", Run: emptyRun}

	output, err := executeCommand(rootCmd, "-v", "arg1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringContains(t, output, "root version 1.0.0")
}

// TestVersionTemplate tests the custom version template of a command.
func TestVersionTemplate(t *testing.T) {
	rootCmd := &Command{Use: "root", Version: "1.0.0", Run: emptyRun}
	rootCmd.SetVersionTemplate(`customized version: {{.Version}}`)

	output, err := executeCommand(rootCmd, "--version", "arg1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringContains(t, output, "customized version: 1.0.0")
}

// TestShorthandVersionTemplate tests the execution of a command with a custom version template.
// It asserts that the output contains the customized version string and no errors occur during execution.
func TestShorthandVersionTemplate(t *testing.T) {
	rootCmd := &Command{Use: "root", Version: "1.0.0", Run: emptyRun}
	rootCmd.SetVersionTemplate(`customized version: {{.Version}}`)

	output, err := executeCommand(rootCmd, "-v", "arg1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringContains(t, output, "customized version: 1.0.0")
}

// TestRootErrPrefixExecutedOnSubcommand tests whether the root command's error prefix is applied to errors returned by subcommands.
func TestRootErrPrefixExecutedOnSubcommand(t *testing.T) {
	rootCmd := &Command{Use: "root", Run: emptyRun}
	rootCmd.SetErrPrefix("root error prefix:")
	rootCmd.AddCommand(&Command{Use: "sub", Run: emptyRun})

	output, err := executeCommand(rootCmd, "sub", "--unknown-flag")
	if err == nil {
		t.Errorf("Expected error")
	}

	checkStringContains(t, output, "root error prefix: unknown flag: --unknown-flag")
}

// TestRootAndSubErrPrefix tests the behavior of setting error prefixes on a root command and its subcommand.
// It checks if the correct error prefix is prepended when an unknown flag is used in either the root or subcommand.
func TestRootAndSubErrPrefix(t *testing.T) {
	rootCmd := &Command{Use: "root", Run: emptyRun}
	subCmd := &Command{Use: "sub", Run: emptyRun}
	rootCmd.AddCommand(subCmd)
	rootCmd.SetErrPrefix("root error prefix:")
	subCmd.SetErrPrefix("sub error prefix:")

	if output, err := executeCommand(rootCmd, "--unknown-root-flag"); err == nil {
		t.Errorf("Expected error")
	} else {
		checkStringContains(t, output, "root error prefix: unknown flag: --unknown-root-flag")
	}

	if output, err := executeCommand(rootCmd, "sub", "--unknown-sub-flag"); err == nil {
		t.Errorf("Expected error")
	} else {
		checkStringContains(t, output, "sub error prefix: unknown flag: --unknown-sub-flag")
	}
}

// TestVersionFlagExecutedOnSubcommand tests if the version flag is executed on a subcommand.
// It creates a root command with a specific version and adds a subcommand without a run function.
// It then executes the command with the --version flag followed by the subcommand name.
// Finally, it checks that the output contains the expected version information.
func TestVersionFlagExecutedOnSubcommand(t *testing.T) {
	rootCmd := &Command{Use: "root", Version: "1.0.0"}
	rootCmd.AddCommand(&Command{Use: "sub", Run: emptyRun})

	output, err := executeCommand(rootCmd, "--version", "sub")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringContains(t, output, "root version 1.0.0")
}

// TestShorthandVersionFlagExecutedOnSubcommand tests if the shorthand version flag (-v) is executed when passed to a subcommand.
//
// Parameters:
//   - t: A testing.T instance used for assertions and error reporting.
//
// Returns:
//   - None, but performs assertions on the output and any errors encountered during command execution.
func TestShorthandVersionFlagExecutedOnSubcommand(t *testing.T) {
	rootCmd := &Command{Use: "root", Version: "1.0.0"}
	rootCmd.AddCommand(&Command{Use: "sub", Run: emptyRun})

	output, err := executeCommand(rootCmd, "-v", "sub")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringContains(t, output, "root version 1.0.0")
}

// TestVersionFlagOnlyAddedToRoot checks that the --version flag is only added to the root command and not to its subcommands.
// It creates a root command with a version and adds a subcommand without the version flag.
// It then tries to execute the subcommand with the --version flag and expects an error indicating that the flag is unknown.
func TestVersionFlagOnlyAddedToRoot(t *testing.T) {
	rootCmd := &Command{Use: "root", Version: "1.0.0", Run: emptyRun}
	rootCmd.AddCommand(&Command{Use: "sub", Run: emptyRun})

	_, err := executeCommand(rootCmd, "sub", "--version")
	if err == nil {
		t.Errorf("Expected error")
	}

	checkStringContains(t, err.Error(), "unknown flag: --version")
}

// TestShortVersionFlagOnlyAddedToRoot checks if the short version flag is only added to the root command.
func TestShortVersionFlagOnlyAddedToRoot(t *testing.T) {
	rootCmd := &Command{Use: "root", Version: "1.0.0", Run: emptyRun}
	rootCmd.AddCommand(&Command{Use: "sub", Run: emptyRun})

	_, err := executeCommand(rootCmd, "sub", "-v")
	if err == nil {
		t.Errorf("Expected error")
	}

	checkStringContains(t, err.Error(), "unknown shorthand flag: 'v' in -v")
}

// TestVersionFlagOnlyExistsIfVersionNonEmpty checks if the `--version` flag is only present when the version string is non-empty.
// It creates a root command and tries to execute it with the `--version` flag, expecting an error since the version is empty.
// If no error is returned, it asserts that an error should have been returned.
// It then checks if the error message contains the expected substring "unknown flag: --version".
func TestVersionFlagOnlyExistsIfVersionNonEmpty(t *testing.T) {
	rootCmd := &Command{Use: "root", Run: emptyRun}

	_, err := executeCommand(rootCmd, "--version")
	if err == nil {
		t.Errorf("Expected error")
	}
	checkStringContains(t, err.Error(), "unknown flag: --version")
}

// TestShorthandVersionFlagOnlyExistsIfVersionNonEmpty tests that the version flag shorthand '-v' only exists if the Version field of the Command is non-empty.
//
// Parameters:
//   - t: A testing.T instance used to assert test conditions.
//
// The function does not return any values. If the Version field is empty and the shorthand flag '-v' is provided, an error is expected with a specific message indicating that the shorthand flag is unknown.
func TestShorthandVersionFlagOnlyExistsIfVersionNonEmpty(t *testing.T) {
	rootCmd := &Command{Use: "root", Run: emptyRun}

	_, err := executeCommand(rootCmd, "-v")
	if err == nil {
		t.Errorf("Expected error")
	}
	checkStringContains(t, err.Error(), "unknown shorthand flag: 'v' in -v")
}

// TestShorthandVersionFlagOnlyAddedIfShorthandNotDefined tests that the shorthand version flag is only added if it's not already defined.
//
// It creates a root command with a non-version flag named "notversion" and shorthand "v". When executing the command with "-v", it should return an error because "v" is not a valid argument for the existing shorthand flag. The test checks that the shorthand lookup returns the correct flag name and that the error message contains the expected text.
func TestShorthandVersionFlagOnlyAddedIfShorthandNotDefined(t *testing.T) {
	rootCmd := &Command{Use: "root", Run: emptyRun, Version: "1.2.3"}
	rootCmd.Flags().StringP("notversion", "v", "", "not a version flag")

	_, err := executeCommand(rootCmd, "-v")
	if err == nil {
		t.Errorf("Expected error")
	}
	check(t, rootCmd.Flags().ShorthandLookup("v").Name, "notversion")
	checkStringContains(t, err.Error(), "flag needs an argument: 'v' in -v")
}

// TestShorthandVersionFlagOnlyAddedIfVersionNotDefined tests that the shorthand version flag is only added if a custom version flag is not defined.
//
// Parameters:
//   - t: A testing.T object for running the test.
func TestShorthandVersionFlagOnlyAddedIfVersionNotDefined(t *testing.T) {
	rootCmd := &Command{Use: "root", Run: emptyRun, Version: "1.2.3"}
	rootCmd.Flags().Bool("version", false, "a different kind of version flag")

	_, err := executeCommand(rootCmd, "-v")
	if err == nil {
		t.Errorf("Expected error")
	}
	checkStringContains(t, err.Error(), "unknown shorthand flag: 'v' in -v")
}

// TestUsageIsNotPrintedTwice tests that the usage message for a command is not printed more than once.
func TestUsageIsNotPrintedTwice(t *testing.T) {
	var cmd = &Command{Use: "root"}
	var sub = &Command{Use: "sub"}
	cmd.AddCommand(sub)

	output, _ := executeCommand(cmd, "")
	if strings.Count(output, "Usage:") != 1 {
		t.Error("Usage output is not printed exactly once")
	}
}

// TestVisitParents tests the VisitParents method to ensure it correctly visits parent commands.
func TestVisitParents(t *testing.T) {
	c := &Command{Use: "app"}
	sub := &Command{Use: "sub"}
	dsub := &Command{Use: "dsub"}
	sub.AddCommand(dsub)
	c.AddCommand(sub)

	total := 0
	add := func(x *Command) {
		total++
	}
	sub.VisitParents(add)
	if total != 1 {
		t.Errorf("Should have visited 1 parent but visited %d", total)
	}

	total = 0
	dsub.VisitParents(add)
	if total != 2 {
		t.Errorf("Should have visited 2 parents but visited %d", total)
	}

	total = 0
	c.VisitParents(add)
	if total != 0 {
		t.Errorf("Should have visited no parents but visited %d", total)
	}
}

// TestSuggestions tests the command suggestion feature of the Command struct.
// It checks if the suggested commands are correctly generated based on user input,
// and ensures that suggestions are disabled when requested. The test covers various
// typo cases and expected suggestions or no suggestions at all.
func TestSuggestions(t *testing.T) {
	rootCmd := &Command{Use: "root", Run: emptyRun}
	timesCmd := &Command{
		Use:        "times",
		SuggestFor: []string{"counts"},
		Run:        emptyRun,
	}
	rootCmd.AddCommand(timesCmd)

	templateWithSuggestions := "Error: unknown command \"%s\" for \"root\"\n\nDid you mean this?\n\t%s\n\nRun 'root --help' for usage.\n"
	templateWithoutSuggestions := "Error: unknown command \"%s\" for \"root\"\nRun 'root --help' for usage.\n"

	tests := map[string]string{
		"time":     "times",
		"tiems":    "times",
		"tims":     "times",
		"timeS":    "times",
		"rimes":    "times",
		"ti":       "times",
		"t":        "times",
		"timely":   "times",
		"ri":       "",
		"timezone": "",
		"foo":      "",
		"counts":   "times",
	}

	for typo, suggestion := range tests {
		for _, suggestionsDisabled := range []bool{true, false} {
			rootCmd.DisableSuggestions = suggestionsDisabled

			var expected string
			output, _ := executeCommand(rootCmd, typo)

			if suggestion == "" || suggestionsDisabled {
				expected = fmt.Sprintf(templateWithoutSuggestions, typo)
			} else {
				expected = fmt.Sprintf(templateWithSuggestions, typo, suggestion)
			}

			if output != expected {
				t.Errorf("Unexpected response.\nExpected:\n %q\nGot:\n %q\n", expected, output)
			}
		}
	}
}

// TestCaseInsensitive tests the functionality of case-insensitive command matching.
// It checks if commands and their aliases are matched correctly based on the setting of EnableCaseInsensitive.
func TestCaseInsensitive(t *testing.T) {
	rootCmd := &Command{Use: "root", Run: emptyRun}
	childCmd := &Command{Use: "child", Run: emptyRun, Aliases: []string{"alternative"}}
	granchildCmd := &Command{Use: "GRANDCHILD", Run: emptyRun, Aliases: []string{"ALIAS"}}

	childCmd.AddCommand(granchildCmd)
	rootCmd.AddCommand(childCmd)

	tests := []struct {
		args                []string
		failWithoutEnabling bool
	}{
		{
			args:                []string{"child"},
			failWithoutEnabling: false,
		},
		{
			args:                []string{"CHILD"},
			failWithoutEnabling: true,
		},
		{
			args:                []string{"chILD"},
			failWithoutEnabling: true,
		},
		{
			args:                []string{"CHIld"},
			failWithoutEnabling: true,
		},
		{
			args:                []string{"alternative"},
			failWithoutEnabling: false,
		},
		{
			args:                []string{"ALTERNATIVE"},
			failWithoutEnabling: true,
		},
		{
			args:                []string{"ALTernatIVE"},
			failWithoutEnabling: true,
		},
		{
			args:                []string{"alternatiVE"},
			failWithoutEnabling: true,
		},
		{
			args:                []string{"child", "GRANDCHILD"},
			failWithoutEnabling: false,
		},
		{
			args:                []string{"child", "grandchild"},
			failWithoutEnabling: true,
		},
		{
			args:                []string{"CHIld", "GRANdchild"},
			failWithoutEnabling: true,
		},
		{
			args:                []string{"alternative", "ALIAS"},
			failWithoutEnabling: false,
		},
		{
			args:                []string{"alternative", "alias"},
			failWithoutEnabling: true,
		},
		{
			args:                []string{"CHILD", "alias"},
			failWithoutEnabling: true,
		},
		{
			args:                []string{"CHIld", "aliAS"},
			failWithoutEnabling: true,
		},
	}

	for _, test := range tests {
		for _, enableCaseInsensitivity := range []bool{true, false} {
			EnableCaseInsensitive = enableCaseInsensitivity

			output, err := executeCommand(rootCmd, test.args...)
			expectedFailure := test.failWithoutEnabling && !enableCaseInsensitivity

			if !expectedFailure && output != "" {
				t.Errorf("Unexpected output: %v", output)
			}
			if !expectedFailure && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		}
	}

	EnableCaseInsensitive = defaultCaseInsensitive
}

// TestCaseSensitivityBackwardCompatibility tests backward compatibility with respect to command names case sensitivity behavior.
func TestCaseSensitivityBackwardCompatibility(t *testing.T) {
	rootCmd := &Command{Use: "root", Run: emptyRun}
	childCmd := &Command{Use: "child", Run: emptyRun}

	rootCmd.AddCommand(childCmd)
	_, err := executeCommand(rootCmd, strings.ToUpper(childCmd.Use))
	if err == nil {
		t.Error("Expected error on calling a command in upper case while command names are case sensitive. Got nil.")
	}

}

// TestRemoveCommand tests the functionality of removing a command from a parent command and ensures that attempting to execute it raises an error.
func TestRemoveCommand(t *testing.T) {
	rootCmd := &Command{Use: "root", Args: NoArgs, Run: emptyRun}
	childCmd := &Command{Use: "child", Run: emptyRun}
	rootCmd.AddCommand(childCmd)
	rootCmd.RemoveCommand(childCmd)

	_, err := executeCommand(rootCmd, "child")
	if err == nil {
		t.Error("Expected error on calling removed command. Got nil.")
	}
}

// TestReplaceCommandWithRemove tests replacing a child command with another.
// It sets up a root command and two child commands, removes one child,
// and then adds another. It asserts that the removed command is not called
// while the new command is called when the child command is executed.
func TestReplaceCommandWithRemove(t *testing.T) {
	childUsed := 0
	rootCmd := &Command{Use: "root", Run: emptyRun}
	child1Cmd := &Command{
		Use: "child",
		Run: func(*Command, []string) { childUsed = 1 },
	}
	child2Cmd := &Command{
		Use: "child",
		Run: func(*Command, []string) { childUsed = 2 },
	}
	rootCmd.AddCommand(child1Cmd)
	rootCmd.RemoveCommand(child1Cmd)
	rootCmd.AddCommand(child2Cmd)

	output, err := executeCommand(rootCmd, "child")
	if output != "" {
		t.Errorf("Unexpected output: %v", output)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if childUsed == 1 {
		t.Error("Removed command shouldn't be called")
	}
	if childUsed != 2 {
		t.Error("Replacing command should have been called but didn't")
	}
}

// TestDeprecatedCommand tests the execution of a deprecated command and verifies that the deprecation message is displayed.
// It creates a root command with a deprecated subcommand, executes the deprecated command, and checks if the deprecation message is present in the output.
func TestDeprecatedCommand(t *testing.T) {
	rootCmd := &Command{Use: "root", Run: emptyRun}
	deprecatedCmd := &Command{
		Use:        "deprecated",
		Deprecated: "This command is deprecated",
		Run:        emptyRun,
	}
	rootCmd.AddCommand(deprecatedCmd)

	output, err := executeCommand(rootCmd, "deprecated")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringContains(t, output, deprecatedCmd.Deprecated)
}

// TestHooks tests the hook functionality of a command by executing it and verifying that all hooks are called with the correct arguments.
func TestHooks(t *testing.T) {
	var (
		persPreArgs  string
		preArgs      string
		runArgs      string
		postArgs     string
		persPostArgs string
	)

	c := &Command{
		Use: "c",
		PersistentPreRun: func(_ *Command, args []string) {
			persPreArgs = strings.Join(args, " ")
		},
		PreRun: func(_ *Command, args []string) {
			preArgs = strings.Join(args, " ")
		},
		Run: func(_ *Command, args []string) {
			runArgs = strings.Join(args, " ")
		},
		PostRun: func(_ *Command, args []string) {
			postArgs = strings.Join(args, " ")
		},
		PersistentPostRun: func(_ *Command, args []string) {
			persPostArgs = strings.Join(args, " ")
		},
	}

	output, err := executeCommand(c, "one", "two")
	if output != "" {
		t.Errorf("Unexpected output: %v", output)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	for _, v := range []struct {
		name string
		got  string
	}{
		{"persPreArgs", persPreArgs},
		{"preArgs", preArgs},
		{"runArgs", runArgs},
		{"postArgs", postArgs},
		{"persPostArgs", persPostArgs},
	} {
		if v.got != onetwo {
			t.Errorf("Expected %s %q, got %q", v.name, onetwo, v.got)
		}
	}
}

// TestPersistentHooks tests the behavior of persistent hooks in various scenarios.
//
// It enables and disables the EnableTraverseRunHooks flag to test how persistent pre-run, run,
// and post-run hooks are invoked during command execution. The function asserts that the correct
// sequence of hooks is executed based on the traversal enabled status.
//
// Parameters:
// - t: A testing.T instance used for assertions and reporting errors in tests.
//
// The function does not return any values; it executes as part of a test suite.
func TestPersistentHooks(t *testing.T) {
	EnableTraverseRunHooks = true
	testPersistentHooks(t, []string{
		"parent PersistentPreRun",
		"child PersistentPreRun",
		"child PreRun",
		"child Run",
		"child PostRun",
		"child PersistentPostRun",
		"parent PersistentPostRun",
	})

	EnableTraverseRunHooks = false
	testPersistentHooks(t, []string{
		"child PersistentPreRun",
		"child PreRun",
		"child Run",
		"child PostRun",
		"child PersistentPostRun",
	})
}

// testPersistentHooks tests the execution order of persistent hooks in a command hierarchy.
// It validates that the expected hook run order is followed and that no unexpected output or errors occur.
// Parameters:
// - t: A testing.T instance for reporting errors.
// - expectedHookRunOrder: A slice of strings representing the expected order in which hooks should be run.
func testPersistentHooks(t *testing.T, expectedHookRunOrder []string) {
	var hookRunOrder []string

	validateHook := func(args []string, hookName string) {
		hookRunOrder = append(hookRunOrder, hookName)
		got := strings.Join(args, " ")
		if onetwo != got {
			t.Errorf("Expected %s %q, got %q", hookName, onetwo, got)
		}
	}

	parentCmd := &Command{
		Use: "parent",
		PersistentPreRun: func(_ *Command, args []string) {
			validateHook(args, "parent PersistentPreRun")
		},
		PreRun: func(_ *Command, args []string) {
			validateHook(args, "parent PreRun")
		},
		Run: func(_ *Command, args []string) {
			validateHook(args, "parent Run")
		},
		PostRun: func(_ *Command, args []string) {
			validateHook(args, "parent PostRun")
		},
		PersistentPostRun: func(_ *Command, args []string) {
			validateHook(args, "parent PersistentPostRun")
		},
	}

	childCmd := &Command{
		Use: "child",
		PersistentPreRun: func(_ *Command, args []string) {
			validateHook(args, "child PersistentPreRun")
		},
		PreRun: func(_ *Command, args []string) {
			validateHook(args, "child PreRun")
		},
		Run: func(_ *Command, args []string) {
			validateHook(args, "child Run")
		},
		PostRun: func(_ *Command, args []string) {
			validateHook(args, "child PostRun")
		},
		PersistentPostRun: func(_ *Command, args []string) {
			validateHook(args, "child PersistentPostRun")
		},
	}
	parentCmd.AddCommand(childCmd)

	output, err := executeCommand(parentCmd, "child", "one", "two")
	if output != "" {
		t.Errorf("Unexpected output: %v", output)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	for idx, exp := range expectedHookRunOrder {
		if len(hookRunOrder) > idx {
			if act := hookRunOrder[idx]; act != exp {
				t.Errorf("Expected %q at %d, got %q", exp, idx, act)
			}
		} else {
			t.Errorf("Expected %q at %d, got nothing", exp, idx)
		}
	}
}

// TestGlobalNormFuncPropagation tests that setting a global normalization function on a parent command propagates to its child commands.
// It uses a custom normalization function and verifies that it is correctly applied to both the parent and child commands.
func TestGlobalNormFuncPropagation(t *testing.T) {
	normFunc := func(f *pflag.FlagSet, name string) pflag.NormalizedName {
		return pflag.NormalizedName(name)
	}

	rootCmd := &Command{Use: "root", Run: emptyRun}
	childCmd := &Command{Use: "child", Run: emptyRun}
	rootCmd.AddCommand(childCmd)

	rootCmd.SetGlobalNormalizationFunc(normFunc)
	if reflect.ValueOf(normFunc).Pointer() != reflect.ValueOf(rootCmd.GlobalNormalizationFunc()).Pointer() {
		t.Error("rootCmd seems to have a wrong normalization function")
	}

	if reflect.ValueOf(normFunc).Pointer() != reflect.ValueOf(childCmd.GlobalNormalizationFunc()).Pointer() {
		t.Error("childCmd should have had the normalization function of rootCmd")
	}
}

// TestNormPassedOnLocal tests if the normalization function is passed to the local flag set.
func TestNormPassedOnLocal(t *testing.T) {
	toUpper := func(f *pflag.FlagSet, name string) pflag.NormalizedName {
		return pflag.NormalizedName(strings.ToUpper(name))
	}

	c := &Command{}
	c.Flags().Bool("flagname", true, "this is a dummy flag")
	c.SetGlobalNormalizationFunc(toUpper)
	if c.LocalFlags().Lookup("flagname") != c.LocalFlags().Lookup("FLAGNAME") {
		t.Error("Normalization function should be passed on to Local flag set")
	}
}

// TestNormPassedOnInherited tests that a normalization function is passed on to inherited flag sets when adding commands before and after flags. It ensures that the normalization logic is applied consistently across different commands and their flag sets.
// The test creates a root command with a custom normalization function, adds two child commands, and verifies that the normalization function affects the flag names in both child commands' inherited flag sets.
func TestNormPassedOnInherited(t *testing.T) {
	toUpper := func(f *pflag.FlagSet, name string) pflag.NormalizedName {
		return pflag.NormalizedName(strings.ToUpper(name))
	}

	c := &Command{}
	c.SetGlobalNormalizationFunc(toUpper)

	child1 := &Command{}
	c.AddCommand(child1)

	c.PersistentFlags().Bool("flagname", true, "")

	child2 := &Command{}
	c.AddCommand(child2)

	inherited := child1.InheritedFlags()
	if inherited.Lookup("flagname") == nil || inherited.Lookup("flagname") != inherited.Lookup("FLAGNAME") {
		t.Error("Normalization function should be passed on to inherited flag set in command added before flag")
	}

	inherited = child2.InheritedFlags()
	if inherited.Lookup("flagname") == nil || inherited.Lookup("flagname") != inherited.Lookup("FLAGNAME") {
		t.Error("Normalization function should be passed on to inherited flag set in command added after flag")
	}
}

// TestConsistentNormalizedName tests that setting different normalization functions does not lead to duplicate flags. It verifies that the global normalization function takes precedence and prevents creation of a flag with an already normalized name.
func TestConsistentNormalizedName(t *testing.T) {
	toUpper := func(f *pflag.FlagSet, name string) pflag.NormalizedName {
		return pflag.NormalizedName(strings.ToUpper(name))
	}
	n := func(f *pflag.FlagSet, name string) pflag.NormalizedName {
		return pflag.NormalizedName(name)
	}

	c := &Command{}
	c.Flags().Bool("flagname", true, "")
	c.SetGlobalNormalizationFunc(toUpper)
	c.SetGlobalNormalizationFunc(n)

	if c.LocalFlags().Lookup("flagname") == c.LocalFlags().Lookup("FLAGNAME") {
		t.Error("Normalizing flag names should not result in duplicate flags")
	}
}

// TestFlagOnPflagCommandLine tests if a pflag is added to the command line.
// It checks if the specified flag appears in the help output of the command.
func TestFlagOnPflagCommandLine(t *testing.T) {
	flagName := "flagOnCommandLine"
	pflag.String(flagName, "", "about my flag")

	c := &Command{Use: "c", Run: emptyRun}
	c.AddCommand(&Command{Use: "child", Run: emptyRun})

	output, _ := executeCommand(c, "--help")
	checkStringContains(t, output, flagName)

	resetCommandLineFlagSet()
}

// TestHiddenCommandExecutes checks if hidden commands run as intended.
func TestHiddenCommandExecutes(t *testing.T) {
	executed := false
	c := &Command{
		Use:    "c",
		Hidden: true,
		Run:    func(*Command, []string) { executed = true },
	}

	output, err := executeCommand(c)
	if output != "" {
		t.Errorf("Unexpected output: %v", output)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !executed {
		t.Error("Hidden command should have been executed")
	}
}

// TestHiddenCommandIsHidden tests that hidden commands do not appear in the usage or help text.
func TestHiddenCommandIsHidden(t *testing.T) {
	c := &Command{Use: "c", Hidden: true, Run: emptyRun}
	if c.IsAvailableCommand() {
		t.Errorf("Hidden command should be unavailable")
	}
}

// TestCommandsAreSorted tests that commands are sorted alphabetically when EnableCommandSorting is enabled.
func TestCommandsAreSorted(t *testing.T) {
	EnableCommandSorting = true

	originalNames := []string{"middle", "zlast", "afirst"}
	expectedNames := []string{"afirst", "middle", "zlast"}

	var rootCmd = &Command{Use: "root"}

	for _, name := range originalNames {
		rootCmd.AddCommand(&Command{Use: name})
	}

	for i, c := range rootCmd.Commands() {
		got := c.Name()
		if expectedNames[i] != got {
			t.Errorf("Expected: %s, got: %s", expectedNames[i], got)
		}
	}

	EnableCommandSorting = defaultCommandSorting
}

// TestEnableCommandSortingIsDisabled tests the scenario where command sorting is disabled.
// It ensures that commands are added in the order they were created without any sorting applied.
func TestEnableCommandSortingIsDisabled(t *testing.T) {
	EnableCommandSorting = false

	originalNames := []string{"middle", "zlast", "afirst"}

	var rootCmd = &Command{Use: "root"}

	for _, name := range originalNames {
		rootCmd.AddCommand(&Command{Use: name})
	}

	for i, c := range rootCmd.Commands() {
		got := c.Name()
		if originalNames[i] != got {
			t.Errorf("expected: %s, got: %s", originalNames[i], got)
		}
	}

	EnableCommandSorting = defaultCommandSorting
}

// TestUsageWithGroup tests the usage of a root command with groups and ensures that the help output is correctly grouped.
func TestUsageWithGroup(t *testing.T) {
	var rootCmd = &Command{Use: "root", Short: "test", Run: emptyRun}
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.AddGroup(&Group{ID: "group1", Title: "group1"})
	rootCmd.AddGroup(&Group{ID: "group2", Title: "group2"})

	rootCmd.AddCommand(&Command{Use: "cmd1", GroupID: "group1", Run: emptyRun})
	rootCmd.AddCommand(&Command{Use: "cmd2", GroupID: "group2", Run: emptyRun})

	output, err := executeCommand(rootCmd, "--help")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// help should be ungrouped here
	checkStringContains(t, output, "\nAdditional Commands:\n  help")
	checkStringContains(t, output, "\ngroup1\n  cmd1")
	checkStringContains(t, output, "\ngroup2\n  cmd2")
}

// TestUsageHelpGroup tests the usage of the help command with groups.
func TestUsageHelpGroup(t *testing.T) {
	var rootCmd = &Command{Use: "root", Short: "test", Run: emptyRun}
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.AddGroup(&Group{ID: "group", Title: "group"})
	rootCmd.AddCommand(&Command{Use: "xxx", GroupID: "group", Run: emptyRun})
	rootCmd.SetHelpCommandGroupID("group")

	output, err := executeCommand(rootCmd, "--help")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// now help should be grouped under "group"
	checkStringOmits(t, output, "\nAdditional Commands:\n  help")
	checkStringContains(t, output, "\ngroup\n  help")
}

// TestUsageCompletionGroup tests the usage of command completion groups in a root command.
func TestUsageCompletionGroup(t *testing.T) {
	var rootCmd = &Command{Use: "root", Short: "test", Run: emptyRun}

	rootCmd.AddGroup(&Group{ID: "group", Title: "group"})
	rootCmd.AddGroup(&Group{ID: "help", Title: "help"})

	rootCmd.AddCommand(&Command{Use: "xxx", GroupID: "group", Run: emptyRun})
	rootCmd.SetHelpCommandGroupID("help")
	rootCmd.SetCompletionCommandGroupID("group")

	output, err := executeCommand(rootCmd, "--help")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// now completion should be grouped under "group"
	checkStringOmits(t, output, "\nAdditional Commands:\n  completion")
	checkStringContains(t, output, "\ngroup\n  completion")
}

// TestUngroupedCommand tests the behavior of a root command with an ungrouped command.
func TestUngroupedCommand(t *testing.T) {
	var rootCmd = &Command{Use: "root", Short: "test", Run: emptyRun}

	rootCmd.AddGroup(&Group{ID: "group", Title: "group"})
	rootCmd.AddGroup(&Group{ID: "help", Title: "help"})

	rootCmd.AddCommand(&Command{Use: "xxx", GroupID: "group", Run: emptyRun})
	rootCmd.SetHelpCommandGroupID("help")
	rootCmd.SetCompletionCommandGroupID("group")

	// Add a command without a group
	rootCmd.AddCommand(&Command{Use: "yyy", Run: emptyRun})

	output, err := executeCommand(rootCmd, "--help")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// The yyy command should be in the additional command "group"
	checkStringContains(t, output, "\nAdditional Commands:\n  yyy")
}

// TestAddGroup tests the functionality of adding a group and a command to a root command.
func TestAddGroup(t *testing.T) {
	var rootCmd = &Command{Use: "root", Short: "test", Run: emptyRun}

	rootCmd.AddGroup(&Group{ID: "group", Title: "Test group"})
	rootCmd.AddCommand(&Command{Use: "cmd", GroupID: "group", Run: emptyRun})

	output, err := executeCommand(rootCmd, "--help")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringContains(t, output, "\nTest group\n  cmd")
}

// TestWrongGroupFirstLevel tests the scenario where a command is added to a non-existent group.
// It verifies that the system panics when attempting to run a command with an invalid group ID.
func TestWrongGroupFirstLevel(t *testing.T) {
	var rootCmd = &Command{Use: "root", Short: "test", Run: emptyRun}

	rootCmd.AddGroup(&Group{ID: "group", Title: "Test group"})
	// Use the wrong group ID
	rootCmd.AddCommand(&Command{Use: "cmd", GroupID: "wrong", Run: emptyRun})

	defer func() {
		if recover() == nil {
			t.Errorf("The code should have panicked due to a missing group")
		}
	}()
	_, err := executeCommand(rootCmd, "--help")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

// TestWrongGroupNestedLevel tests the behavior of adding a command to a non-existent group within a nested command structure.
//
// It sets up a root command with a child command and attempts to add a new command to a non-existent group. The test expects
// the code to panic when trying to execute the command, as the group does not exist. The test also verifies that an unexpected
// error is not returned when attempting to execute the command.
//
// Parameters:
//   - t: A testing.T instance for reporting test results.
func TestWrongGroupNestedLevel(t *testing.T) {
	var rootCmd = &Command{Use: "root", Short: "test", Run: emptyRun}
	var childCmd = &Command{Use: "child", Run: emptyRun}
	rootCmd.AddCommand(childCmd)

	childCmd.AddGroup(&Group{ID: "group", Title: "Test group"})
	// Use the wrong group ID
	childCmd.AddCommand(&Command{Use: "cmd", GroupID: "wrong", Run: emptyRun})

	defer func() {
		if recover() == nil {
			t.Errorf("The code should have panicked due to a missing group")
		}
	}()
	_, err := executeCommand(rootCmd, "child", "--help")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

// TestWrongGroupForHelp tests the scenario where a command help is requested using a non-existent group ID.
// It sets up a root command with a child command and attempts to set an invalid help command group ID.
// The test expects a panic due to the missing group and confirms that no error occurs during command execution.
func TestWrongGroupForHelp(t *testing.T) {
	var rootCmd = &Command{Use: "root", Short: "test", Run: emptyRun}
	var childCmd = &Command{Use: "child", Run: emptyRun}
	rootCmd.AddCommand(childCmd)

	rootCmd.AddGroup(&Group{ID: "group", Title: "Test group"})
	// Use the wrong group ID
	rootCmd.SetHelpCommandGroupID("wrong")

	defer func() {
		if recover() == nil {
			t.Errorf("The code should have panicked due to a missing group")
		}
	}()
	_, err := executeCommand(rootCmd, "--help")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

// TestWrongGroupForCompletion verifies that the system reacts correctly when setting a completion command group ID that does not exist.
// It creates a root command with a child command and attempts to set an invalid completion group ID, expecting a panic.
// If no panic occurs, it asserts an error. If an error is returned, it is unexpected and causes a test failure.
func TestWrongGroupForCompletion(t *testing.T) {
	var rootCmd = &Command{Use: "root", Short: "test", Run: emptyRun}
	var childCmd = &Command{Use: "child", Run: emptyRun}
	rootCmd.AddCommand(childCmd)

	rootCmd.AddGroup(&Group{ID: "group", Title: "Test group"})
	// Use the wrong group ID
	rootCmd.SetCompletionCommandGroupID("wrong")

	defer func() {
		if recover() == nil {
			t.Errorf("The code should have panicked due to a missing group")
		}
	}()
	_, err := executeCommand(rootCmd, "--help")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

// TestSetOutput tests the behavior of Setting the output to nil and verifying it reverts to stdout.
// It creates a new Command instance, sets its output to nil, and checks if calling OutOrStdout()
// returns os.Stdout as expected. If not, it fails the test with an error message.
func TestSetOutput(t *testing.T) {
	c := &Command{}
	c.SetOutput(nil)
	if out := c.OutOrStdout(); out != os.Stdout {
		t.Errorf("Expected setting output to nil to revert back to stdout")
	}
}

// TestSetOut tests the behavior of setting the output to nil in a Command instance and verifies that it reverts back to standard output.
// It uses a testing.T instance to perform assertions.
func TestSetOut(t *testing.T) {
	c := &Command{}
	c.SetOut(nil)
	if out := c.OutOrStdout(); out != os.Stdout {
		t.Errorf("Expected setting output to nil to revert back to stdout")
	}
}

// TestSetErr tests the behavior of the SetErr method.
// It verifies that setting an error to nil reverts the Command instance's ErrOrStderr() output to os.Stderr.
func TestSetErr(t *testing.T) {
	c := &Command{}
	c.SetErr(nil)
	if out := c.ErrOrStderr(); out != os.Stderr {
		t.Errorf("Expected setting error to nil to revert back to stderr")
	}
}

// TestSetIn tests the SetIn method of Command.
//
// It sets the input of a command to nil and checks if it reverts back to stdin. If not, it fails the test.
func TestSetIn(t *testing.T) {
	c := &Command{}
	c.SetIn(nil)
	if out := c.InOrStdin(); out != os.Stdin {
		t.Errorf("Expected setting input to nil to revert back to stdin")
	}
}

// TestUsageStringRedirected tests the functionality of capturing both standard output and standard error in UsageString.
// It ensures that when multiple Print and PrintErr calls are made, they are consolidated into a single UsageString.
func TestUsageStringRedirected(t *testing.T) {
	c := &Command{}

	c.usageFunc = func(cmd *Command) error {
		cmd.Print("[stdout1]")
		cmd.PrintErr("[stderr2]")
		cmd.Print("[stdout3]")
		return nil
	}

	expected := "[stdout1][stderr2][stdout3]"
	if got := c.UsageString(); got != expected {
		t.Errorf("Expected usage string to consider both stdout and stderr")
	}
}

// TestCommandPrintRedirection tests the print redirection functionality of the Command struct.
func TestCommandPrintRedirection(t *testing.T) {
	errBuff, outBuff := bytes.NewBuffer(nil), bytes.NewBuffer(nil)
	root := &Command{
		Run: func(cmd *Command, args []string) {

			cmd.PrintErr("PrintErr")
			cmd.PrintErrln("PrintErr", "line")
			cmd.PrintErrf("PrintEr%s", "r")

			cmd.Print("Print")
			cmd.Println("Print", "line")
			cmd.Printf("Prin%s", "t")
		},
	}

	root.SetErr(errBuff)
	root.SetOut(outBuff)

	if err := root.Execute(); err != nil {
		t.Error(err)
	}

	gotErrBytes, err := io.ReadAll(errBuff)
	if err != nil {
		t.Error(err)
	}

	gotOutBytes, err := io.ReadAll(outBuff)
	if err != nil {
		t.Error(err)
	}

	if wantErr := []byte("PrintErrPrintErr line\nPrintErr"); !bytes.Equal(gotErrBytes, wantErr) {
		t.Errorf("got: '%s' want: '%s'", gotErrBytes, wantErr)
	}

	if wantOut := []byte("PrintPrint line\nPrint"); !bytes.Equal(gotOutBytes, wantOut) {
		t.Errorf("got: '%s' want: '%s'", gotOutBytes, wantOut)
	}
}

// TestFlagErrorFunc tests the FlagErrorFunc of a Command.
// It sets a custom error function that formats an error message with "This is expected:" prefix.
// The test executes a command with an unknown flag and checks if the returned error matches the expected format.
func TestFlagErrorFunc(t *testing.T) {
	c := &Command{Use: "c", Run: emptyRun}

	expectedFmt := "This is expected: %v"
	c.SetFlagErrorFunc(func(_ *Command, err error) error {
		return fmt.Errorf(expectedFmt, err)
	})

	_, err := executeCommand(c, "--unknown-flag")

	got := err.Error()
	expected := fmt.Sprintf(expectedFmt, "unknown flag: --unknown-flag")
	if got != expected {
		t.Errorf("Expected %v, got %v", expected, got)
	}
}

// TestFlagErrorFuncHelp tests the FlagErrorFunc functionality by creating a command with a persistent flag and a custom error function.
// It then executes the command with both "--help" and "-h" flags to ensure they do not fail and return the expected output.
func TestFlagErrorFuncHelp(t *testing.T) {
	c := &Command{Use: "c", Run: emptyRun}
	c.PersistentFlags().Bool("help", false, "help for c")
	c.SetFlagErrorFunc(func(_ *Command, err error) error {
		return fmt.Errorf("wrap error: %w", err)
	})

	out, err := executeCommand(c, "--help")
	if err != nil {
		t.Errorf("--help should not fail: %v", err)
	}

	expected := `Usage:
  c [flags]

Flags:
      --help   help for c
`
	if out != expected {
		t.Errorf("Expected: %v, got: %v", expected, out)
	}

	out, err = executeCommand(c, "-h")
	if err != nil {
		t.Errorf("-h should not fail: %v", err)
	}

	if out != expected {
		t.Errorf("Expected: %v, got: %v", expected, out)
	}
}

// TestSortedFlags checks if cmd.LocalFlags() is unsorted when cmd.Flags().SortFlags set to false.
// This test is related to https://github.com/spf13/cobra/issues/404.
func TestSortedFlags(t *testing.T) {
	c := &Command{}
	c.Flags().SortFlags = false
	names := []string{"C", "B", "A", "D"}
	for _, name := range names {
		c.Flags().Bool(name, false, "")
	}

	i := 0
	c.LocalFlags().VisitAll(func(f *pflag.Flag) {
		if i == len(names) {
			return
		}
		if stringInSlice(f.Name, names) {
			if names[i] != f.Name {
				t.Errorf("Incorrect order. Expected %v, got %v", names[i], f.Name)
			}
			i++
		}
	})
}

// TestMergeCommandLineToFlags checks if pflag.CommandLine is correctly merged to c.Flags() after the first call of c.mergePersistentFlags.
// It verifies that flags from CommandLine are available in c.Flags().
// This function addresses issue https://github.com/spf13/cobra/issues/443.
//
// Parameters:
//   - t: The testing.T instance for running the test.
//
// Expected behavior:
//   - After merging, the flag "boolflag" should be present in c.Flags().
func TestMergeCommandLineToFlags(t *testing.T) {
	pflag.Bool("boolflag", false, "")
	c := &Command{Use: "c", Run: emptyRun}
	c.mergePersistentFlags()
	if c.Flags().Lookup("boolflag") == nil {
		t.Fatal("Expecting to have flag from CommandLine in c.Flags()")
	}

	resetCommandLineFlagSet()
}

// TestUseDeprecatedFlags checks if cobra.Execute() prints a message when a deprecated flag is used.
// The function tests the behavior of Cobra's Flag system when a deprecated flag is utilized and
// verifies that a specific deprecation message is printed in the output. This test addresses issue #463 from the Cobra repository.
//
// Parameters:
//   - t: A testing.T instance to provide context for assertions and error handling during the test.
//
// Returns:
//   None
func TestUseDeprecatedFlags(t *testing.T) {
	c := &Command{Use: "c", Run: emptyRun}
	c.Flags().BoolP("deprecated", "d", false, "deprecated flag")
	assertNoErr(t, c.Flags().MarkDeprecated("deprecated", "This flag is deprecated"))

	output, err := executeCommand(c, "c", "-d")
	if err != nil {
		t.Error("Unexpected error:", err)
	}
	checkStringContains(t, output, "This flag is deprecated")
}

// TestTraverseWithParentFlags tests the Traverse method of a Command with parent flags.
//
// It creates a root command and a child command, sets up flags on both, adds the child to the root,
// and then traverses the command tree with specific flag arguments. The test checks if the traversal
// is performed correctly, including handling of flags from parent commands.
//
// Parameters:
//   - t: A *testing.T instance for running the test.
//
// Returns:
//   None
//
// Errors:
//   This function does not return any errors directly but uses t.Errorf to report failures.
func TestTraverseWithParentFlags(t *testing.T) {
	rootCmd := &Command{Use: "root", TraverseChildren: true}
	rootCmd.Flags().String("str", "", "")
	rootCmd.Flags().BoolP("bool", "b", false, "")

	childCmd := &Command{Use: "child"}
	childCmd.Flags().Int("int", -1, "")

	rootCmd.AddCommand(childCmd)

	c, args, err := rootCmd.Traverse([]string{"-b", "--str", "ok", "child", "--int"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(args) != 1 && args[0] != "--add" {
		t.Errorf("Wrong args: %v", args)
	}
	if c.Name() != childCmd.Name() {
		t.Errorf("Expected command: %q, got: %q", childCmd.Name(), c.Name())
	}
}

// TestTraverseNoParentFlags tests the Traverse method of a Command with no parent flags.
//
// It sets up a root command with a flag and a child command without any flags. The test then calls
// the Traverse method on the root command to navigate to the child command using the path "child".
// It verifies that there are no arguments returned, that the command name is correct, and that no error
// occurs during the traversal.
func TestTraverseNoParentFlags(t *testing.T) {
	rootCmd := &Command{Use: "root", TraverseChildren: true}
	rootCmd.Flags().String("foo", "", "foo things")

	childCmd := &Command{Use: "child"}
	childCmd.Flags().String("str", "", "")
	rootCmd.AddCommand(childCmd)

	c, args, err := rootCmd.Traverse([]string{"child"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(args) != 0 {
		t.Errorf("Wrong args %v", args)
	}
	if c.Name() != childCmd.Name() {
		t.Errorf("Expected command: %q, got: %q", childCmd.Name(), c.Name())
	}
}

// TestTraverseWithBadParentFlags tests the behavior of the Traverse method when encountering unknown flags.
//
// It creates a root command with a child command and attempts to traverse with a bad parent flag.
// The expected error message is checked, and it's verified that no command is returned.
func TestTraverseWithBadParentFlags(t *testing.T) {
	rootCmd := &Command{Use: "root", TraverseChildren: true}

	childCmd := &Command{Use: "child"}
	childCmd.Flags().String("str", "", "")
	rootCmd.AddCommand(childCmd)

	expected := "unknown flag: --str"

	c, _, err := rootCmd.Traverse([]string{"--str", "ok", "child"})
	if err == nil || !strings.Contains(err.Error(), expected) {
		t.Errorf("Expected error, %q, got %q", expected, err)
	}
	if c != nil {
		t.Errorf("Expected nil command")
	}
}

// TestTraverseWithBadChildFlag tests the behavior of the Traverse method when a bad flag is provided to a child command.
// It sets up a root command with a child command and expects that the Traverse method returns the correct child command and the remaining args without parsing the flags.
func TestTraverseWithBadChildFlag(t *testing.T) {
	rootCmd := &Command{Use: "root", TraverseChildren: true}
	rootCmd.Flags().String("str", "", "")

	childCmd := &Command{Use: "child"}
	rootCmd.AddCommand(childCmd)

	// Expect no error because the last commands args shouldn't be parsed in
	// Traverse.
	c, args, err := rootCmd.Traverse([]string{"child", "--str"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(args) != 1 && args[0] != "--str" {
		t.Errorf("Wrong args: %v", args)
	}
	if c.Name() != childCmd.Name() {
		t.Errorf("Expected command %q, got: %q", childCmd.Name(), c.Name())
	}
}

// TestTraverseWithTwoSubcommands tests the Traverse method when navigating through a command tree with two levels of subcommands. It verifies that the traversal correctly reaches the deepest subcommand and returns it without errors.
func TestTraverseWithTwoSubcommands(t *testing.T) {
	rootCmd := &Command{Use: "root", TraverseChildren: true}

	subCmd := &Command{Use: "sub", TraverseChildren: true}
	rootCmd.AddCommand(subCmd)

	subsubCmd := &Command{
		Use: "subsub",
	}
	subCmd.AddCommand(subsubCmd)

	c, _, err := rootCmd.Traverse([]string{"sub", "subsub"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if c.Name() != subsubCmd.Name() {
		t.Fatalf("Expected command: %q, got %q", subsubCmd.Name(), c.Name())
	}
}

// TestUpdateName verifies that the Name method updates when the Use field of a Command is modified.
// This test addresses issue #422 regarding command name updating in Cobra.
func TestUpdateName(t *testing.T) {
	c := &Command{Use: "name xyz"}
	originalName := c.Name()

	c.Use = "changedName abc"
	if originalName == c.Name() || c.Name() != "changedName" {
		t.Error("c.Name() should be updated on changed c.Use")
	}
}

type calledAsTestcase struct {
	args []string
	call string
	want string
	epm  bool
}

// test runs the command with the given arguments and asserts that the expected command was called.
// It uses a mock function to capture the called command and its name.
// If the expected command is not called or the CalledAs method does not return the expected value, it fails the test.
func (tc *calledAsTestcase) test(t *testing.T) {
	defer func(ov bool) { EnablePrefixMatching = ov }(EnablePrefixMatching)
	EnablePrefixMatching = tc.epm

	var called *Command
	run := func(c *Command, _ []string) { t.Logf("called: %q", c.Name()); called = c }

	parent := &Command{Use: "parent", Run: run}
	child1 := &Command{Use: "child1", Run: run, Aliases: []string{"this"}}
	child2 := &Command{Use: "child2", Run: run, Aliases: []string{"that"}}

	parent.AddCommand(child1)
	parent.AddCommand(child2)
	parent.SetArgs(tc.args)

	output := new(bytes.Buffer)
	parent.SetOut(output)
	parent.SetErr(output)

	_ = parent.Execute()

	if called == nil {
		if tc.call != "" {
			t.Errorf("missing expected call to command: %s", tc.call)
		}
		return
	}

	if called.Name() != tc.call {
		t.Errorf("called command == %q; Wanted %q", called.Name(), tc.call)
	} else if got := called.CalledAs(); got != tc.want {
		t.Errorf("%s.CalledAs() == %q; Wanted: %q", tc.call, got, tc.want)
	}
}

// TestCalledAs runs a series of test cases to verify the behavior of the calledAs function.
func TestCalledAs(t *testing.T) {
	tests := map[string]calledAsTestcase{
		"find/no-args":            {nil, "parent", "parent", false},
		"find/real-name":          {[]string{"child1"}, "child1", "child1", false},
		"find/full-alias":         {[]string{"that"}, "child2", "that", false},
		"find/part-no-prefix":     {[]string{"thi"}, "", "", false},
		"find/part-alias":         {[]string{"thi"}, "child1", "this", true},
		"find/conflict":           {[]string{"th"}, "", "", true},
		"traverse/no-args":        {nil, "parent", "parent", false},
		"traverse/real-name":      {[]string{"child1"}, "child1", "child1", false},
		"traverse/full-alias":     {[]string{"that"}, "child2", "that", false},
		"traverse/part-no-prefix": {[]string{"thi"}, "", "", false},
		"traverse/part-alias":     {[]string{"thi"}, "child1", "this", true},
		"traverse/conflict":       {[]string{"th"}, "", "", true},
	}

	for name, tc := range tests {
		t.Run(name, tc.test)
	}
}

// TestFParseErrWhitelistBackwardCompatibility tests the backward compatibility of the fparse error handling when encountering an unknown flag.
// It creates a command with a boolean flag and executes it with an unknown flag. The test expects an error indicating an unknown flag and checks if the output contains the expected error message.
// Parameters:
//   - t: *testing.T, the testing environment
// Returns:
//   None
func TestFParseErrWhitelistBackwardCompatibility(t *testing.T) {
	c := &Command{Use: "c", Run: emptyRun}
	c.Flags().BoolP("boola", "a", false, "a boolean flag")

	output, err := executeCommand(c, "c", "-a", "--unknown", "flag")
	if err == nil {
		t.Error("expected unknown flag error")
	}
	checkStringContains(t, output, "unknown flag: --unknown")
}

// TestFParseErrWhitelistSameCommand tests the behavior of FParseErrWhitelist when encountering unknown flags.
// It creates a command with a flag whitelist that allows unknown flags and attempts to execute the command with an unknown flag.
// The test expects no error to be returned, as the unknown flag is whitelisted.
func TestFParseErrWhitelistSameCommand(t *testing.T) {
	c := &Command{
		Use: "c",
		Run: emptyRun,
		FParseErrWhitelist: FParseErrWhitelist{
			UnknownFlags: true,
		},
	}
	c.Flags().BoolP("boola", "a", false, "a boolean flag")

	_, err := executeCommand(c, "c", "-a", "--unknown", "flag")
	if err != nil {
		t.Error("unexpected error: ", err)
	}
}

// TestFParseErrWhitelistParentCommand tests the FParseErrWhitelist for a parent command.
// It creates a root command with a whitelist that allows unknown flags and adds a child command with its own flags.
// The test executes the command with an unknown flag, expecting an error indicating an unknown flag.
func TestFParseErrWhitelistParentCommand(t *testing.T) {
	root := &Command{
		Use: "root",
		Run: emptyRun,
		FParseErrWhitelist: FParseErrWhitelist{
			UnknownFlags: true,
		},
	}

	c := &Command{
		Use: "child",
		Run: emptyRun,
	}
	c.Flags().BoolP("boola", "a", false, "a boolean flag")

	root.AddCommand(c)

	output, err := executeCommand(root, "child", "-a", "--unknown", "flag")
	if err == nil {
		t.Error("expected unknown flag error")
	}
	checkStringContains(t, output, "unknown flag: --unknown")
}

// TestFParseErrWhitelistChildCommand tests the FParseErrWhitelist functionality for a child command.
func TestFParseErrWhitelistChildCommand(t *testing.T) {
	root := &Command{
		Use: "root",
		Run: emptyRun,
	}

	c := &Command{
		Use: "child",
		Run: emptyRun,
		FParseErrWhitelist: FParseErrWhitelist{
			UnknownFlags: true,
		},
	}
	c.Flags().BoolP("boola", "a", false, "a boolean flag")

	root.AddCommand(c)

	_, err := executeCommand(root, "child", "-a", "--unknown", "flag")
	if err != nil {
		t.Error("unexpected error: ", err.Error())
	}
}

// TestFParseErrWhitelistSiblingCommand tests the FParseErrWhitelist feature when sibling commands are involved.
func TestFParseErrWhitelistSiblingCommand(t *testing.T) {
	root := &Command{
		Use: "root",
		Run: emptyRun,
	}

	c := &Command{
		Use: "child",
		Run: emptyRun,
		FParseErrWhitelist: FParseErrWhitelist{
			UnknownFlags: true,
		},
	}
	c.Flags().BoolP("boola", "a", false, "a boolean flag")

	s := &Command{
		Use: "sibling",
		Run: emptyRun,
	}
	s.Flags().BoolP("boolb", "b", false, "a boolean flag")

	root.AddCommand(c)
	root.AddCommand(s)

	output, err := executeCommand(root, "sibling", "-b", "--unknown", "flag")
	if err == nil {
		t.Error("expected unknown flag error")
	}
	checkStringContains(t, output, "unknown flag: --unknown")
}

// TestSetContext tests the SetContext method of a Command struct by setting a value in the context and verifying that it can be retrieved within the Run function.
// It takes a testing.T pointer as an argument, which is used for assertion and error reporting during the test.
func TestSetContext(t *testing.T) {
	type key struct{}
	val := "foobar"
	root := &Command{
		Use: "root",
		Run: func(cmd *Command, args []string) {
			key := cmd.Context().Value(key{})
			got, ok := key.(string)
			if !ok {
				t.Error("key not found in context")
			}
			if got != val {
				t.Errorf("Expected value: \n %v\nGot:\n %v\n", val, got)
			}
		},
	}

	ctx := context.WithValue(context.Background(), key{}, val)
	root.SetContext(ctx)
	err := root.Execute()
	if err != nil {
		t.Error(err)
	}
}

// TestSetContextPreRun tests that the PreRun function sets a value in the context before running the command.
func TestSetContextPreRun(t *testing.T) {
	type key struct{}
	val := "barr"
	root := &Command{
		Use: "root",
		PreRun: func(cmd *Command, args []string) {
			ctx := context.WithValue(cmd.Context(), key{}, val)
			cmd.SetContext(ctx)
		},
		Run: func(cmd *Command, args []string) {
			val := cmd.Context().Value(key{})
			got, ok := val.(string)
			if !ok {
				t.Error("key not found in context")
			}
			if got != val {
				t.Errorf("Expected value: \n %v\nGot:\n %v\n", val, got)
			}
		},
	}
	err := root.Execute()
	if err != nil {
		t.Error(err)
	}
}

// TestSetContextPreRunOverwrite tests that setting a context with a key and value in the Run method overwrites any existing value for that key.
// It checks if the expected error is returned when trying to access the overwritten key in the context.
func TestSetContextPreRunOverwrite(t *testing.T) {
	type key struct{}
	val := "blah"
	root := &Command{
		Use: "root",
		Run: func(cmd *Command, args []string) {
			key := cmd.Context().Value(key{})
			_, ok := key.(string)
			if ok {
				t.Error("key found in context when not expected")
			}
		},
	}
	ctx := context.WithValue(context.Background(), key{}, val)
	root.SetContext(ctx)
	err := root.ExecuteContext(context.Background())
	if err != nil {
		t.Error(err)
	}
}

// TestSetContextPersistentPreRun tests if the PersistentPreRun function correctly sets a context value that is accessible to its child command.
func TestSetContextPersistentPreRun(t *testing.T) {
	type key struct{}
	val := "barbar"
	root := &Command{
		Use: "root",
		PersistentPreRun: func(cmd *Command, args []string) {
			ctx := context.WithValue(cmd.Context(), key{}, val)
			cmd.SetContext(ctx)
		},
	}
	child := &Command{
		Use: "child",
		Run: func(cmd *Command, args []string) {
			key := cmd.Context().Value(key{})
			got, ok := key.(string)
			if !ok {
				t.Error("key not found in context")
			}
			if got != val {
				t.Errorf("Expected value: \n %v\nGot:\n %v\n", val, got)
			}
		},
	}
	root.AddCommand(child)
	root.SetArgs([]string{"child"})
	err := root.Execute()
	if err != nil {
		t.Error(err)
	}
}

const VersionFlag = "--version"
const HelpFlag = "--help"

// TestNoRootRunCommandExecutedWithVersionSet tests that when a command without a root run function is executed with version set, the appropriate output is produced.
//
// Parameters:
//   - t: A testing.T instance for running the test and reporting errors.
//
// The function sets up a Command tree with a root command that has a version but no run function. It then executes this command and checks if the output contains the long description, help flag, and version flag as expected. If any of these are missing or an error occurs during execution, the test will fail.
func TestNoRootRunCommandExecutedWithVersionSet(t *testing.T) {
	rootCmd := &Command{Use: "root", Version: "1.0.0", Long: "Long description"}
	rootCmd.AddCommand(&Command{Use: "child", Run: emptyRun})

	output, err := executeCommand(rootCmd)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringContains(t, output, rootCmd.Long)
	checkStringContains(t, output, HelpFlag)
	checkStringContains(t, output, VersionFlag)
}

// TestNoRootRunCommandExecutedWithoutVersionSet verifies that the root command is executed without a version flag set.
func TestNoRootRunCommandExecutedWithoutVersionSet(t *testing.T) {
	rootCmd := &Command{Use: "root", Long: "Long description"}
	rootCmd.AddCommand(&Command{Use: "child", Run: emptyRun})

	output, err := executeCommand(rootCmd)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringContains(t, output, rootCmd.Long)
	checkStringContains(t, output, HelpFlag)
	checkStringOmits(t, output, VersionFlag)
}

// TestHelpCommandExecutedWithVersionSet tests that the help command is executed with version set.
func TestHelpCommandExecutedWithVersionSet(t *testing.T) {
	rootCmd := &Command{Use: "root", Version: "1.0.0", Long: "Long description", Run: emptyRun}
	rootCmd.AddCommand(&Command{Use: "child", Run: emptyRun})

	output, err := executeCommand(rootCmd, "help")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringContains(t, output, rootCmd.Long)
	checkStringContains(t, output, HelpFlag)
	checkStringContains(t, output, VersionFlag)
}

// TestHelpCommandExecutedWithoutVersionSet tests the scenario where the help command is executed without a version flag set.
//
// It initializes a root command with a child command and then executes the "help" command. The function asserts that the output contains the long description of the root command and the help flag, but does not contain the version flag.
func TestHelpCommandExecutedWithoutVersionSet(t *testing.T) {
	rootCmd := &Command{Use: "root", Long: "Long description", Run: emptyRun}
	rootCmd.AddCommand(&Command{Use: "child", Run: emptyRun})

	output, err := executeCommand(rootCmd, "help")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringContains(t, output, rootCmd.Long)
	checkStringContains(t, output, HelpFlag)
	checkStringOmits(t, output, VersionFlag)
}

// TestHelpflagCommandExecutedWithVersionSet tests that the help flag command is executed when the version flag is set.
// It verifies that the output contains the root command's long description, the help flag, and the version flag.
func TestHelpflagCommandExecutedWithVersionSet(t *testing.T) {
	rootCmd := &Command{Use: "root", Version: "1.0.0", Long: "Long description", Run: emptyRun}
	rootCmd.AddCommand(&Command{Use: "child", Run: emptyRun})

	output, err := executeCommand(rootCmd, HelpFlag)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringContains(t, output, rootCmd.Long)
	checkStringContains(t, output, HelpFlag)
	checkStringContains(t, output, VersionFlag)
}

// TestHelpflagCommandExecutedWithoutVersionSet tests that the help flag is executed without a version set.
//
// Parameters:
//   - t: A testing.T instance for running test assertions.
//
// Returns:
//   - None
func TestHelpflagCommandExecutedWithoutVersionSet(t *testing.T) {
	rootCmd := &Command{Use: "root", Long: "Long description", Run: emptyRun}
	rootCmd.AddCommand(&Command{Use: "child", Run: emptyRun})

	output, err := executeCommand(rootCmd, HelpFlag)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringContains(t, output, rootCmd.Long)
	checkStringContains(t, output, HelpFlag)
	checkStringOmits(t, output, VersionFlag)
}

// TestFind tests the Find method of the Command type.
//
// It verifies that the method correctly identifies the child command and collects any flags passed to it.
func TestFind(t *testing.T) {
	var foo, bar string
	root := &Command{
		Use: "root",
	}
	root.PersistentFlags().StringVarP(&foo, "foo", "f", "", "")
	root.PersistentFlags().StringVarP(&bar, "bar", "b", "something", "")

	child := &Command{
		Use: "child",
	}
	root.AddCommand(child)

	testCases := []struct {
		args              []string
		expectedFoundArgs []string
	}{
		{
			[]string{"child"},
			[]string{},
		},
		{
			[]string{"child", "child"},
			[]string{"child"},
		},
		{
			[]string{"child", "foo", "child", "bar", "child", "baz", "child"},
			[]string{"foo", "child", "bar", "child", "baz", "child"},
		},
		{
			[]string{"-f", "child", "child"},
			[]string{"-f", "child"},
		},
		{
			[]string{"child", "-f", "child"},
			[]string{"-f", "child"},
		},
		{
			[]string{"-b", "child", "child"},
			[]string{"-b", "child"},
		},
		{
			[]string{"child", "-b", "child"},
			[]string{"-b", "child"},
		},
		{
			[]string{"child", "-b"},
			[]string{"-b"},
		},
		{
			[]string{"-b", "-f", "child", "child"},
			[]string{"-b", "-f", "child"},
		},
		{
			[]string{"-f", "child", "-b", "something", "child"},
			[]string{"-f", "child", "-b", "something"},
		},
		{
			[]string{"-f", "child", "child", "-b"},
			[]string{"-f", "child", "-b"},
		},
		{
			[]string{"-f=child", "-b=something", "child"},
			[]string{"-f=child", "-b=something"},
		},
		{
			[]string{"--foo", "child", "--bar", "something", "child"},
			[]string{"--foo", "child", "--bar", "something"},
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%v", tc.args), func(t *testing.T) {
			cmd, foundArgs, err := root.Find(tc.args)
			if err != nil {
				t.Fatal(err)
			}

			if cmd != child {
				t.Fatal("Expected cmd to be child, but it was not")
			}

			if !reflect.DeepEqual(tc.expectedFoundArgs, foundArgs) {
				t.Fatalf("Wrong args\nExpected: %v\nGot: %v", tc.expectedFoundArgs, foundArgs)
			}
		})
	}
}

// TestUnknownFlagShouldReturnSameErrorRegardlessOfArgPosition tests that an unknown flag returns the same error regardless of its position in the argument list.
func TestUnknownFlagShouldReturnSameErrorRegardlessOfArgPosition(t *testing.T) {
	testCases := [][]string{
		// {"--unknown", "--namespace", "foo", "child", "--bar"}, // FIXME: This test case fails, returning the error `unknown command "foo" for "root"` instead of the expected error `unknown flag: --unknown`
		{"--namespace", "foo", "--unknown", "child", "--bar"},
		{"--namespace", "foo", "child", "--unknown", "--bar"},
		{"--namespace", "foo", "child", "--bar", "--unknown"},

		{"--unknown", "--namespace=foo", "child", "--bar"},
		{"--namespace=foo", "--unknown", "child", "--bar"},
		{"--namespace=foo", "child", "--unknown", "--bar"},
		{"--namespace=foo", "child", "--bar", "--unknown"},

		{"--unknown", "--namespace=foo", "child", "--bar=true"},
		{"--namespace=foo", "--unknown", "child", "--bar=true"},
		{"--namespace=foo", "child", "--unknown", "--bar=true"},
		{"--namespace=foo", "child", "--bar=true", "--unknown"},
	}

	root := &Command{
		Use: "root",
		Run: emptyRun,
	}
	root.PersistentFlags().String("namespace", "", "a string flag")

	c := &Command{
		Use: "child",
		Run: emptyRun,
	}
	c.Flags().Bool("bar", false, "a boolean flag")

	root.AddCommand(c)

	for _, tc := range testCases {
		t.Run(strings.Join(tc, " "), func(t *testing.T) {
			output, err := executeCommand(root, tc...)
			if err == nil {
				t.Error("expected unknown flag error")
			}
			checkStringContains(t, output, "unknown flag: --unknown")
		})
	}
}

// TestHelpFuncExecuted tests if the help function is executed correctly when called with a specific context. It checks both the content of the help text and whether the correct context is being used.
//
// Parameters:
//   - t: The testing.T object for running the test.
//
// Returns:
//   None
func TestHelpFuncExecuted(t *testing.T) {
	helpText := "Long description"

	// Create a context that will be unique, not just the background context
	//nolint:golint,staticcheck // We can safely use a basic type as key in tests.
	executionCtx := context.WithValue(context.Background(), "testKey", "123")

	child := &Command{Use: "child", Run: emptyRun}
	child.SetHelpFunc(func(cmd *Command, args []string) {
		_, err := cmd.OutOrStdout().Write([]byte(helpText))
		if err != nil {
			t.Error(err)
		}

		// Test for https://github.com/spf13/cobra/issues/2240
		if cmd.Context() != executionCtx {
			t.Error("Context doesn't equal the execution context")
		}
	})

	rootCmd := &Command{Use: "root", Run: emptyRun}
	rootCmd.AddCommand(child)

	output, err := executeCommandWithContext(executionCtx, rootCmd, "help", "child")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringContains(t, output, helpText)
}
