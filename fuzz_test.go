//go:build go1.18
// +build go1.18

// Copyright 2013-2023 The Cobra Authors
// SPDX-License-Identifier: Apache-2.0

package cobra_test

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
)

// FuzzCobraExecute tests command execution with arbitrary CLI args.
// Cobra is the #1 Go CLI framework (44K+ stars). Every Go CLI tool
// built with Cobra processes untrusted user input through this path.
func FuzzCobraExecute(f *testing.F) {
	f.Add("--help")
	f.Add("subcommand --flag value")
	f.Add("")
	f.Add(string(make([]byte, 10000)))

	f.Fuzz(func(t *testing.T, argLine string) {
		if len(argLine) > 10000 {
			return
		}

		cmd := &cobra.Command{
			Use: "test",
			Run: func(cmd *cobra.Command, args []string) {},
		}
		cmd.Flags().String("flag", "", "test flag")
		cmd.AddCommand(&cobra.Command{
			Use: "sub",
			Run: func(cmd *cobra.Command, args []string) {},
		})

		// Split args and execute — must never panic
		args := splitArgs(argLine)
		cmd.SetArgs(args)
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})
		_ = cmd.Execute()
	})
}

func splitArgs(s string) []string {
	if s == "" {
		return nil
	}
	var args []string
	inQuote := false
	current := ""
	for _, c := range s {
		if c == '"' {
			inQuote = !inQuote
		} else if c == ' ' && !inQuote {
			if current != "" {
				args = append(args, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		args = append(args, current)
	}
	return args
}
