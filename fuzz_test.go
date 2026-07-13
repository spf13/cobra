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

// Go native fuzz targets for the shell-completion machinery.
//
// These targets address upstream issue #2261 ("Investigate fuzz testing"),
// the only kind/security-labeled issue, which asks for Go native fuzzing.
// Prior to this file the repository had zero `func Fuzz*` targets.
//
// The two natural attack surfaces exercised here are:
//
//  1. The static shell-completion *generators* (GenBashCompletion,
//     GenBashCompletionV2, GenZshCompletion(NoDesc), GenFishCompletion,
//     GenPowerShellCompletion(WithDesc)). Command metadata (Use, Short,
//     flag names/usages, ValidArgs) is embedded into the generated shell
//     script, so hostile metadata could in principle break the script or
//     panic the generator.
//
//  2. The runtime `__complete` emit path in completions.go (roughly lines
//     242-294), where the value+TAB+description stream returned by a
//     ValidArgsFunction / RegisterFlagCompletionFunc is trimmed to its first
//     line and written to stdout for the shell script to consume. This is
//     the path prior recon flagged as thinly escaped (notably for zsh).
//
// The fuzz targets only assert "does not panic". They deliberately do not
// assert on output shape, so a discovered crash is a genuine robustness bug
// rather than a brittle-golden-file mismatch.

import (
	"io"
	"testing"
)

// seeds is the shared seed corpus used by every fuzz target. It concentrates
// on shell metacharacters and control bytes that are the most likely to break
// a generated completion script or the runtime emit path.
var fuzzSeeds = []string{
	"",
	"normal",
	"with space",
	"back`tick`",
	"$(command-substitution)",
	"${brace-expansion}",
	"semi;colon",
	"pipe|char",
	"amper&sand",
	"redirect>out",
	"glob*star",
	"question?mark",
	"at@sign",
	"colon:separated",
	"double\"quote",
	"single'quote",
	"back\\slash",
	"-leadingdash",
	"--doubledash",
	"tab\tembedded",
	"new\nline",
	"carriage\rreturn",
	"ansi\x1b[31mred\x1b[0m",
	"nul\x00byte",
	"percent%s%d%v",
	"bracket[expr]",
	"paren(expr)",
	"brace{expr}",
	"hash#comment",
	"bang!bang",
	"unicode-café-λ-🔥",
	"quote:and`tick`:$(sub)",
}

// addSeeds registers every entry of fuzzSeeds with the fuzzer.
func addSeeds(f *testing.F, add func(string)) {
	for _, s := range fuzzSeeds {
		add(s)
	}
}

// buildFuzzCmd constructs a small but representative command tree whose
// user-controllable metadata is entirely derived from the fuzz input. This is
// the shared setup used by the generator fuzz targets.
func buildFuzzCmd(use, short, flagName, flagShort, validArg string) *Command {
	root := &Command{
		Use:   sanitizeUse(use),
		Short: short,
		Run:   emptyRun,
	}

	// A flag whose name/shorthand/usage come from the fuzz input.
	root.Flags().String(sanitizeFlagName(flagName), "", short)
	if fs := sanitizeShorthand(flagShort); fs != "" {
		root.Flags().StringP("fuzzshort"+fs, fs, "", short)
	}

	// ValidArgs seeded from the fuzz input.
	root.ValidArgs = []string{validArg}

	// A child command so subcommand generation is exercised too.
	child := &Command{
		Use:   sanitizeUse(short + "child"),
		Short: use,
		Run:   emptyRun,
	}
	root.AddCommand(child)

	return root
}

// sanitizeUse keeps a non-empty first word for the command's Use line while
// still passing the raw remainder through. cobra derives the command Name()
// from the first whitespace-delimited token of Use; an empty token yields a
// nameless command which is not a meaningful fuzz case.
func sanitizeUse(use string) string {
	if len(use) == 0 {
		return "fuzzcmd"
	}
	return "fuzzcmd " + use
}

// sanitizeFlagName produces a flag name that pflag will accept as a
// registration target (pflag panics on empty names or names containing "=" or
// whitespace at *registration* time, which is a programming error rather than
// the injection surface we are fuzzing). The fuzzed bytes are still carried
// through in the flag's usage string via the caller.
func sanitizeFlagName(name string) string {
	out := make([]rune, 0, len(name)+len("fuzzflag"))
	out = append(out, []rune("fuzzflag")...)
	for _, r := range name {
		switch r {
		case '=', ' ', '\t', '\n', '\r', 0:
			// skip characters pflag rejects at registration
		default:
			out = append(out, r)
		}
	}
	return string(out)
}

// sanitizeShorthand returns at most a single-byte shorthand acceptable to
// pflag, or "" to skip the shorthand flag. pflag requires a shorthand to be
// exactly one ASCII character, so multi-byte runes and the reserved
// registration characters are skipped (these are pflag registration-time
// programming errors, not the injection surface we are fuzzing).
func sanitizeShorthand(s string) string {
	for _, r := range s {
		if r > 127 {
			continue
		}
		switch r {
		case '=', ' ', '\t', '\n', '\r', 0, '-':
			continue
		default:
			return string(r)
		}
	}
	return ""
}

// FuzzGenBashCompletion fuzzes the classic (v1) bash completion generator.
func FuzzGenBashCompletion(f *testing.F) {
	addSeeds(f, func(s string) { f.Add(s) })
	f.Fuzz(func(t *testing.T, s string) {
		root := buildFuzzCmd(s, s, s, s, s)
		if err := root.GenBashCompletion(io.Discard); err != nil {
			// An error return is acceptable; a panic is not.
			_ = err
		}
	})
}

// FuzzGenBashCompletionV2 fuzzes the newer bash V2 generator (the one the
// default `completion bash` command now emits).
func FuzzGenBashCompletionV2(f *testing.F) {
	addSeeds(f, func(s string) { f.Add(s) })
	f.Fuzz(func(t *testing.T, s string) {
		root := buildFuzzCmd(s, s, s, s, s)
		if err := root.GenBashCompletionV2(io.Discard, true); err != nil {
			_ = err
		}
	})
}

// FuzzGenZshCompletion fuzzes the zsh generator, both with and without
// descriptions. The zsh emit/escape path was specifically flagged by prior
// recon as the thinnest.
func FuzzGenZshCompletion(f *testing.F) {
	addSeeds(f, func(s string) { f.Add(s) })
	f.Fuzz(func(t *testing.T, s string) {
		root := buildFuzzCmd(s, s, s, s, s)
		if err := root.GenZshCompletion(io.Discard); err != nil {
			_ = err
		}
		root2 := buildFuzzCmd(s, s, s, s, s)
		if err := root2.GenZshCompletionNoDesc(io.Discard); err != nil {
			_ = err
		}
	})
}

// FuzzGenFishCompletion fuzzes the fish generator.
func FuzzGenFishCompletion(f *testing.F) {
	addSeeds(f, func(s string) { f.Add(s) })
	f.Fuzz(func(t *testing.T, s string) {
		root := buildFuzzCmd(s, s, s, s, s)
		if err := root.GenFishCompletion(io.Discard, true); err != nil {
			_ = err
		}
	})
}

// FuzzGenPowerShellCompletion fuzzes the powershell generator, both with and
// without descriptions.
func FuzzGenPowerShellCompletion(f *testing.F) {
	addSeeds(f, func(s string) { f.Add(s) })
	f.Fuzz(func(t *testing.T, s string) {
		root := buildFuzzCmd(s, s, s, s, s)
		if err := root.GenPowerShellCompletion(io.Discard); err != nil {
			_ = err
		}
		root2 := buildFuzzCmd(s, s, s, s, s)
		if err := root2.GenPowerShellCompletionWithDesc(io.Discard); err != nil {
			_ = err
		}
	})
}

// FuzzCompletionValues exercises the *runtime* `__complete` emit path.
//
// It registers a ValidArgsFunction that returns the fuzzed string as a
// completion value carrying a fuzzed description, then drives the hidden
// __complete command through the normal Execute() flow (exactly as the shell
// scripts do at runtime). This walks completions.go's per-completion loop
// (tab split, newline split, TrimSpace, Fprintln) for hostile input.
func FuzzCompletionValues(f *testing.F) {
	// Two-string corpus: (completion value, description).
	f.Add("value", "description")
	for _, s := range fuzzSeeds {
		f.Add(s, "desc-"+s)
		f.Add("val-"+s, s)
	}

	f.Fuzz(func(t *testing.T, value, desc string) {
		root := &Command{
			Use: "fuzzroot",
			Run: emptyRun,
			ValidArgsFunction: func(cmd *Command, args []string, toComplete string) ([]Completion, ShellCompDirective) {
				return []Completion{
					value,
					CompletionWithDesc(value, desc),
				}, ShellCompDirectiveNoFileComp
			},
		}

		// Drive both the with-description and no-description request commands,
		// which take slightly different branches in the emit loop.
		if _, err := executeCommand(root, ShellCompRequestCmd, ""); err != nil {
			_ = err
		}

		root2 := &Command{
			Use: "fuzzroot",
			Run: emptyRun,
			ValidArgsFunction: func(cmd *Command, args []string, toComplete string) ([]Completion, ShellCompDirective) {
				return []Completion{CompletionWithDesc(value, desc)}, ShellCompDirectiveDefault
			},
		}
		if _, err := executeCommand(root2, ShellCompNoDescRequestCmd, ""); err != nil {
			_ = err
		}
	})
}

// FuzzFlagCompletionValues exercises the runtime `__complete` emit path via a
// registered *flag* completion function (RegisterFlagCompletionFunc) rather
// than a ValidArgsFunction, covering the flag-value branch of getCompletions.
func FuzzFlagCompletionValues(f *testing.F) {
	f.Add("value", "description")
	for _, s := range fuzzSeeds {
		f.Add(s, "desc-"+s)
		f.Add("val-"+s, s)
	}

	f.Fuzz(func(t *testing.T, value, desc string) {
		root := &Command{
			Use: "fuzzroot",
			Run: emptyRun,
		}
		root.Flags().String("fuzzflag", "", "a fuzzed flag")
		if err := root.RegisterFlagCompletionFunc("fuzzflag",
			func(cmd *Command, args []string, toComplete string) ([]Completion, ShellCompDirective) {
				return []Completion{
					value,
					CompletionWithDesc(value, desc),
				}, ShellCompDirectiveNoFileComp
			}); err != nil {
			t.Fatalf("failed to register flag completion func: %v", err)
		}

		// Request completion of the flag's value.
		if _, err := executeCommand(root, ShellCompRequestCmd, "--fuzzflag", ""); err != nil {
			_ = err
		}
	})
}
