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

// getCommand returns a Command object configured based on the provided PositionalArgs and a boolean indicating whether to include valid arguments.
// If withValid is true, the returned command will have ValidArgs set to ["one", "two", "three"].
// The Run field of the Command is initialized to emptyRun, which does nothing when called.
// Args determines what arguments are valid for this command.
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

// expectSuccess checks if the provided output and error are as expected. If output is not empty, it logs an error. If there is an error, it fails the test with a fatal error.
// Parameters:
//   - output: The string output to check.
//   - err: The error to check.
//   - t: The testing.T instance used for logging errors and failing tests.
// Returns: no value.
func expectSuccess(output string, err error, t *testing.T) {
	if output != "" {
		t.Errorf("Unexpected output: %v", output)
	}
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

// validOnlyWithInvalidArgs checks if the given error is not nil and has the expected error message.
// It takes an error and a testing.T instance as parameters. If the error is nil, it calls t.Fatal with a message indicating that an error was expected.
// It then compares the actual error message with the expected one and calls t.Errorf if they do not match.
func validOnlyWithInvalidArgs(err error, t *testing.T) {
	if err == nil {
		t.Fatal("Expected an error")
	}
	got := err.Error()
	expected := `invalid argument "a" for "c"`
	if got != expected {
		t.Errorf("Expected: %q, got: %q", expected, got)
	}
}

// noArgsWithArgs checks if an error is returned when calling a function with no arguments and the specified argument.
// It asserts that the error message matches the expected value for the given command.
// If no error is returned, it calls t.Fatal with an appropriate message.
// Parameters:
//   - err: The error to check.
//   - t: A testing.T instance used for assertions.
//   - arg: The argument passed to the function that triggered the error.
func noArgsWithArgs(err error, t *testing.T, arg string) {
	if err == nil {
		t.Fatal("Expected an error")
	}
	got := err.Error()
	expected := `unknown command "` + arg + `" for "c"`
	if got != expected {
		t.Errorf("Expected: %q, got: %q", expected, got)
	}
}

// minimumNArgsWithLessArgs checks if the provided error is due to a function requiring at least two arguments but receiving fewer.
//
// Parameters:
//   - err: The error returned by the function.
//   - t: A testing.T instance used for assertions.
//
// This function asserts that the error message matches the expected "requires at least 2 arg(s), only received 1".
// If the assertion fails, it calls t.Fatalf with a formatted error message.
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

// maximumNArgsWithMoreArgs checks if the error is related to accepting more arguments than allowed.
// It expects an error and a testing.T instance. If no error is received, it fails the test with a fatal error.
// It then compares the expected error message with the actual one; if they don't match, it fails the test with a fatal error.
// Parameters:
//   - err: The error to be checked.
//   - t: The testing.T instance used for assertions and logging.
// Returns:
//   - No return value.
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

// exactArgsWithInvalidCount checks if the given error indicates that a function expects exactly two arguments but received three.
//
// Parameters:
//   - err: The error to check.
//   - t: The testing.T instance used for assertions.
//
// Returns:
//   This function does not return any value.
//
// Errors:
//   - If err is nil, this function calls t.Fatal with an "Expected an error" message.
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

// rangeArgsWithInvalidCount checks if the provided error is related to invalid argument count.
// It expects an error indicating that a function accepts between 2 and 4 arguments,
// but received only one. If the error does not match this expectation or if there is no error, it will call t.Fatal.
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

// TestNoArgs tests the behavior of the command when no arguments are provided. It verifies that the command executes successfully without any errors.
func TestNoArgs(t *testing.T) {
	c := getCommand(NoArgs, false)
	output, err := executeCommand(c)
	expectSuccess(output, err, t)
}

// TestNoArgs_WithArgs tests the command execution when it expects no arguments but receives one.
// It creates a command with NoArgs flag and executes it with an argument "one".
// The function checks if the error returned is as expected for a no-arguments command receiving an argument.
func TestNoArgs_WithArgs(t *testing.T) {
	c := getCommand(NoArgs, false)
	_, err := executeCommand(c, "one")
	noArgsWithArgs(err, t, "one")
}

// TestNoArgs_WithValid_WithArgs tests the behavior of a command with no arguments when provided with valid and unexpected arguments.
// It checks if the command execution returns an error for an unexpected argument.
func TestNoArgs_WithValid_WithArgs(t *testing.T) {
	c := getCommand(NoArgs, true)
	_, err := executeCommand(c, "one")
	noArgsWithArgs(err, t, "one")
}

// TestNoArgs_WithValid_WithInvalidArgs tests the NoArgs command with valid and invalid arguments.
func TestNoArgs_WithValid_WithInvalidArgs(t *testing.T) {
	c := getCommand(NoArgs, true)
	_, err := executeCommand(c, "a")
	noArgsWithArgs(err, t, "a")
}

// TestNoArgs_WithValidOnly_WithInvalidArgs tests the execution of a command with both valid and invalid arguments when OnlyValidArgs and NoArgs matchers are used.
// It asserts that executing the command with an argument results in an error as expected.
// The test uses the getCommand function to create a command with specific matchers and then checks if the error returned by executeCommand is valid using validOnlyWithInvalidArgs.
func TestNoArgs_WithValidOnly_WithInvalidArgs(t *testing.T) {
	c := getCommand(MatchAll(OnlyValidArgs, NoArgs), true)
	_, err := executeCommand(c, "a")
	validOnlyWithInvalidArgs(err, t)
}

// OnlyValidArgs

// TestOnlyValidArgs tests the command with valid arguments.
func TestOnlyValidArgs(t *testing.T) {
	c := getCommand(OnlyValidArgs, true)
	output, err := executeCommand(c, "one", "two")
	expectSuccess(output, err, t)
}

// TestOnlyValidArgs_WithInvalidArgs tests the behavior of the command when invalid arguments are provided.
// It verifies that an error is returned when executing the command with one invalid argument.
func TestOnlyValidArgs_WithInvalidArgs(t *testing.T) {
	c := getCommand(OnlyValidArgs, true)
	_, err := executeCommand(c, "a")
	validOnlyWithInvalidArgs(err, t)
}

// ArbitraryArgs

// TestArbitraryArgs tests the execution of a command with arbitrary arguments.
// It creates a command with the ArbitraryArgs flag set and executes it with two arguments "a" and "b".
// It then checks if the command executed successfully by expecting no errors and a successful output.
func TestArbitraryArgs(t *testing.T) {
	c := getCommand(ArbitraryArgs, false)
	output, err := executeCommand(c, "a", "b")
	expectSuccess(output, err, t)
}

// TestArbitraryArgs_WithValid tests the execution of a command with arbitrary arguments when they are valid.
// It verifies that the command executes successfully and returns expected output without errors.
func TestArbitraryArgs_WithValid(t *testing.T) {
	c := getCommand(ArbitraryArgs, true)
	output, err := executeCommand(c, "one", "two")
	expectSuccess(output, err, t)
}

// TestArbitraryArgs_WithValid_WithInvalidArgs tests the execution of a command with arbitrary arguments, including both valid and invalid cases.
// It verifies that the command executes successfully with valid arguments and handles errors for invalid arguments correctly.
// The function takes a testing.T instance to run assertions and checks the output and error of the executed command.
func TestArbitraryArgs_WithValid_WithInvalidArgs(t *testing.T) {
	c := getCommand(ArbitraryArgs, true)
	output, err := executeCommand(c, "a")
	expectSuccess(output, err, t)
}

// TestArbitraryArgs_WithValidOnly_WithInvalidArgs tests the command execution when only valid arguments are provided along with arbitrary arguments.
// It expects an error when invalid arguments are present during execution.
func TestArbitraryArgs_WithValidOnly_WithInvalidArgs(t *testing.T) {
	c := getCommand(MatchAll(OnlyValidArgs, ArbitraryArgs), true)
	_, err := executeCommand(c, "a")
	validOnlyWithInvalidArgs(err, t)
}

// MinimumNArgs

// TestMinimumNArgs tests the MinimumNArgs command with a minimum of 2 arguments.
func TestMinimumNArgs(t *testing.T) {
	c := getCommand(MinimumNArgs(2), false)
	output, err := executeCommand(c, "a", "b", "c")
	expectSuccess(output, err, t)
}

// TestMinimumNArgs_WithValid tests the execution of a command with at least two arguments.
// It asserts that the command executes successfully when provided with valid arguments.
func TestMinimumNArgs_WithValid(t *testing.T) {
	c := getCommand(MinimumNArgs(2), true)
	output, err := executeCommand(c, "one", "three")
	expectSuccess(output, err, t)
}

// TestMinimumNArgs_WithValid checks that the MinimumNArgs command works correctly with valid arguments.
// It tests both a successful execution and an error case when invalid arguments are provided.
func TestMinimumNArgs_WithValid__WithInvalidArgs(t *testing.T) {
	c := getCommand(MinimumNArgs(2), true)
	output, err := executeCommand(c, "a", "b")
	expectSuccess(output, err, t)
}

// TestMinimumNArgs_WithValidOnly_WithInvalidArgs tests the behavior of the command when it expects a minimum number of arguments (2) and only valid args, but receives invalid ones.
// It validates that the command execution fails with an error indicating invalid arguments.
func TestMinimumNArgs_WithValidOnly_WithInvalidArgs(t *testing.T) {
	c := getCommand(MatchAll(OnlyValidArgs, MinimumNArgs(2)), true)
	_, err := executeCommand(c, "a", "b")
	validOnlyWithInvalidArgs(err, t)
}

// TestMinimumNArgs_WithLessArgs tests the MinimumNArgs function when provided with fewer arguments than expected.
func TestMinimumNArgs_WithLessArgs(t *testing.T) {
	c := getCommand(MinimumNArgs(2), false)
	_, err := executeCommand(c, "a")
	minimumNArgsWithLessArgs(err, t)
}

// TestMinimumNArgs_WithLessArgs_WithValid tests the MinimumNArgs function with fewer arguments than expected.
func TestMinimumNArgs_WithLessArgs_WithValid(t *testing.T) {
	c := getCommand(MinimumNArgs(2), true)
	_, err := executeCommand(c, "one")
	minimumNArgsWithLessArgs(err, t)
}

// TestMinimumNArgs_WithLessArgs_WithValid_WithInvalidArgs tests the MinimumNArgs validator with fewer arguments than expected.
// It verifies both valid and invalid cases to ensure proper error handling.
func TestMinimumNArgs_WithLessArgs_WithValid_WithInvalidArgs(t *testing.T) {
	c := getCommand(MinimumNArgs(2), true)
	_, err := executeCommand(c, "a")
	minimumNArgsWithLessArgs(err, t)
}

// TestMinimumNArgs_WithLessArgs_WithValidOnly_WithInvalidArgs tests the behavior of a command when executed with less than the minimum required arguments and only valid args are allowed.
// It verifies that the command returns an error indicating invalid argument count when called with insufficient arguments.
func TestMinimumNArgs_WithLessArgs_WithValidOnly_WithInvalidArgs(t *testing.T) {
	c := getCommand(MatchAll(OnlyValidArgs, MinimumNArgs(2)), true)
	_, err := executeCommand(c, "a")
	validOnlyWithInvalidArgs(err, t)
}

// MaximumNArgs

// TestMaximumNArgs tests the execution of a command with a maximum number of arguments.
// It verifies that the command is executed successfully with up to 3 arguments and fails when more than 3 arguments are provided.
func TestMaximumNArgs(t *testing.T) {
	c := getCommand(MaximumNArgs(3), false)
	output, err := executeCommand(c, "a", "b")
	expectSuccess(output, err, t)
}

// TestMaximumNArgs_WithValid tests the functionality of the MaximumNArgs function with valid arguments.
// It verifies that the command is executed successfully when provided with two arguments.
func TestMaximumNArgs_WithValid(t *testing.T) {
	c := getCommand(MaximumNArgs(2), true)
	output, err := executeCommand(c, "one", "three")
	expectSuccess(output, err, t)
}

// TestMaximumNArgs_WithValid_WithInvalidArgs tests the behavior of the MaximumNArgs function with both valid and invalid arguments.
// It ensures that the function correctly handles different sets of input arguments.
func TestMaximumNArgs_WithValid_WithInvalidArgs(t *testing.T) {
	c := getCommand(MaximumNArgs(2), true)
	output, err := executeCommand(c, "a", "b")
	expectSuccess(output, err, t)
}

// TestMaximumNArgs_WithValidOnly_WithInvalidArgs tests the behavior of a command with the MatchAll validator, OnlyValidArgs option, and MaximumNArgs set to 2 when provided with both valid and invalid arguments. It asserts that the command execution results in an error indicating the presence of invalid arguments.
// Parameters:
//   - t: A pointer to a testing.T instance for running tests.
// Returns:
//   None
// Errors:
//   - An error if the command does not return an error when it should, or if the error message does not indicate the presence of invalid arguments.
func TestMaximumNArgs_WithValidOnly_WithInvalidArgs(t *testing.T) {
	c := getCommand(MatchAll(OnlyValidArgs, MaximumNArgs(2)), true)
	_, err := executeCommand(c, "a", "b")
	validOnlyWithInvalidArgs(err, t)
}

// TestMaximumNArgs_WithMoreArgs tests the behavior of the MaximumNArgs constraint when provided with more arguments than allowed. It checks if the resulting command, when executed, returns an error indicating too many arguments.

// getCommand creates and returns a test command with the given argument constraint and allowStdin flag.
// executeCommand executes the given command with the specified arguments and returns the result and any error encountered.
// maximumNArgsWithMoreArgs asserts that the provided error indicates the use of more arguments than allowed for MaximumNArgs constraints.
func TestMaximumNArgs_WithMoreArgs(t *testing.T) {
	c := getCommand(MaximumNArgs(2), false)
	_, err := executeCommand(c, "a", "b", "c")
	maximumNArgsWithMoreArgs(err, t)
}

// TestMaximumNArgs_WithMoreArgs_WithValid tests the execution of a command with more arguments than allowed by MaximumNArgs.
// It verifies that an error is returned when executing the command with three arguments when only two are permitted.
func TestMaximumNArgs_WithMoreArgs_WithValid(t *testing.T) {
	c := getCommand(MaximumNArgs(2), true)
	_, err := executeCommand(c, "one", "three", "two")
	maximumNArgsWithMoreArgs(err, t)
}

// TestMaximumNArgs_WithMoreArgs_WithValid_WithInvalidArgs tests the behavior of the MaximumNArgs decorator when called with more arguments than expected.
// It checks both valid and invalid argument scenarios.
// It uses the getCommand function to create a command with the MaximumNArgs decorator applied.
// The executeCommand function is then used to run the command with "a", "b", and "c" as arguments.
// The maximumNArgsWithMoreArgs function asserts the expected error behavior based on the number of arguments.
func TestMaximumNArgs_WithMoreArgs_WithValid_WithInvalidArgs(t *testing.T) {
	c := getCommand(MaximumNArgs(2), true)
	_, err := executeCommand(c, "a", "b", "c")
	maximumNArgsWithMoreArgs(err, t)
}

// TestMaximumNArgs_WithMoreArgs_WithValidOnly_WithInvalidArgs tests the behavior of a command with a maximum number of arguments and only valid arguments.
// It ensures that an error is returned when more than two arguments are provided.
// The test checks if the command correctly identifies invalid arguments when they are present.
func TestMaximumNArgs_WithMoreArgs_WithValidOnly_WithInvalidArgs(t *testing.T) {
	c := getCommand(MatchAll(OnlyValidArgs, MaximumNArgs(2)), true)
	_, err := executeCommand(c, "a", "b", "c")
	validOnlyWithInvalidArgs(err, t)
}

// ExactArgs

// TestExactArgs tests the functionality of a command that requires exactly three arguments.
// It sets up a command with ExactArgs set to 3 and executes it with three arguments.
// The output and error are then checked for success using expectSuccess.
func TestExactArgs(t *testing.T) {
	c := getCommand(ExactArgs(3), false)
	output, err := executeCommand(c, "a", "b", "c")
	expectSuccess(output, err, t)
}

// TestExactArgs_WithValid tests the ExactArgs function with valid arguments.
// It verifies that the command is executed successfully when the correct number of arguments is provided.
func TestExactArgs_WithValid(t *testing.T) {
	c := getCommand(ExactArgs(3), true)
	output, err := executeCommand(c, "three", "one", "two")
	expectSuccess(output, err, t)
}

// TestExactArgs_WithValid_WithInvalidArgs tests the ExactArgs validator with valid and invalid arguments.
// It creates a command with ExactArgs set to 3, executes it with various argument sets,
// and checks if the execution behaves as expected for both valid and invalid cases.
func TestExactArgs_WithValid_WithInvalidArgs(t *testing.T) {
	c := getCommand(ExactArgs(3), true)
	output, err := executeCommand(c, "three", "a", "two")
	expectSuccess(output, err, t)
}

// TestExactArgs_WithValidOnly_WithInvalidArgs tests the behavior of a command that expects exactly 3 arguments and allows only valid args.
// It executes the command with invalid arguments and checks if the error returned is as expected when using OnlyValidArgs.
func TestExactArgs_WithValidOnly_WithInvalidArgs(t *testing.T) {
	c := getCommand(MatchAll(OnlyValidArgs, ExactArgs(3)), true)
	_, err := executeCommand(c, "three", "a", "two")
	validOnlyWithInvalidArgs(err, t)
}

// TestExactArgs_WithInvalidCount tests the behavior of a command that expects exactly two arguments but receives an invalid count.
// It creates a command with ExactArgs(2) validator and executes it with three arguments, expecting an error due to the mismatched argument count.
func TestExactArgs_WithInvalidCount(t *testing.T) {
	c := getCommand(ExactArgs(2), false)
	_, err := executeCommand(c, "a", "b", "c")
	exactArgsWithInvalidCount(err, t)
}

// TestExactArgs_WithInvalidCount_WithValid tests the execution of a command with an exact number of arguments when provided with an invalid count.
// It expects an error related to the argument count being incorrect.
// The test ensures that the command handling correctly identifies and reports errors for too many arguments.
func TestExactArgs_WithInvalidCount_WithValid(t *testing.T) {
	c := getCommand(ExactArgs(2), true)
	_, err := executeCommand(c, "three", "one", "two")
	exactArgsWithInvalidCount(err, t)
}

// TestExactArgs_WithInvalidCount_WithValid_WithInvalidArgs tests the behavior of a command with exact arguments when executed with an invalid count and valid arguments.
// It checks if the command returns an error when provided with more than the expected number of arguments.
func TestExactArgs_WithInvalidCount_WithValid_WithInvalidArgs(t *testing.T) {
	c := getCommand(ExactArgs(2), true)
	_, err := executeCommand(c, "three", "a", "two")
	exactArgsWithInvalidCount(err, t)
}

// TestExactArgs_WithInvalidCount_WithValidOnly_WithInvalidArgs tests the behavior of a command with exact argument count and valid-only filter when provided with invalid arguments.
// It expects an error due to the mismatch in the number of arguments passed.
// The test ensures that the OnlyValidArgs filter is applied correctly, allowing only valid arguments through.
func TestExactArgs_WithInvalidCount_WithValidOnly_WithInvalidArgs(t *testing.T) {
	c := getCommand(MatchAll(OnlyValidArgs, ExactArgs(2)), true)
	_, err := executeCommand(c, "three", "a", "two")
	validOnlyWithInvalidArgs(err, t)
}

// RangeArgs

// TestRangeArgs tests the functionality of RangeArgs with provided arguments.
// It creates a command using RangeArgs and executes it with given input arguments.
// Finally, it checks if the command execution was successful.
func TestRangeArgs(t *testing.T) {
	c := getCommand(RangeArgs(2, 4), false)
	output, err := executeCommand(c, "a", "b", "c")
	expectSuccess(output, err, t)
}

// TestRangeArgs_WithValid tests the RangeArgs function with valid arguments and ensures that the command execution is successful. It uses a test function to verify the output and error returned by the executeCommand function. The getCommand function is used to create a command based on the provided arguments and the debug flag.
func TestRangeArgs_WithValid(t *testing.T) {
	c := getCommand(RangeArgs(2, 4), true)
	output, err := executeCommand(c, "three", "one", "two")
	expectSuccess(output, err, t)
}

// TestRangeArgs_WithValid_WithInvalidArgs tests the RangeArgs function with valid and invalid arguments.
// It verifies that the command execution behaves as expected when provided with a range of integers (2 to 4).
// It checks both valid and invalid arguments and ensures successful execution for valid inputs.
func TestRangeArgs_WithValid_WithInvalidArgs(t *testing.T) {
	c := getCommand(RangeArgs(2, 4), true)
	output, err := executeCommand(c, "three", "a", "two")
	expectSuccess(output, err, t)
}

// TestRangeArgs_WithValidOnly_WithInvalidArgs tests the behavior of the command when it expects a range of arguments (2 to 4) and only valid arguments.
// It verifies that an error is returned when the provided arguments do not meet the specified criteria.
func TestRangeArgs_WithValidOnly_WithInvalidArgs(t *testing.T) {
	c := getCommand(MatchAll(OnlyValidArgs, RangeArgs(2, 4)), true)
	_, err := executeCommand(c, "three", "a", "two")
	validOnlyWithInvalidArgs(err, t)
}

// TestRangeArgs_WithInvalidCount tests the behavior of the RangeArgs command when provided with an invalid number of arguments.
// It creates a command with an expected range and executes it with an incorrect argument count, expecting an error related to the number of arguments being out of range.
// The test uses a helper function `rangeArgsWithInvalidCount` to verify that the correct error is returned.
func TestRangeArgs_WithInvalidCount(t *testing.T) {
	c := getCommand(RangeArgs(2, 4), false)
	_, err := executeCommand(c, "a")
	rangeArgsWithInvalidCount(err, t)
}

// TestRangeArgs_WithInvalidCount_WithValid tests the RangeArgs function with an invalid count but valid arguments.
// It creates a command using the getCommand function and executes it with a single argument "two".
// The function then checks if the error returned by executeCommand matches the expected error for invalid count.
func TestRangeArgs_WithInvalidCount_WithValid(t *testing.T) {
	c := getCommand(RangeArgs(2, 4), true)
	_, err := executeCommand(c, "two")
	rangeArgsWithInvalidCount(err, t)
}

// TestRangeArgs_WithInvalidCount_WithValid_WithInvalidArgs tests the execution of a command with invalid count and valid arguments.
// It checks if the function handles errors correctly when provided with an invalid number of arguments.
// Parameters:
//   - t: A pointer to a testing.T object used for test assertions.
// The function does not return any values but asserts on the behavior of the executeCommand function.
func TestRangeArgs_WithInvalidCount_WithValid_WithInvalidArgs(t *testing.T) {
	c := getCommand(RangeArgs(2, 4), true)
	_, err := executeCommand(c, "a")
	rangeArgsWithInvalidCount(err, t)
}

// TestRangeArgs_WithInvalidCount_WithValidOnly_WithInvalidArgs tests the behavior of a command with range arguments when given an invalid count and valid only option, expecting errors for invalid arguments.
// It uses a test function to verify that executing the command with "a" as an argument results in an error due to the mismatch between expected and actual arguments.
func TestRangeArgs_WithInvalidCount_WithValidOnly_WithInvalidArgs(t *testing.T) {
	c := getCommand(MatchAll(OnlyValidArgs, RangeArgs(2, 4)), true)
	_, err := executeCommand(c, "a")
	validOnlyWithInvalidArgs(err, t)
}

// Takes(No)Args

// TestRootTakesNoArgs tests that calling the root command with illegal arguments returns an error.
// It creates a root command and a child command, adds the child to the root, and then attempts to execute
// the root command with invalid arguments. It expects an error indicating that the "illegal" command is unknown for "root".
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

// TestRootTakesArgs tests if the root command correctly handles arguments.
func TestRootTakesArgs(t *testing.T) {
	rootCmd := &Command{Use: "root", Args: ArbitraryArgs, Run: emptyRun}
	childCmd := &Command{Use: "child", Run: emptyRun}
	rootCmd.AddCommand(childCmd)

	_, err := executeCommand(rootCmd, "legal", "args")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

// TestChildTakesNoArgs tests that the command handles unexpected arguments correctly when a subcommand does not accept any arguments.
//
// It verifies that calling "child" with additional arguments results in an error.
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

// TestChildTakesArgs tests that a child command with arguments can be executed successfully.
// It creates a root command and adds a child command with arbitrary argument requirements.
// Then, it executes the child command with legal arguments and verifies there is no error.
func TestChildTakesArgs(t *testing.T) {
	rootCmd := &Command{Use: "root", Run: emptyRun}
	childCmd := &Command{Use: "child", Args: ArbitraryArgs, Run: emptyRun}
	rootCmd.AddCommand(childCmd)

	_, err := executeCommand(rootCmd, "child", "legal", "args")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

// TestMatchAll tests the MatchAll function with various test cases to ensure it correctly validates command arguments.
// It checks that the number of arguments is exactly 3 and each argument is exactly 2 bytes long.
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

// TestExactValidArgs tests the behavior of a command when provided with exactly three valid arguments.
// It checks if the command executes successfully with the given arguments and validates the output.
func TestExactValidArgs(t *testing.T) {
	c := getCommand(ExactValidArgs(3), true)
	output, err := executeCommand(c, "three", "one", "two")
	expectSuccess(output, err, t)
}

// TestExactValidArgs_WithInvalidCount tests the ExactValidArgs function with an invalid argument count.
// It asserts that an error is returned when the number of arguments does not match the expected count.
func TestExactValidArgs_WithInvalidCount(t *testing.T) {
	c := getCommand(ExactValidArgs(2), false)
	_, err := executeCommand(c, "three", "one", "two")
	exactArgsWithInvalidCount(err, t)
}

// TestExactValidArgs_WithInvalidCount_WithInvalidArgs tests the behavior of the ExactValidArgs validator when provided with an invalid count and invalid arguments.
// It creates a command with the ExactValidArgs validator, executes it with three arguments (two valid and one invalid), and checks if the error returned is as expected for invalid argument counts.
func TestExactValidArgs_WithInvalidCount_WithInvalidArgs(t *testing.T) {
	c := getCommand(ExactValidArgs(2), true)
	_, err := executeCommand(c, "three", "a", "two")
	exactArgsWithInvalidCount(err, t)
}

// TestExactValidArgs_WithInvalidArgs tests the behavior of ExactValidArgs when invoked with invalid arguments.
// It uses a test command and checks if the execution results in an error as expected.
func TestExactValidArgs_WithInvalidArgs(t *testing.T) {
	c := getCommand(ExactValidArgs(2), true)
	_, err := executeCommand(c, "three", "a")
	validOnlyWithInvalidArgs(err, t)
}

// TestLegacyArgsRootAcceptsArgs ensures that the root command accepts arguments if it does not have sub-commands.
// It verifies backwards-compatibility with respect to the legacyArgs() function.
func TestLegacyArgsRootAcceptsArgs(t *testing.T) {
	rootCmd := &Command{Use: "root", Args: nil, Run: emptyRun}

	_, err := executeCommand(rootCmd, "somearg")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

// TestLegacyArgsSubcmdAcceptsArgs verifies that a sub-command accepts arguments and further sub-commands while maintaining backwards-compatibility with the legacyArgs() function. It ensures that executing commands like "root child somearg" does not result in errors.
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
