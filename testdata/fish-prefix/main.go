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

// A minimal completion backend for reproducing Fish flag-prefix handling.
// Build from the repository root: go build -o /tmp/fish-prefix ./testdata/fish-prefix
package main

import (
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func candidates(_ *cobra.Command, _ []string, prefix string) ([]string, cobra.ShellCompDirective) {
	var result []string
	for _, value := range []string{"web", "worker", "app=web", "app=worker", "app-name=web", "app-name=worker", "app=x=web", "app=x=worker"} {
		if strings.HasPrefix(value, prefix) {
			result = append(result, value)
		}
	}
	return result, cobra.ShellCompDirectiveNoFileComp
}

func main() {
	root := &cobra.Command{Use: "fish-prefix", ValidArgsFunction: candidates, Run: func(*cobra.Command, []string) {}}
	root.Flags().StringP("selector", "s", "", "Select a value")
	if err := root.RegisterFlagCompletionFunc("selector", candidates); err != nil {
		panic(err)
	}
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
