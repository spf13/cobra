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

package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{Use: "cobrafixture"}
	root.PersistentFlags().String("config", "", "config file")
	run := &cobra.Command{Use: "run", Run: func(*cobra.Command, []string) {}, ValidArgsFunction: func(cmd *cobra.Command, args []string, prefix string) ([]string, cobra.ShellCompDirective) {
		switch prefix {
		case "error":
			return []string{"ignored"}, cobra.ShellCompDirectiveError
		case "nospace":
			return []string{"nospace:"}, cobra.ShellCompDirectiveNoSpace
		case "order":
			return []string{"order-z", "order-a"}, cobra.ShellCompDirectiveKeepOrder
		case "files":
			return []string{"ino"}, cobra.ShellCompDirectiveFilterFileExt
		case "dirs":
			return nil, cobra.ShellCompDirectiveFilterDirs
		case "none":
			return nil, cobra.ShellCompDirectiveNoFileComp
		case "fallback":
			return nil, cobra.ShellCompDirectiveDefault
		case "help":
			return cobra.AppendActiveHelp(nil, "Choose a target"), cobra.ShellCompDirectiveNoFileComp
		}
		config, _ := cmd.Flags().GetString("config")
		return []string{"alpha\tFirst target", "beta\tSecond target", fmt.Sprintf("config:%s", config)}, cobra.ShellCompDirectiveNoFileComp
	}}
	root.AddCommand(run)
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
