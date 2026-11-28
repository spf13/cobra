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
	"io"
	"strconv"
	"testing"

	"github.com/spf13/pflag"
)

func benchBashTree(n int) *Command {
	root := &Command{Use: "app", Run: func(cmd *Command, args []string) {}}
	for i := 0; i < n; i++ {
		name := "cmd" + strconv.Itoa(i)
		child := &Command{Use: name, Run: func(cmd *Command, args []string) {}}
		child.Flags().String(name+"-flag", "", "flag")
		if err := child.RegisterFlagCompletionFunc(name+"-flag", func(cmd *Command, args []string, toComplete string) ([]Completion, ShellCompDirective) {
			return nil, ShellCompDirectiveDefault
		}); err != nil {
			panic(err)
		}
		root.AddCommand(child)
	}
	return root
}

func BenchmarkGenBashCompletionRegisteredFlags(b *testing.B) {
	for _, n := range []int{100, 400} {
		b.Run(strconv.Itoa(n), func(b *testing.B) {
			flagCompletionMutex.Lock()
			saved := flagCompletionFunctions
			flagCompletionFunctions = map[*pflag.Flag]CompletionFunc{}
			flagCompletionMutex.Unlock()
			b.Cleanup(func() {
				flagCompletionMutex.Lock()
				flagCompletionFunctions = saved
				flagCompletionMutex.Unlock()
			})
			root := benchBashTree(n)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if err := root.GenBashCompletion(io.Discard); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
