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
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/spf13/pflag"
)

// provides the user input from stdio in default - separated for better testability
var readerFactory = func() *bufio.Reader {
	return bufio.NewReader(os.Stdin)
}

// provides the user selection for the command - separated for better testability
var selectionFactory = func(options []string) (string, error) {
	idx, err := fuzzyfinder.Find(
		options,
		func(i int) string { return options[i] },
	)
	if err != nil {
		return "", err
	}
	return options[idx], nil
}

// This function enables a fuzzy style interactive execution, without
// passing all required sub commands and flags at start time.
// cmd - cobra root command
func RunInteractive(cmd *Command) (*Command, error) {
	subCommands := cmd.Commands()
	if len(subCommands) == 0 {
		return nil, fmt.Errorf("command has no sub commads")
	}
	currentCmd := cmd
	for {
		options := getOptionsFromCommands(subCommands...)
		selected, err := selectionFactory(options)
		if err != nil {
			return nil, fmt.Errorf("error in interactive run: %v", err)
		}
		if strings.HasPrefix(selected, "help ") {
			helpCmd, _, err := currentCmd.Find([]string{"help"})
			return helpCmd, err
		}
		nextCmd, _, err := currentCmd.Find([]string{selected})
		if err != nil {
			return nil, fmt.Errorf("error finding seleted sub command: %v", err)
		}
		subCommands = nextCmd.Commands()
		if len(subCommands) == 0 {
			// reached end of the chain ..
			cmdChain, txt := getCommandChain(nextCmd)
			setFlagsForCommands(txt, cmdChain...)
			return nextCmd, nil
		}
		currentCmd = nextCmd
	}
}

// Provides the chain of all included sub commands for a given command
func getCommandChain(cmd *Command) ([]*Command, string) {
	txt := ""
	chain := make([]*Command, 0)
	for c := cmd; c != nil; c = c.Parent() {
		chain = append(chain, c)
		if txt != "" {
			txt = c.Use + " " + txt
		} else {
			txt = c.Use
		}
	}
	return chain, txt
}

// Returns true if the given flag is marked as required
func isFlagRequired(f *pflag.Flag) bool {
	if f.Annotations != nil {
		if v, ok := f.Annotations[BashCompOneRequiredFlag]; ok && len(v) > 0 && v[0] == "true" {
			return true
		}
	}
	return false
}

func isRepeatableFlag(f *pflag.Flag) bool {
	switch f.Value.Type() {
	case "stringSlice", "intSlice", "float32Slice", "boolSlice", "stringArray":
		return true
	default:
		return false
	}
}

// iterates over the selected commands and collects input for their configured flags
func setFlagsForCommands(cmdChain string, cmds ...*Command) {
	showedChain := false
	reader := readerFactory()
	for _, cmd := range cmds {
		// Iterate over all flags of the command
		cmd.Flags().VisitAll(func(f *pflag.Flag) {
			// Ask interactively for input
			if f.Name == "help" {
				return
			}

			if !showedChain {
				fmt.Printf("\n`%s` will be called.\nIn the following steps the possible flags will be collected ...", cmdChain)
				showedChain = true
			}

			defValue := "empty"
			if f.DefValue != "" {
				defValue = f.DefValue
			}
			skipTxt := "imput required!"
			flagRequired := isFlagRequired(f)
			if !flagRequired {
				skipTxt = "press ⏎ to skip"
			}
			if isRepeatableFlag(f) {
				collectRepeatedFlagInput(cmd, f, defValue, skipTxt, reader)
			} else {
				collectFlagInput(cmd, f, flagRequired, defValue, skipTxt, reader)
			}
		})
	}
}

// implements the user interaction to get the required input for a flag
func collectFlagInput(cmd *Command, f *pflag.Flag, flagRequired bool, defValue, skipTxt string, reader *bufio.Reader) {
	for {
		fmt.Printf("\n%s (default: %s), %s: ", f.Usage, defValue, skipTxt)
		input, _ := reader.ReadString('\n') // read entire line
		input = strings.TrimSpace(input)    // remove newline and spaces

		if input != "" {
			// User provided a value -> set it
			if err := cmd.Flags().Set(f.Name, input); err != nil {
				fmt.Printf("⚠️  Could not set flag %s: %v\n", f.Name, err)
			} else {
				break
			}
		} else if flagRequired {
			fmt.Printf("⚠️  Flag %s is required, so input is needed!\n", f.Name)
		} else {
			break
		}
	}
}

func collectRepeatedFlagInput(cmd *Command, f *pflag.Flag, defValue, skipTxt string, reader *bufio.Reader) {
	bFirst := true
	for {
		if bFirst {
			fmt.Printf("\n%s (default: %s), %s\nmultiple values possible, press ⏎ to finish: ", f.Usage, defValue, skipTxt)
			bFirst = false
		} else {
			fmt.Printf("\nnext value, press ⏎ to finish: ")
		}
		input, _ := reader.ReadString('\n') // read entire line
		input = strings.TrimSpace(input)    // remove newline and spaces

		if input != "" {
			// User provided a value -> set it
			if err := cmd.Flags().Set(f.Name, input); err != nil {
				fmt.Printf("⚠️  Could not set flag %s: %v\n", f.Name, err)
			}
		} else {
			break
		}
	}
}

// produces the fuzzy search input for interactive command selection
func getOptionsFromCommands(cmds ...*Command) []string {
	ret := make([]string, 0)
	for _, cmd := range cmds {
		if cmd.Use == "completion" {
			continue
		}
		ret = append(ret, cmd.Use)
	}
	return ret
}
