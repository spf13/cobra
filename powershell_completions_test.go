package cobra

import (
	"bytes"
	"strings"
	"testing"
)

func TestPowerShellCompletion(t *testing.T) {
	tcs := []struct {
		name                string
		root                *Command
		expectedExpressions []string
	}{
		{
			name: "trivial",
			root: &Command{Use: "trivialapp"},
			expectedExpressions: []string{
				"Register-ArgumentCompleter -Native -CommandName 'trivialapp' -ScriptBlock",
				"$command = @(\n        'trivialapp'\n",
			},
		},
		{
			name: "tree",
			root: func() *Command {
				r := &Command{Use: "tree"}

				sub1 := &Command{Use: "sub1"}
				r.AddCommand(sub1)

				sub11 := &Command{Use: "sub11"}
				sub12 := &Command{Use: "sub12"}

				sub1.AddCommand(sub11)
				sub1.AddCommand(sub12)

				sub2 := &Command{Use: "sub2"}
				r.AddCommand(sub2)

				sub21 := &Command{Use: "sub21"}
				sub22 := &Command{Use: "sub22"}

				sub2.AddCommand(sub21)
				sub2.AddCommand(sub22)

				return r
			}(),
			expectedExpressions: []string{
				"'tree'",
				"[CompletionResult]::new('sub1', 'sub1', [CompletionResultType]::ParameterValue, '')",
				"[CompletionResult]::new('sub2', 'sub2', [CompletionResultType]::ParameterValue, '')",
				"'tree;sub1'",
				"[CompletionResult]::new('sub11', 'sub11', [CompletionResultType]::ParameterValue, '')",
				"[CompletionResult]::new('sub12', 'sub12', [CompletionResultType]::ParameterValue, '')",
				"'tree;sub1;sub11'",
				"'tree;sub1;sub12'",
				"'tree;sub2'",
				"[CompletionResult]::new('sub21', 'sub21', [CompletionResultType]::ParameterValue, '')",
				"[CompletionResult]::new('sub22', 'sub22', [CompletionResultType]::ParameterValue, '')",
				"'tree;sub2;sub21'",
				"'tree;sub2;sub22'",
			},
		},
		{
			name: "flags",
			root: func() *Command {
				r := &Command{Use: "flags"}
				r.Flags().StringP("flag1", "a", "", "")
				r.Flags().String("flag2", "", "")

				sub1 := &Command{Use: "sub1"}
				sub1.Flags().StringP("flag3", "c", "", "")
				r.AddCommand(sub1)

				return r
			}(),
			expectedExpressions: []string{
				"'flags'",
				"[CompletionResult]::new('-a', 'a', [CompletionResultType]::ParameterName, '')",
				"[CompletionResult]::new('--flag1', 'flag1', [CompletionResultType]::ParameterName, '')",
				"[CompletionResult]::new('--flag2', 'flag2', [CompletionResultType]::ParameterName, '')",
				"[CompletionResult]::new('sub1', 'sub1', [CompletionResultType]::ParameterValue, '')",
				"'flags;sub1'",
				"[CompletionResult]::new('-c', 'c', [CompletionResultType]::ParameterName, '')",
				"[CompletionResult]::new('--flag3', 'flag3', [CompletionResultType]::ParameterName, '')",
			},
		},
		{
			name: "usage",
			root: func() *Command {
				r := &Command{Use: "usage"}
				r.Flags().String("flag", "", "this describes the usage of the 'flag' flag")

				sub1 := &Command{
					Use:   "sub1",
					Short: "short describes 'sub1'",
				}
				r.AddCommand(sub1)

				return r
			}(),
			expectedExpressions: []string{
				"[CompletionResult]::new('--flag', 'flag', [CompletionResultType]::ParameterName, 'this describes the usage of the ''flag'' flag')",
				"[CompletionResult]::new('sub1', 'sub1', [CompletionResultType]::ParameterValue, 'short describes ''sub1''')",
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			tc.root.GenPowerShellCompletion(buf)
			output := buf.String()

			for _, expectedExpression := range tc.expectedExpressions {
				if !strings.Contains(output, expectedExpression) {
					t.Errorf("Expected completion to contain %q somewhere; got %q", expectedExpression, output)
				}
			}
		})
	}
}
