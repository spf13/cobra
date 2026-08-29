package cobra

import (
	"bytes"
	"strings"
	"testing"
)

// TestIssue2154_HelpFuncArgs tests that the help function receives correct args
// See: https://github.com/spf13/cobra/issues/2154
func TestIssue2154_HelpFuncArgs(t *testing.T) {
	var receivedCmd *Command
	var receivedArgs []string

	rootCmd := &Command{Use: "root"}
	subCmd := &Command{Use: "sub"}
	rootCmd.AddCommand(subCmd)

	rootCmd.SetHelpFunc(func(c *Command, args []string) {
		receivedCmd = c
		receivedArgs = args
		// Output something so we know help was called
		c.Print("Help called\n")
	})

	// Test case 1: using -h flag with args
	// Command: prog sub arg1 arg2 -h
	// Expected: cmd=sub, args=[arg1, arg2]
	t.Run("help flag with args", func(t *testing.T) {
		receivedCmd = nil
		receivedArgs = nil

		output := new(bytes.Buffer)
		rootCmd.SetOut(output)
		rootCmd.SetArgs([]string{"sub", "arg1", "arg2", "-h"})
		rootCmd.Execute()

		if receivedCmd == nil {
			t.Fatal("help function was not called")
		}

		if receivedCmd.Name() != "sub" {
			t.Errorf("expected cmd.Name()='sub', got '%s'", receivedCmd.Name())
		}

		expected := []string{"arg1", "arg2"}
		if !stringSliceEqual(receivedArgs, expected) {
			t.Errorf("expected args=%v, got %v", expected, receivedArgs)
		}
	})

	// Test case 2: using help command with args
	// Command: prog help sub arg1 arg2
	// Expected: cmd=sub, args=[arg1, arg2]
	t.Run("help command with args", func(t *testing.T) {
		receivedCmd = nil
		receivedArgs = nil

		output := new(bytes.Buffer)
		rootCmd.SetOut(output)
		rootCmd.SetArgs([]string{"help", "sub", "arg1", "arg2"})
		rootCmd.Execute()

		if receivedCmd == nil {
			t.Fatal("help function was not called")
		}

		if receivedCmd.Name() != "sub" {
			t.Errorf("expected cmd.Name()='sub', got '%s'", receivedCmd.Name())
		}

		expected := []string{"arg1", "arg2"}
		if !stringSliceEqual(receivedArgs, expected) {
			t.Errorf("expected args=%v, got %v", expected, receivedArgs)
		}
	})

	// Test case 3: help with no args
	// Command: prog sub -h
	// Expected: cmd=sub, args=[]
	t.Run("help flag without args", func(t *testing.T) {
		receivedCmd = nil
		receivedArgs = nil

		output := new(bytes.Buffer)
		rootCmd.SetOut(output)
		rootCmd.SetArgs([]string{"sub", "-h"})
		rootCmd.Execute()

		if receivedCmd == nil {
			t.Fatal("help function was not called")
		}

		if receivedCmd.Name() != "sub" {
			t.Errorf("expected cmd.Name()='sub', got '%s'", receivedCmd.Name())
		}

		expected := []string{}
		if !stringSliceEqual(receivedArgs, expected) {
			t.Errorf("expected args=%v, got %v", expected, receivedArgs)
		}
	})

	// Test case 4: help command with no args
	// Command: prog help sub
	// Expected: cmd=sub, args=[]
	t.Run("help command without args", func(t *testing.T) {
		receivedCmd = nil
		receivedArgs = nil

		output := new(bytes.Buffer)
		rootCmd.SetOut(output)
		rootCmd.SetArgs([]string{"help", "sub"})
		rootCmd.Execute()

		if receivedCmd == nil {
			t.Fatal("help function was not called")
		}

		if receivedCmd.Name() != "sub" {
			t.Errorf("expected cmd.Name()='sub', got '%s'", receivedCmd.Name())
		}

		expected := []string{}
		if !stringSliceEqual(receivedArgs, expected) {
			t.Errorf("expected args=%v, got %v", expected, receivedArgs)
		}
	})
}

func stringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
