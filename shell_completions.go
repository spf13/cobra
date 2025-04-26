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
	"github.com/spf13/pflag"
)

// MarkFlagRequired instructs the various shell completion implementations to prioritize the named flag when performing completion, and causes your command to report an error if invoked without the flag.
func (c *Command) MarkFlagRequired(name string) error {
	return MarkFlagRequired(c.Flags(), name)
}

// MarkPersistentFlagRequired instructs the various shell completion implementations to prioritize the named persistent flag when performing completion, and causes your command to report an error if invoked without the flag.
func (c *Command) MarkPersistentFlagRequired(name string) error {
	return MarkFlagRequired(c.PersistentFlags(), name)
}

// MarkFlagRequired instructs the various shell completion implementations to prioritize the named flag when performing completion,
// and causes your command to report an error if invoked without the flag.
// It takes a pointer to a pflag.FlagSet and the name of the flag as parameters.
// The function returns an error if setting the annotation fails.
func MarkFlagRequired(flags *pflag.FlagSet, name string) error {
	return flags.SetAnnotation(name, BashCompOneRequiredFlag, []string{"true"})
}

// MarkFlagFilename instructs the various shell completion implementations to
// limit completions for the named flag to the specified file extensions.
func (c *Command) MarkFlagFilename(name string, extensions ...string) error {
	return MarkFlagFilename(c.Flags(), name, extensions...)
}

// MarkFlagCustom adds the BashCompCustom annotation to the named flag, if it exists.
// The bash completion script will call the bash function `f` for the flag.
//
// This will only work for bash completion. It is recommended to instead use
// `c.RegisterFlagCompletionFunc(...)` which allows registering a Go function that works across all shells.
//
// Parameters:
//   - name: the name of the flag to annotate
//   - f: the name of the bash function to call for bash completion
//
// Returns:
//   - error if marking the flag fails, nil otherwise
func (c *Command) MarkFlagCustom(name string, f string) error {
	return MarkFlagCustom(c.Flags(), name, f)
}

// MarkPersistentFlagFilename instructs the various shell completion
// implementations to limit completions for the named persistent flag to the
// specified file extensions.
func (c *Command) MarkPersistentFlagFilename(name string, extensions ...string) error {
	return MarkFlagFilename(c.PersistentFlags(), name, extensions...)
}

// MarkFlagFilename instructs the various shell completion implementations to limit completions for the named flag to the specified file extensions.
//
// Parameters:
// - flags: A pointer to a pflag.FlagSet containing the flags.
// - name: The name of the flag to be marked.
// - extensions: Variable number of string arguments representing the file extensions to limit completions to.
//
// Returns:
// - error: An error if setting the annotation fails, otherwise nil.
func MarkFlagFilename(flags *pflag.FlagSet, name string, extensions ...string) error {
	return flags.SetAnnotation(name, BashCompFilenameExt, extensions)
}

// MarkFlagCustom adds the BashCompCustom annotation to the named flag, if it exists.
// The bash completion script will call the bash function `f` for the flag.
//
// This will only work for bash completion.
// It is recommended to instead use c.RegisterFlagCompletionFunc(...) which allows
// to register a Go function which will work across all shells.
//
// Parameters:
//   - flags: The FlagSet containing the flags.
//   - name: The name of the flag to annotate.
//   - f: The bash function to call for completion.
//
// Returns:
//   - error if setting the annotation fails.
func MarkFlagCustom(flags *pflag.FlagSet, name string, f string) error {
	return flags.SetAnnotation(name, BashCompCustom, []string{f})
}

// MarkFlagDirname instructs the various shell completion implementations to
// limit completions for the named flag to directory names. It takes a pointer to a Command and a flag name as parameters.
// The function calls MarkFlagDirname on the command's flags with the provided name and returns any errors encountered during the operation.
func (c *Command) MarkFlagDirname(name string) error {
	return MarkFlagDirname(c.Flags(), name)
}

// MarkPersistentFlagDirname instructs the various shell completion
// implementations to limit completions for the named persistent flag to
// directory names.
//
// Parameters:
//   - c: The command instance that contains the persistent flags.
//   - name: The name of the persistent flag to mark.
//
// Returns:
//   - error: If any error occurs during the marking process, it is returned here.
func (c *Command) MarkPersistentFlagDirname(name string) error {
	return MarkFlagDirname(c.PersistentFlags(), name)
}

// MarkFlagDirname instructs the various shell completion implementations to limit completions for the named flag to directory names.
// Parameters:
//   - flags: A pointer to a pflag.FlagSet containing all available flags.
//   - name: The name of the flag to be modified.
// Returns:
//   - error: If an error occurs during the annotation setting, it is returned. Otherwise, nil is returned.
func MarkFlagDirname(flags *pflag.FlagSet, name string) error {
	return flags.SetAnnotation(name, BashCompSubdirsInDir, []string{})
}
