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

// Package cobra is a commander providing a simple interface to create powerful modern CLI interfaces.
// In addition to providing an interface, Cobra simultaneously provides a controller to organize your application code.
package cobra

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	flag "github.com/spf13/pflag"
)

const (
	FlagSetByCobraAnnotation     = "cobra_annotation_flag_set_by_cobra"
	CommandDisplayNameAnnotation = "cobra_annotation_command_display_name"

	helpFlagName    = "help"
	helpCommandName = "help"
)

// FParseErrWhitelist configures Flag parse errors to be ignored
type FParseErrWhitelist flag.ParseErrorsWhitelist

// Group Structure to manage groups for commands
type Group struct {
	ID    string
	Title string
}

// Command is just that, a command for your application.
// E.g.  'go run ...' - 'run' is the command. Cobra requires
// you to define the usage and description as part of your command
// definition to ensure usability.
type Command struct {
	// Use is the one-line usage message.
	// Recommended syntax is as follows:
	//   [ ] identifies an optional argument. Arguments that are not enclosed in brackets are required.
	//   ... indicates that you can specify multiple values for the previous argument.
	//   |   indicates mutually exclusive information. You can use the argument to the left of the separator or the
	//       argument to the right of the separator. You cannot use both arguments in a single use of the command.
	//   { } delimits a set of mutually exclusive arguments when one of the arguments is required. If the arguments are
	//       optional, they are enclosed in brackets ([ ]).
	// Example: add [-F file | -D dir]... [-f format] profile
	Use string

	// Aliases is an array of aliases that can be used instead of the first word in Use.
	Aliases []string

	// SuggestFor is an array of command names for which this command will be suggested -
	// similar to aliases but only suggests.
	SuggestFor []string

	// Short is the short description shown in the 'help' output.
	Short string

	// The group id under which this subcommand is grouped in the 'help' output of its parent.
	GroupID string

	// Long is the long message shown in the 'help <this-command>' output.
	Long string

	// Example is examples of how to use the command.
	Example string

	// ValidArgs is list of all valid non-flag arguments that are accepted in shell completions
	ValidArgs []Completion
	// ValidArgsFunction is an optional function that provides valid non-flag arguments for shell completion.
	// It is a dynamic version of using ValidArgs.
	// Only one of ValidArgs and ValidArgsFunction can be used for a command.
	ValidArgsFunction CompletionFunc

	// Expected arguments
	Args PositionalArgs

	// ArgAliases is List of aliases for ValidArgs.
	// These are not suggested to the user in the shell completion,
	// but accepted if entered manually.
	ArgAliases []string

	// BashCompletionFunction is custom bash functions used by the legacy bash autocompletion generator.
	// For portability with other shells, it is recommended to instead use ValidArgsFunction
	BashCompletionFunction string

	// Deprecated defines, if this command is deprecated and should print this string when used.
	Deprecated string

	// Annotations are key/value pairs that can be used by applications to identify or
	// group commands or set special options.
	Annotations map[string]string

	// Version defines the version for this command. If this value is non-empty and the command does not
	// define a "version" flag, a "version" boolean flag will be added to the command and, if specified,
	// will print content of the "Version" variable. A shorthand "v" flag will also be added if the
	// command does not define one.
	Version string

	// The *Run functions are executed in the following order:
	//   * PersistentPreRun()
	//   * PreRun()
	//   * Run()
	//   * PostRun()
	//   * PersistentPostRun()
	// All functions get the same args, the arguments after the command name.
	// The *PreRun and *PostRun functions will only be executed if the Run function of the current
	// command has been declared.
	//
	// PersistentPreRun: children of this command will inherit and execute.
	PersistentPreRun func(cmd *Command, args []string)
	// PersistentPreRunE: PersistentPreRun but returns an error.
	PersistentPreRunE func(cmd *Command, args []string) error
	// PreRun: children of this command will not inherit.
	PreRun func(cmd *Command, args []string)
	// PreRunE: PreRun but returns an error.
	PreRunE func(cmd *Command, args []string) error
	// Run: Typically the actual work function. Most commands will only implement this.
	Run func(cmd *Command, args []string)
	// RunE: Run but returns an error.
	RunE func(cmd *Command, args []string) error
	// PostRun: run after the Run command.
	PostRun func(cmd *Command, args []string)
	// PostRunE: PostRun but returns an error.
	PostRunE func(cmd *Command, args []string) error
	// PersistentPostRun: children of this command will inherit and execute after PostRun.
	PersistentPostRun func(cmd *Command, args []string)
	// PersistentPostRunE: PersistentPostRun but returns an error.
	PersistentPostRunE func(cmd *Command, args []string) error

	// groups for subcommands
	commandgroups []*Group

	// args is actual args parsed from flags.
	args []string
	// flagErrorBuf contains all error messages from pflag.
	flagErrorBuf *bytes.Buffer
	// flags is full set of flags.
	flags *flag.FlagSet
	// pflags contains persistent flags.
	pflags *flag.FlagSet
	// lflags contains local flags.
	// This field does not represent internal state, it's used as a cache to optimise LocalFlags function call
	lflags *flag.FlagSet
	// iflags contains inherited flags.
	// This field does not represent internal state, it's used as a cache to optimise InheritedFlags function call
	iflags *flag.FlagSet
	// parentsPflags is all persistent flags of cmd's parents.
	parentsPflags *flag.FlagSet
	// globNormFunc is the global normalization function
	// that we can use on every pflag set and children commands
	globNormFunc func(f *flag.FlagSet, name string) flag.NormalizedName

	// usageFunc is usage func defined by user.
	usageFunc func(*Command) error
	// usageTemplate is usage template defined by user.
	usageTemplate *tmplFunc
	// flagErrorFunc is func defined by user and it's called when the parsing of
	// flags returns an error.
	flagErrorFunc func(*Command, error) error
	// helpTemplate is help template defined by user.
	helpTemplate *tmplFunc
	// helpFunc is help func defined by user.
	helpFunc func(*Command, []string)
	// helpCommand is command with usage 'help'. If it's not defined by user,
	// cobra uses default help command.
	helpCommand *Command
	// helpCommandGroupID is the group id for the helpCommand
	helpCommandGroupID string

	// completionCommandGroupID is the group id for the completion command
	completionCommandGroupID string

	// versionTemplate is the version template defined by user.
	versionTemplate *tmplFunc

	// errPrefix is the error message prefix defined by user.
	errPrefix string

	// inReader is a reader defined by the user that replaces stdin
	inReader io.Reader
	// outWriter is a writer defined by the user that replaces stdout
	outWriter io.Writer
	// errWriter is a writer defined by the user that replaces stderr
	errWriter io.Writer

	// FParseErrWhitelist flag parse errors to be ignored
	FParseErrWhitelist FParseErrWhitelist

	// CompletionOptions is a set of options to control the handling of shell completion
	CompletionOptions CompletionOptions

	// commandsAreSorted defines, if command slice are sorted or not.
	commandsAreSorted bool
	// commandCalledAs is the name or alias value used to call this command.
	commandCalledAs struct {
		name   string
		called bool
	}

	ctx context.Context

	// commands is the list of commands supported by this program.
	commands []*Command
	// parent is a parent command for this command.
	parent *Command
	// Max lengths of commands' string lengths for use in padding.
	commandsMaxUseLen         int
	commandsMaxCommandPathLen int
	commandsMaxNameLen        int

	// TraverseChildren parses flags on all parents before executing child command.
	TraverseChildren bool

	// Hidden defines, if this command is hidden and should NOT show up in the list of available commands.
	Hidden bool

	// SilenceErrors is an option to quiet errors down stream.
	SilenceErrors bool

	// SilenceUsage is an option to silence usage when an error occurs.
	SilenceUsage bool

	// DisableFlagParsing disables the flag parsing.
	// If this is true all flags will be passed to the command as arguments.
	DisableFlagParsing bool

	// DisableAutoGenTag defines, if gen tag ("Auto generated by spf13/cobra...")
	// will be printed by generating docs for this command.
	DisableAutoGenTag bool

	// DisableFlagsInUseLine will disable the addition of [flags] to the usage
	// line of a command when printing help or generating docs
	DisableFlagsInUseLine bool

	// DisableSuggestions disables the suggestions based on Levenshtein distance
	// that go along with 'unknown command' messages.
	DisableSuggestions bool

	// SuggestionsMinimumDistance defines minimum levenshtein distance to display suggestions.
	// Must be > 0.
	SuggestionsMinimumDistance int
}

// Command represents a single CLI command and its options.
func (c *Command) Context() context.Context {
	return c.ctx
}

// SetContext sets the context for the command, which can be overridden by subsequent calls to ExecuteContext or ExecuteContextC. This method allows you to control the lifecycle and cancellation of operations associated with the command.
func (c *Command) SetContext(ctx context.Context) {
	c.ctx = ctx
}

// SetArgs sets the arguments for the command. By default, it uses `os.Args[1:]`. This method allows overriding the default arguments, which is particularly useful during testing.
func (c *Command) SetArgs(a []string) {
	c.args = a
}

// SetOutput sets the destination for usage and error messages.
//
// If output is nil, os.Stderr is used.
//
// Deprecated: Use SetOut and/or SetErr instead
func (c *Command) SetOutput(output io.Writer) {
	c.outWriter = output
	c.errWriter = output
}

// SetOut sets the destination for usage messages.
// If newOut is nil, os.Stdout is used.
func (c *Command) SetOut(newOut io.Writer) {
	c.outWriter = newOut
}

// SetErr sets the destination for error messages.
// If newErr is nil, os.Stderr is used.
func (c *Command) SetErr(newErr io.Writer) {
	c.errWriter = newErr
}

// SetIn sets the source for input data. If newIn is nil, os.Stdin is used.
func (c *Command) SetIn(newIn io.Reader) {
	c.inReader = newIn
}

// SetUsageFunc sets a custom usage function for the Command. This function allows you to define how the command's usage should be displayed, providing flexibility in customization. The provided function receives a pointer to the Command and can return an error if there is an issue with the usage definition.
// Parameters:
//   - f: A function that takes a pointer to a Command and returns an error.
// Returns:
//   - None
// Errors:
//   - Any errors returned by the custom usage function.
func (c *Command) SetUsageFunc(f func(*Command) error) {
	c.usageFunc = f
}

// SetUsageTemplate sets the custom usage template for the command. If an empty string is provided, it resets to the default template.
// Parameters:
//   - s: A string representing the custom usage template.
// Returns:
//   - None.
// Errors:
//   - None.
func (c *Command) SetUsageTemplate(s string) {
	if s == "" {
		c.usageTemplate = nil
		return
	}
	c.usageTemplate = tmpl(s)
}

// SetFlagErrorFunc sets a function to generate an error when flag parsing fails. The provided function `f` is called with the command instance and the error that occurred during flag parsing. This allows customization of error handling for command-line flags in your application.
func (c *Command) SetFlagErrorFunc(f func(*Command, error) error) {
	c.flagErrorFunc = f
}

// SetHelpFunc sets a help function for the command. This function can be defined by the application to provide custom help functionality when invoked. The provided function takes a pointer to the Command and a slice of strings representing arguments, and is expected to handle displaying help information accordingly.
func (c *Command) SetHelpFunc(f func(*Command, []string)) {
	c.helpFunc = f
}

// SetHelpCommand sets the help command for the receiver.
// It takes a pointer to a Command and assigns it to the receiver's helpCommand field.
func (c *Command) SetHelpCommand(cmd *Command) {
	c.helpCommand = cmd
}

// SetHelpCommandGroupID sets the group id of the help command.
func (c *Command) SetHelpCommandGroupID(groupID string) {
	if c.helpCommand != nil {
		c.helpCommand.GroupID = groupID
	}
	// helpCommandGroupID is used if no helpCommand is defined by the user
	c.helpCommandGroupID = groupID
}

// SetCompletionCommandGroupID sets the group id of the completion command.
//
// Parameters:
//   - groupID: The group ID to be set for the completion command.
//
// This method updates the root command's completion command group ID. If no completion
// command is defined by the user, this group ID will be used.
func (c *Command) SetCompletionCommandGroupID(groupID string) {
	// completionCommandGroupID is used if no completion command is defined by the user
	c.Root().completionCommandGroupID = groupID
}

// SetHelpTemplate sets the help template to be used by the command. If an empty string is provided, the default template will be used.
//
// Parameters:
//   - s: A string representing the custom template for help. An empty string resets to the default template.
//
// Returns:
//   None
func (c *Command) SetHelpTemplate(s string) {
	if s == "" {
		c.helpTemplate = nil
		return
	}
	c.helpTemplate = tmpl(s)
}

// SetVersionTemplate sets the version template to be used. Application can use it to set custom templates.
// If the provided string is empty, it clears the version template.
// Parameters:
//   - s: The custom version template as a string.
func (c *Command) SetVersionTemplate(s string) {
	if s == "" {
		c.versionTemplate = nil
		return
	}
	c.versionTemplate = tmpl(s)
}

// SetErrPrefix sets the error message prefix for the Command instance. The application can use this method to set a custom prefix for error messages.
func (c *Command) SetErrPrefix(s string) {
	c.errPrefix = s
}

// SetGlobalNormalizationFunc sets a normalization function to all flag sets and also to child commands.
// The user should not have a cyclic dependency on commands.
func (c *Command) SetGlobalNormalizationFunc(n func(f *flag.FlagSet, name string) flag.NormalizedName) {
	c.Flags().SetNormalizeFunc(n)
	c.PersistentFlags().SetNormalizeFunc(n)
	c.globNormFunc = n

	for _, command := range c.commands {
		command.SetGlobalNormalizationFunc(n)
	}
}

// OutOrStdout returns an io.Writer that outputs data to stdout, or a custom output if set.
func (c *Command) OutOrStdout() io.Writer {
	return c.getOut(os.Stdout)
}

// OutOrStderr returns the standard error writer for the command.
// If no specific output is set, it defaults to writing to os.Stderr.
func (c *Command) OutOrStderr() io.Writer {
	return c.getOut(os.Stderr)
}

// ErrOrStderr returns the standard error output writer.
// If the command is not running, it returns os.Stderr.
// It is a helper function for directing errors to the appropriate output.
//
// Returns:
//   - io.Writer: The writer to which errors should be written.
func (c *Command) ErrOrStderr() io.Writer {
	return c.getErr(os.Stderr)
}

// InOrStdin returns an io.Reader that provides input from the command's standard input if set, or falls back to os.Stdin otherwise.
func (c *Command) InOrStdin() io.Reader {
	return c.getIn(os.Stdin)
}

// getOut returns the output writer for the command. If a specific output writer is defined, it returns that; otherwise, it recursively calls the parent command's getOut method until it finds one or reaches the default output writer.
// Parameters:
//   - def: The default output writer to use if no specific writer is found.
// Returns:
//   - io.Writer: The output writer for the command.
func (c *Command) getOut(def io.Writer) io.Writer {
	if c.outWriter != nil {
		return c.outWriter
	}
	if c.HasParent() {
		return c.parent.getOut(def)
	}
	return def
}

// getErr returns the error writer for the command, or the default if none is set. It recursively checks parent commands if necessary.
// Parameters:
//   - def: The default error writer to use if no specific writer is found.
// Returns:
//   - io.Writer: The error writer associated with the command or the provided default.
// Errors:
//   - None.
func (c *Command) getErr(def io.Writer) io.Writer {
	if c.errWriter != nil {
		return c.errWriter
	}
	if c.HasParent() {
		return c.parent.getErr(def)
	}
	return def
}

// GetIn returns the input reader for the command.
// If a specific input reader is set, it returns that reader; otherwise,
// if the command has a parent, it recursively calls the parent's getIn method.
// If there are no parent readers, it returns the default reader provided.
func (c *Command) getIn(def io.Reader) io.Reader {
	if c.inReader != nil {
		return c.inReader
	}
	if c.HasParent() {
		return c.parent.getIn(def)
	}
	return def
}

// UsageFunc returns the usage function for the command, starting from this command and moving up to parent commands.
// If no specific usage function is set, it falls back to a default implementation that prints the usage template with persistent flags merged.
func (c *Command) UsageFunc() (f func(*Command) error) {
	if c.usageFunc != nil {
		return c.usageFunc
	}
	if c.HasParent() {
		return c.Parent().UsageFunc()
	}
	return func(c *Command) error {
		c.mergePersistentFlags()
		fn := c.getUsageTemplateFunc()
		err := fn(c.OutOrStderr(), c)
		if err != nil {
			c.PrintErrln(err)
		}
		return err
	}
}

// getUsageTemplateFunc returns the usage template function for the command. If no specific template is set for the command,
// it recursively searches up the command tree until it finds a parent with a template or reaches the root, returning the default template if none found.
// The returned function takes an io.Writer to write to and an interface{} containing data to be used in the template.
// It returns an error if there is an issue executing the template.
func (c *Command) getUsageTemplateFunc() func(w io.Writer, data interface{}) error {
	if c.usageTemplate != nil {
		return c.usageTemplate.fn
	}

	if c.HasParent() {
		return c.parent.getUsageTemplateFunc()
	}
	return defaultUsageFunc
}

// Usage returns an error indicating the usage of the command.
// It is invoked when a user inputs invalid data.
// This method can be customized by users through overriding UsageFunc.
func (c *Command) Usage() error {
	return c.UsageFunc()(c)
}

// HelpFunc returns the function set by SetHelpFunc for this command or a parent, or it returns a function with default help behavior. If the current command has no help function and does not have a parent, it provides a fallback that merges persistent flags and renders the help template to stdout. The returned function takes two parameters: the command itself and a slice of strings representing arguments, and returns an error if rendering the help fails.

// HasParent checks if the current command has a parent command.
// Returns true if there is a parent command, false otherwise.
func (c *Command) HelpFunc() func(*Command, []string) {
	if c.helpFunc != nil {
		return c.helpFunc
	}
	if c.HasParent() {
		return c.Parent().HelpFunc()
	}
	return func(c *Command, a []string) {
		c.mergePersistentFlags()
		fn := c.getHelpTemplateFunc()
		// The help should be sent to stdout
		// See https://github.com/spf13/cobra/issues/1002
		err := fn(c.OutOrStdout(), c)
		if err != nil {
			c.PrintErrln(err)
		}
	}
}

// GetHelpTemplateFunc returns the help template function for the command, going up the command tree if necessary. If the current command has a help template, it returns that; otherwise, it recursively checks its parent commands until it finds one with a help template or reaches the root, returning the default help function if no template is found.
func (c *Command) getHelpTemplateFunc() func(w io.Writer, data interface{}) error {
	if c.helpTemplate != nil {
		return c.helpTemplate.fn
	}

	if c.HasParent() {
		return c.parent.getHelpTemplateFunc()
	}

	return defaultHelpFunc
}

// Help outputs the help information for the command.
// It is invoked when a user requests help for a specific command using 'help [command]'.
// This method allows for custom implementation by overriding the HelpFunc field of the Command struct.
func (c *Command) Help() error {
	c.HelpFunc()(c, []string{})
	return nil
}

// UsageString returns the usage string of the command.
func (c *Command) UsageString() string {
	// Storing normal writers
	tmpOutput := c.outWriter
	tmpErr := c.errWriter

	bb := new(bytes.Buffer)
	c.outWriter = bb
	c.errWriter = bb

	CheckErr(c.Usage())

	// Setting things back to normal
	c.outWriter = tmpOutput
	c.errWriter = tmpErr

	return bb.String()
}

// FlagErrorFunc returns the function set by SetFlagErrorFunc for this command or a parent.
// If no function is set, it returns a function that returns the original error.
func (c *Command) FlagErrorFunc() (f func(*Command, error) error) {
	if c.flagErrorFunc != nil {
		return c.flagErrorFunc
	}

	if c.HasParent() {
		return c.parent.FlagErrorFunc()
	}
	return func(c *Command, err error) error {
		return err
	}
}

var minUsagePadding = 25

// UsagePadding returns the padding required for the command's usage based on its parent command's maximum usage length. If there is no parent or the parent's maximum usage length is less than minUsagePadding, it returns minUsagePadding. Otherwise, it returns the parent's maximum usage length.
func (c *Command) UsagePadding() int {
	if c.parent == nil || minUsagePadding > c.parent.commandsMaxUseLen {
		return minUsagePadding
	}
	return c.parent.commandsMaxUseLen
}

var minCommandPathPadding = 11

// CommandPathPadding returns the padding for the command path.
// It checks if the parent command exists and if its maximum command path length is greater than a minimum padding value.
// If true, it returns the maximum command path length; otherwise, it returns the minimum padding value.
func (c *Command) CommandPathPadding() int {
	if c.parent == nil || minCommandPathPadding > c.parent.commandsMaxCommandPathLen {
		return minCommandPathPadding
	}
	return c.parent.commandsMaxCommandPathLen
}

var minNamePadding = 11

// NamePadding calculates the padding needed for the command's name.
// It returns the minimum padding if the parent is nil or if the current maximum name length in the parent's commands exceeds minNamePadding.
// Otherwise, it returns the maximum name length of the parent's commands.
func (c *Command) NamePadding() int {
	if c.parent == nil || minNamePadding > c.parent.commandsMaxNameLen {
		return minNamePadding
	}
	return c.parent.commandsMaxNameLen
}

// Command represents a command in the application.
func (c *Command) UsageTemplate() string {
	if c.usageTemplate != nil {
		return c.usageTemplate.tmpl
	}

	if c.HasParent() {
		return c.parent.UsageTemplate()
	}
	return defaultUsageTemplate
}

// HelpTemplate returns the help template for the command.
//
// This function is kept for backwards-compatibility reasons. If a custom help template is set,
// it returns that template. Otherwise, it recursively calls the parent command's HelpTemplate
// method until it finds one or reaches the root command, returning the default help template if necessary.
func (c *Command) HelpTemplate() string {
	if c.helpTemplate != nil {
		return c.helpTemplate.tmpl
	}

	if c.HasParent() {
		return c.parent.HelpTemplate()
	}
	return defaultHelpTemplate
}

// VersionTemplate returns the version template for the command.
// This function is kept for backwards-compatibility reasons.
func (c *Command) VersionTemplate() string {
	if c.versionTemplate != nil {
		return c.versionTemplate.tmpl
	}

	if c.HasParent() {
		return c.parent.VersionTemplate()
	}
	return defaultVersionTemplate
}

// Command represents a command with its own version template function.
func (c *Command) getVersionTemplateFunc() func(w io.Writer, data interface{}) error {
	if c.versionTemplate != nil {
		return c.versionTemplate.fn
	}

	if c.HasParent() {
		return c.parent.getVersionTemplateFunc()
	}
	return defaultVersionFunc
}

// ErrPrefix returns the error message prefix for the command.
// If a custom prefix is set, it returns that.
// Otherwise, if the command has a parent, it recursively calls the parent's ErrPrefix.
// If no custom prefix or parent exists, it returns "Error:".
func (c *Command) ErrPrefix() string {
	if c.errPrefix != "" {
		return c.errPrefix
	}

	if c.HasParent() {
		return c.parent.ErrPrefix()
	}
	return "Error:"
}

// hasNoOptDefVal checks if the flag with the given name in the provided FlagSet has a non-empty NoOptDefVal.
// It returns true if such a flag exists and its NoOptDefVal is not empty, false otherwise.
// If the flag does not exist, it also returns false.
func hasNoOptDefVal(name string, fs *flag.FlagSet) bool {
	flag := fs.Lookup(name)
	if flag == nil {
		return false
	}
	return flag.NoOptDefVal != ""
}

// shortHasNoOptDefVal checks if a shorthand flag has a non-empty default value.
//
// It takes two parameters:
//   - name: the name of the flag to check.
//   - fs: a pointer to the FlagSet containing the flags.
//
// It returns true if the flag exists and its NoOptDefVal is not empty, false otherwise.
func shortHasNoOptDefVal(name string, fs *flag.FlagSet) bool {
	if len(name) == 0 {
		return false
	}

	flag := fs.ShorthandLookup(name[:1])
	if flag == nil {
		return false
	}
	return flag.NoOptDefVal != ""
}

// stripFlags processes the input arguments by removing any flags.
// It takes a slice of strings `args` and a pointer to a Command `c`.
// If no arguments are provided, it returns the original args.
// It merges persistent flags from the command.
// The function iterates over the arguments, stripping out flags based on their syntax.
// If "--" is encountered, flag processing stops.
// Flags starting with '--' or '-' without '=' are considered options.
// If a short flag ('-f') requires an argument and one is provided, it removes that argument from the list.
// It appends non-flag arguments to `commands` slice and returns this slice.
func stripFlags(args []string, c *Command) []string {
	if len(args) == 0 {
		return args
	}
	c.mergePersistentFlags()

	commands := []string{}
	flags := c.Flags()

Loop:
	for len(args) > 0 {
		s := args[0]
		args = args[1:]
		switch {
		case s == "--":
			// "--" terminates the flags
			break Loop
		case strings.HasPrefix(s, "--") && !strings.Contains(s, "=") && !hasNoOptDefVal(s[2:], flags):
			// If '--flag arg' then
			// delete arg from args.
			fallthrough // (do the same as below)
		case strings.HasPrefix(s, "-") && !strings.Contains(s, "=") && len(s) == 2 && !shortHasNoOptDefVal(s[1:], flags):
			// If '-f arg' then
			// delete 'arg' from args or break the loop if len(args) <= 1.
			if len(args) <= 1 {
				break Loop
			} else {
				args = args[1:]
				continue
			}
		case s != "" && !strings.HasPrefix(s, "-"):
			commands = append(commands, s)
		}
	}

	return commands
}

// argsMinusFirstX removes only the first occurrence of x from args. If no flags are present in args, it will remove the first non-flag element that matches x.
//
// It handles both long and short options, ensuring that flag values are not removed. If "--" is encountered in args, it stops processing further.
//
// Parameters:
//   - args: The slice of strings representing command line arguments.
//   - x: The string to remove from the args.
//
// Returns:
//   - A new slice of strings with the first occurrence of x removed (if found).
func (c *Command) argsMinusFirstX(args []string, x string) []string {
	if len(args) == 0 {
		return args
	}
	c.mergePersistentFlags()
	flags := c.Flags()

Loop:
	for pos := 0; pos < len(args); pos++ {
		s := args[pos]
		switch {
		case s == "--":
			// -- means we have reached the end of the parseable args. Break out of the loop now.
			break Loop
		case strings.HasPrefix(s, "--") && !strings.Contains(s, "=") && !hasNoOptDefVal(s[2:], flags):
			fallthrough
		case strings.HasPrefix(s, "-") && !strings.Contains(s, "=") && len(s) == 2 && !shortHasNoOptDefVal(s[1:], flags):
			// This is a flag without a default value, and an equal sign is not used. Increment pos in order to skip
			// over the next arg, because that is the value of this flag.
			pos++
			continue
		case !strings.HasPrefix(s, "-"):
			// This is not a flag or a flag value. Check to see if it matches what we're looking for, and if so,
			// return the args, excluding the one at this position.
			if s == x {
				ret := make([]string, 0, len(args)-1)
				ret = append(ret, args[:pos]...)
				ret = append(ret, args[pos+1:]...)
				return ret
			}
		}
	}
	return args
}

// isFlagArg checks if the provided argument is a flag.
// It returns true if the argument starts with "--" (e.g., --flag)
// or starts with "-" followed by a non-dash character (e.g., -f).
func isFlagArg(arg string) bool {
	return ((len(arg) >= 3 && arg[0:2] == "--") ||
		(len(arg) >= 2 && arg[0] == '-' && arg[1] != '-'))
}

// Find returns the target command given the arguments and command tree.
// It starts from the highest node and searches down.
// If the target command is found, it returns the command and any remaining arguments.
// If no matching command is found, it returns the current node and the original arguments.
func (c *Command) Find(args []string) (*Command, []string, error) {
	var innerfind func(*Command, []string) (*Command, []string)

	innerfind = func(c *Command, innerArgs []string) (*Command, []string) {
		argsWOflags := stripFlags(innerArgs, c)
		if len(argsWOflags) == 0 {
			return c, innerArgs
		}
		nextSubCmd := argsWOflags[0]

		cmd := c.findNext(nextSubCmd)
		if cmd != nil {
			return innerfind(cmd, c.argsMinusFirstX(innerArgs, nextSubCmd))
		}
		return c, innerArgs
	}

	commandFound, a := innerfind(c, args)
	if commandFound.Args == nil {
		return commandFound, a, legacyArgs(commandFound, stripFlags(a, commandFound))
	}
	return commandFound, a, nil
}

// findSuggestions returns a string of suggestions based on the given argument.
// It checks if suggestions are disabled or if the minimum distance is not set,
// and adjusts them accordingly. If there are any suggestions, it formats them into a string
// and appends them to the output.
// Returns an empty string if no suggestions are available or suggestions are disabled.
func (c *Command) findSuggestions(arg string) string {
	if c.DisableSuggestions {
		return ""
	}
	if c.SuggestionsMinimumDistance <= 0 {
		c.SuggestionsMinimumDistance = 2
	}
	var sb strings.Builder
	if suggestions := c.SuggestionsFor(arg); len(suggestions) > 0 {
		sb.WriteString("\n\nDid you mean this?\n")
		for _, s := range suggestions {
			_, _ = fmt.Fprintf(&sb, "\t%v\n", s)
		}
	}
	return sb.String()
}

// findNext searches for the next command based on the given name or alias.
// It returns the first matching command if found, otherwise returns nil.
// If multiple commands match but EnablePrefixMatching is enabled, it returns the first match.
// If no matches are found, it returns nil.
func (c *Command) findNext(next string) *Command {
	matches := make([]*Command, 0)
	for _, cmd := range c.commands {
		if commandNameMatches(cmd.Name(), next) || cmd.HasAlias(next) {
			cmd.commandCalledAs.name = next
			return cmd
		}
		if EnablePrefixMatching && cmd.hasNameOrAliasPrefix(next) {
			matches = append(matches, cmd)
		}
	}

	if len(matches) == 1 {
		// Temporarily disable gosec G602, which produces a false positive.
		// See https://github.com/securego/gosec/issues/1005.
		return matches[0] // #nosec G602
	}

	return nil
}

// Traverse parses command flags and arguments by traversing the command tree.
// It returns the final command and any remaining arguments, or an error if parsing fails.
func (c *Command) Traverse(args []string) (*Command, []string, error) {
	flags := []string{}
	inFlag := false

	for i, arg := range args {
		switch {
		// A long flag with a space separated value
		case strings.HasPrefix(arg, "--") && !strings.Contains(arg, "="):
			// TODO: this isn't quite right, we should really check ahead for 'true' or 'false'
			inFlag = !hasNoOptDefVal(arg[2:], c.Flags())
			flags = append(flags, arg)
			continue
		// A short flag with a space separated value
		case strings.HasPrefix(arg, "-") && !strings.Contains(arg, "=") && len(arg) == 2 && !shortHasNoOptDefVal(arg[1:], c.Flags()):
			inFlag = true
			flags = append(flags, arg)
			continue
		// The value for a flag
		case inFlag:
			inFlag = false
			flags = append(flags, arg)
			continue
		// A flag without a value, or with an `=` separated value
		case isFlagArg(arg):
			flags = append(flags, arg)
			continue
		}

		cmd := c.findNext(arg)
		if cmd == nil {
			return c, args, nil
		}

		if err := c.ParseFlags(flags); err != nil {
			return nil, args, err
		}
		return cmd.Traverse(args[i+1:])
	}
	return c, args, nil
}

// SuggestionsFor returns a list of suggestions for the given typedName based on available commands.
// It considers both levenshtein distance and prefix matching to find relevant suggestions.
// Parameters:
//   - typedName: The name that needs suggestions.
// Returns:
//   - A slice of strings containing suggested command names.
func (c *Command) SuggestionsFor(typedName string) []string {
	suggestions := []string{}
	for _, cmd := range c.commands {
		if cmd.IsAvailableCommand() {
			levenshteinDistance := ld(typedName, cmd.Name(), true)
			suggestByLevenshtein := levenshteinDistance <= c.SuggestionsMinimumDistance
			suggestByPrefix := strings.HasPrefix(strings.ToLower(cmd.Name()), strings.ToLower(typedName))
			if suggestByLevenshtein || suggestByPrefix {
				suggestions = append(suggestions, cmd.Name())
			}
			for _, explicitSuggestion := range cmd.SuggestFor {
				if strings.EqualFold(typedName, explicitSuggestion) {
					suggestions = append(suggestions, cmd.Name())
				}
			}
		}
	}
	return suggestions
}

// VisitParents traverses all parent commands starting from the current command.
// It invokes the provided function on each parent command and recursively visits their parents.
// If there are no parents, it does nothing.
func (c *Command) VisitParents(fn func(*Command)) {
	if c.HasParent() {
		fn(c.Parent())
		c.Parent().VisitParents(fn)
	}
}

// Root returns the root command of the given command, traversing up the parent chain until it reaches the root. If the command does not have a parent, it returns itself as the root.
func (c *Command) Root() *Command {
	if c.HasParent() {
		return c.Parent().Root()
	}
	return c
}

// ArgsLenAtDash returns the number of arguments that were provided before encountering a "--" flag during argument parsing. This can be useful for determining how many positional arguments have been specified up to a certain point in a command's flags.
func (c *Command) ArgsLenAtDash() int {
	return c.Flags().ArgsLenAtDash()
}

// execute runs the command with the provided arguments.
//
// It handles deprecated commands, initializes help and version flags,
// parses flags, checks for help or version requests, validates arguments,
// and executes pre-run hooks before running the main command logic.
// It also runs post-run hooks after the main logic completes.
//
// Parameters:
//   - a: A slice of strings representing the command line arguments.
//
// Returns:
//   - err: An error if any step fails; otherwise, nil.
func (c *Command) execute(a []string) (err error) {
	if c == nil {
		return fmt.Errorf("called Execute() on a nil Command")
	}

	if len(c.Deprecated) > 0 {
		c.Printf("Command %q is deprecated, %s\n", c.Name(), c.Deprecated)
	}

	// initialize help and version flag at the last point possible to allow for user
	// overriding
	c.InitDefaultHelpFlag()
	c.InitDefaultVersionFlag()

	err = c.ParseFlags(a)
	if err != nil {
		return c.FlagErrorFunc()(c, err)
	}

	// If help is called, regardless of other flags, return we want help.
	// Also say we need help if the command isn't runnable.
	helpVal, err := c.Flags().GetBool(helpFlagName)
	if err != nil {
		// should be impossible to get here as we always declare a help
		// flag in InitDefaultHelpFlag()
		c.Println("\"help\" flag declared as non-bool. Please correct your code")
		return err
	}

	if helpVal {
		return flag.ErrHelp
	}

	// for back-compat, only add version flag behavior if version is defined
	if c.Version != "" {
		versionVal, err := c.Flags().GetBool("version")
		if err != nil {
			c.Println("\"version\" flag declared as non-bool. Please correct your code")
			return err
		}
		if versionVal {
			fn := c.getVersionTemplateFunc()
			err := fn(c.OutOrStdout(), c)
			if err != nil {
				c.Println(err)
			}
			return err
		}
	}

	if !c.Runnable() {
		return flag.ErrHelp
	}

	c.preRun()

	defer c.postRun()

	argWoFlags := c.Flags().Args()
	if c.DisableFlagParsing {
		argWoFlags = a
	}

	if err := c.ValidateArgs(argWoFlags); err != nil {
		return err
	}

	parents := make([]*Command, 0, 5)
	for p := c; p != nil; p = p.Parent() {
		if EnableTraverseRunHooks {
			// When EnableTraverseRunHooks is set:
			// - Execute all persistent pre-runs from the root parent till this command.
			// - Execute all persistent post-runs from this command till the root parent.
			parents = append([]*Command{p}, parents...)
		} else {
			// Otherwise, execute only the first found persistent hook.
			parents = append(parents, p)
		}
	}
	for _, p := range parents {
		if p.PersistentPreRunE != nil {
			if err := p.PersistentPreRunE(c, argWoFlags); err != nil {
				return err
			}
			if !EnableTraverseRunHooks {
				break
			}
		} else if p.PersistentPreRun != nil {
			p.PersistentPreRun(c, argWoFlags)
			if !EnableTraverseRunHooks {
				break
			}
		}
	}
	if c.PreRunE != nil {
		if err := c.PreRunE(c, argWoFlags); err != nil {
			return err
		}
	} else if c.PreRun != nil {
		c.PreRun(c, argWoFlags)
	}

	if err := c.ValidateRequiredFlags(); err != nil {
		return err
	}
	if err := c.ValidateFlagGroups(); err != nil {
		return err
	}

	if c.RunE != nil {
		if err := c.RunE(c, argWoFlags); err != nil {
			return err
		}
	} else {
		c.Run(c, argWoFlags)
	}
	if c.PostRunE != nil {
		if err := c.PostRunE(c, argWoFlags); err != nil {
			return err
		}
	} else if c.PostRun != nil {
		c.PostRun(c, argWoFlags)
	}
	for p := c; p != nil; p = p.Parent() {
		if p.PersistentPostRunE != nil {
			if err := p.PersistentPostRunE(c, argWoFlags); err != nil {
				return err
			}
			if !EnableTraverseRunHooks {
				break
			}
		} else if p.PersistentPostRun != nil {
			p.PersistentPostRun(c, argWoFlags)
			if !EnableTraverseRunHooks {
				break
			}
		}
	}

	return nil
}

// preRun executes the initializers for the command.
func (c *Command) preRun() {
	for _, x := range initializers {
		x()
	}
}

// postRun is called after a command has been executed. It runs all registered finalizers.
func (c *Command) postRun() {
	for _, x := range finalizers {
		x()
	}
}

// ExecuteContext is the same as Execute(), but sets the ctx on the command. It allows setting a context for the command, which can be retrieved using cmd.Context() inside lifecycle or ValidArgs functions.
// This method ensures that any operations within the command can utilize the provided context for managing timeouts, cancellations, and other asynchronous behaviors.
// Parameters:
//   - ctx: The context to set on the command.
// Returns:
//   - error: If an error occurs during the execution of the command.
func (c *Command) ExecuteContext(ctx context.Context) error {
	c.ctx = ctx
	return c.Execute()
}

// Execute runs the command with the provided arguments, using os.Args[1:] by default if no arguments are provided.
// It traverses the command tree to find appropriate matches for commands and their corresponding flags.
// Returns an error if there is an issue during execution.
func (c *Command) Execute() error {
	_, err := c.ExecuteC()
	return err
}

// ExecuteContextC is the same as ExecuteC(), but sets the ctx on the command.
// Retrieve ctx by calling cmd.Context() inside your *Run lifecycle or ValidArgs
// functions.
func (c *Command) ExecuteContextC(ctx context.Context) (*Command, error) {
	c.ctx = ctx
	return c.ExecuteC()
}

// ExecuteC executes the command.
func (c *Command) ExecuteC() (cmd *Command, err error) {
	if c.ctx == nil {
		c.ctx = context.Background()
	}

	// Regardless of what command execute is called on, run on Root only
	if c.HasParent() {
		return c.Root().ExecuteC()
	}

	// windows hook
	if preExecHookFn != nil {
		preExecHookFn(c)
	}

	// initialize help at the last point to allow for user overriding
	c.InitDefaultHelpCmd()

	args := c.args

	// Workaround FAIL with "go test -v" or "cobra.test -test.v", see #155
	if c.args == nil && filepath.Base(os.Args[0]) != "cobra.test" {
		args = os.Args[1:]
	}

	// initialize the __complete command to be used for shell completion
	c.initCompleteCmd(args)

	// initialize the default completion command
	c.InitDefaultCompletionCmd(args...)

	// Now that all commands have been created, let's make sure all groups
	// are properly created also
	c.checkCommandGroups()

	var flags []string
	if c.TraverseChildren {
		cmd, flags, err = c.Traverse(args)
	} else {
		cmd, flags, err = c.Find(args)
	}
	if err != nil {
		// If found parse to a subcommand and then failed, talk about the subcommand
		if cmd != nil {
			c = cmd
		}
		if !c.SilenceErrors {
			c.PrintErrln(c.ErrPrefix(), err.Error())
			c.PrintErrf("Run '%v --help' for usage.\n", c.CommandPath())
		}
		return c, err
	}

	cmd.commandCalledAs.called = true
	if cmd.commandCalledAs.name == "" {
		cmd.commandCalledAs.name = cmd.Name()
	}

	// We have to pass global context to children command
	// if context is present on the parent command.
	if cmd.ctx == nil {
		cmd.ctx = c.ctx
	}

	err = cmd.execute(flags)
	if err != nil {
		// Always show help if requested, even if SilenceErrors is in
		// effect
		if errors.Is(err, flag.ErrHelp) {
			cmd.HelpFunc()(cmd, args)
			return cmd, nil
		}

		// If root command has SilenceErrors flagged,
		// all subcommands should respect it
		if !cmd.SilenceErrors && !c.SilenceErrors {
			c.PrintErrln(cmd.ErrPrefix(), err.Error())
		}

		// If root command has SilenceUsage flagged,
		// all subcommands should respect it
		if !cmd.SilenceUsage && !c.SilenceUsage {
			c.Println(cmd.UsageString())
		}
	}
	return cmd, err
}

// ValidateArgs validates the arguments provided to a command.
//
// If the command's Args field is nil, it calls ArbitraryArgs with the command and arguments.
// Otherwise, it calls the Args function assigned to the command with the command and arguments.
//
// It returns an error if validation fails or if no args are provided for a non-nil Args function.
func (c *Command) ValidateArgs(args []string) error {
	if c.Args == nil {
		return ArbitraryArgs(c, args)
	}
	return c.Args(c, args)
}

// ValidateRequiredFlags validates all required flags are present and returns an error otherwise. It checks if the DisableFlagParsing option is enabled. If not, it iterates over all flags to find those marked as required but not set. If any required flags are missing, it returns an error listing them; otherwise, it returns nil.
func (c *Command) ValidateRequiredFlags() error {
	if c.DisableFlagParsing {
		return nil
	}

	flags := c.Flags()
	missingFlagNames := []string{}
	flags.VisitAll(func(pflag *flag.Flag) {
		requiredAnnotation, found := pflag.Annotations[BashCompOneRequiredFlag]
		if !found {
			return
		}
		if (requiredAnnotation[0] == "true") && !pflag.Changed {
			missingFlagNames = append(missingFlagNames, pflag.Name)
		}
	})

	if len(missingFlagNames) > 0 {
		return fmt.Errorf(`required flag(s) "%s" not set`, strings.Join(missingFlagNames, `", "`))
	}
	return nil
}

// checkCommandGroups checks if a command has been added to a group that does not exist.
// If so, it panics because it indicates a coding error that should be corrected. This function
// is typically used during the initialization or validation phase of the application to ensure
// that all command groups are properly defined and referenced.
func (c *Command) checkCommandGroups() {
	for _, sub := range c.commands {
		// if Group is not defined let the developer know right away
		if sub.GroupID != "" && !c.ContainsGroup(sub.GroupID) {
			panic(fmt.Sprintf("group id '%s' is not defined for subcommand '%s'", sub.GroupID, sub.CommandPath()))
		}

		sub.checkCommandGroups()
	}
}

// InitDefaultHelpFlag adds a default help flag to the command.
// It is automatically called when executing the command or by explicitly calling the help and usage functions.
// If the command already has a help flag, this method does nothing.
func (c *Command) InitDefaultHelpFlag() {
	c.mergePersistentFlags()
	if c.Flags().Lookup(helpFlagName) == nil {
		usage := "help for "
		name := c.DisplayName()
		if name == "" {
			usage += "this command"
		} else {
			usage += name
		}
		c.Flags().BoolP(helpFlagName, "h", false, usage)
		_ = c.Flags().SetAnnotation(helpFlagName, FlagSetByCobraAnnotation, []string{"true"})
	}
}

// InitDefaultVersionFlag adds a default version flag to the command.
// It is called automatically when the command is executed.
// If the command already has a version flag or if the Version field is empty, it does nothing.
func (c *Command) InitDefaultVersionFlag() {
	if c.Version == "" {
		return
	}

	c.mergePersistentFlags()
	if c.Flags().Lookup("version") == nil {
		usage := "version for "
		if c.Name() == "" {
			usage += "this command"
		} else {
			usage += c.DisplayName()
		}
		if c.Flags().ShorthandLookup("v") == nil {
			c.Flags().BoolP("version", "v", false, usage)
		} else {
			c.Flags().Bool("version", false, usage)
		}
		_ = c.Flags().SetAnnotation("version", FlagSetByCobraAnnotation, []string{"true"})
	}
}

// InitDefaultHelpCmd initializes the default help command for a Command.
// If the command already has a help command or if it has no subcommands, it does nothing.
func (c *Command) InitDefaultHelpCmd() {
	if !c.HasSubCommands() {
		return
	}

	if c.helpCommand == nil {
		c.helpCommand = &Command{
			Use:   "help [command]",
			Short: "Help about any command",
			Long: `Help provides help for any command in the application.
Simply type ` + c.DisplayName() + ` help [path to command] for full details.`,
			ValidArgsFunction: func(c *Command, args []string, toComplete string) ([]Completion, ShellCompDirective) {
				var completions []Completion
				cmd, _, e := c.Root().Find(args)
				if e != nil {
					return nil, ShellCompDirectiveNoFileComp
				}
				if cmd == nil {
					// Root help command.
					cmd = c.Root()
				}
				for _, subCmd := range cmd.Commands() {
					if subCmd.IsAvailableCommand() || subCmd == cmd.helpCommand {
						if strings.HasPrefix(subCmd.Name(), toComplete) {
							completions = append(completions, CompletionWithDesc(subCmd.Name(), subCmd.Short))
						}
					}
				}
				return completions, ShellCompDirectiveNoFileComp
			},
			Run: func(c *Command, args []string) {
				cmd, _, e := c.Root().Find(args)
				if cmd == nil || e != nil {
					c.Printf("Unknown help topic %#q\n", args)
					CheckErr(c.Root().Usage())
				} else {
					// FLow the context down to be used in help text
					if cmd.ctx == nil {
						cmd.ctx = c.ctx
					}

					cmd.InitDefaultHelpFlag()    // make possible 'help' flag to be shown
					cmd.InitDefaultVersionFlag() // make possible 'version' flag to be shown
					CheckErr(cmd.Help())
				}
			},
			GroupID: c.helpCommandGroupID,
		}
	}
	c.RemoveCommand(c.helpCommand)
	c.AddCommand(c.helpCommand)
}

// ResetCommands deletes the parent, subcommands, and help command associated with the receiver Command. It also clears any parent flag sets.
// This method is useful for resetting a Command instance to its initial state before reconfiguration.
// The receiver *Command should be non-nil, otherwise, this function will panic.
func (c *Command) ResetCommands() {
	c.parent = nil
	c.commands = nil
	c.helpCommand = nil
	c.parentsPflags = nil
}

// Sorts commands by their names.
type commandSorterByName []*Command

// Len returns the number of commands in the sorter.
func (c commandSorterByName) Len() int           { return len(c) }
// Swap swaps elements i and j in the commandSorterByName slice. It is used to implement the sort.Interface for sorting commands by name.
func (c commandSorterByName) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
// Less returns true if the name of the command at index i should sort before the command at index j. This is used by sorting algorithms to determine the order of elements in a slice of commands.
func (c commandSorterByName) Less(i, j int) bool { return c[i].Name() < c[j].Name() }

// Commands returns a sorted slice of child commands.
//
// If EnableCommandSorting is true and the commands are not already sorted, it will sort them by name using a custom sorter.
// The method marks the command list as sorted to avoid redundant sorting in future calls.
func (c *Command) Commands() []*Command {
	// do not sort commands if it already sorted or sorting was disabled
	if EnableCommandSorting && !c.commandsAreSorted {
		sort.Sort(commandSorterByName(c.commands))
		c.commandsAreSorted = true
	}
	return c.commands
}

// AddCommand adds one or more commands to this parent command.
// It iterates over the provided commands, checks if any of them is a child of itself (which would cause a panic),
// sets their parent to the current command, updates maximum lengths for usage, command path, and name,
// and appends them to the list of commands. If a global normalization function exists, it applies it to all added commands.
// It also marks the commands as unsorted.
func (c *Command) AddCommand(cmds ...*Command) {
	for i, x := range cmds {
		if cmds[i] == c {
			panic("Command can't be a child of itself")
		}
		cmds[i].parent = c
		// update max lengths
		usageLen := len(x.Use)
		if usageLen > c.commandsMaxUseLen {
			c.commandsMaxUseLen = usageLen
		}
		commandPathLen := len(x.CommandPath())
		if commandPathLen > c.commandsMaxCommandPathLen {
			c.commandsMaxCommandPathLen = commandPathLen
		}
		nameLen := len(x.Name())
		if nameLen > c.commandsMaxNameLen {
			c.commandsMaxNameLen = nameLen
		}
		// If global normalization function exists, update all children
		if c.globNormFunc != nil {
			x.SetGlobalNormalizationFunc(c.globNormFunc)
		}
		c.commands = append(c.commands, x)
		c.commandsAreSorted = false
	}
}

// Groups returns a slice of child command groups. It provides access to all sub-groups associated with the current command. The returned slice is not nil but may be empty if there are no child groups.
func (c *Command) Groups() []*Group {
	return c.commandgroups
}

// AllChildCommandsHaveGroup returns true if all subcommands of the command are assigned to a group.
// It iterates over each subcommand and checks if it is an available command or the help command, and if its GroupID is empty.
// If any subcommand meets these conditions but has an empty GroupID, the function returns false.
// Otherwise, it returns true.
func (c *Command) AllChildCommandsHaveGroup() bool {
	for _, sub := range c.commands {
		if (sub.IsAvailableCommand() || sub == c.helpCommand) && sub.GroupID == "" {
			return false
		}
	}
	return true
}

// ContainsGroup checks if a given groupID exists within the list of command groups associated with the Command instance.
// It iterates through each command group and compares its ID with the provided groupID.
// Returns true if the groupID is found, otherwise returns false.
func (c *Command) ContainsGroup(groupID string) bool {
	for _, x := range c.commandgroups {
		if x.ID == groupID {
			return true
		}
	}
	return false
}

// AddGroup appends one or more command groups to the current command.
func (c *Command) AddGroup(groups ...*Group) {
	c.commandgroups = append(c.commandgroups, groups...)
}

// RemoveCommand removes one or more commands from the receiver command.
// It takes a variadic parameter of commands to remove and updates the receiver's internal state accordingly.
func (c *Command) RemoveCommand(cmds ...*Command) {
	commands := []*Command{}
main:
	for _, command := range c.commands {
		for _, cmd := range cmds {
			if command == cmd {
				command.parent = nil
				continue main
			}
		}
		commands = append(commands, command)
	}
	c.commands = commands
	// recompute all lengths
	c.commandsMaxUseLen = 0
	c.commandsMaxCommandPathLen = 0
	c.commandsMaxNameLen = 0
	for _, command := range c.commands {
		usageLen := len(command.Use)
		if usageLen > c.commandsMaxUseLen {
			c.commandsMaxUseLen = usageLen
		}
		commandPathLen := len(command.CommandPath())
		if commandPathLen > c.commandsMaxCommandPathLen {
			c.commandsMaxCommandPathLen = commandPathLen
		}
		nameLen := len(command.Name())
		if nameLen > c.commandsMaxNameLen {
			c.commandsMaxNameLen = nameLen
		}
	}
}

// Print is a convenience method to print to the defined output. If no output is set, it falls back to using standard error.
// It takes a variadic number of interface{} arguments and prints them according to the format specified.
// Parameters:
//   - i ...interface{}: The data to be printed.
// There are no return values for this method.
func (c *Command) Print(i ...interface{}) {
	fmt.Fprint(c.OutOrStderr(), i...)
}

// Command represents a command with optional output redirection.
func (c *Command) Println(i ...interface{}) {
	c.Print(fmt.Sprintln(i...))
}

// Command represents a command with various options and settings.
func (c *Command) Printf(format string, i ...interface{}) {
	c.Print(fmt.Sprintf(format, i...))
}

// PrintErr prints to the error output of the Command. If the error output is not set, it defaults to printing to standard error.
// It takes a variadic number of interface{} arguments and formats them according to the default format for fmt.Print.
func (c *Command) PrintErr(i ...interface{}) {
	fmt.Fprint(c.ErrOrStderr(), i...)
}

// PrintErrln prints the provided arguments to the error output of the Command. If no error output is set, it defaults to printing to Stderr.
// It accepts a variable number of interface{} arguments, which are formatted and printed similarly to fmt.Println.
// This method provides a convenient way to handle errors by automatically directing them to an appropriate output.
func (c *Command) PrintErrln(i ...interface{}) {
	c.PrintErr(fmt.Sprintln(i...))
}

// PrintErrf formats according to a format specifier and writes to the command's error output. If no specific error output is set, it defaults to Stderr.
// The function uses Sprintf to format the string before passing it to PrintErr for further handling.
// Parameters:
//   - format: A format string in Go's printf style.
//   - i: Arguments to be formatted into the string.
// This method does not return any value but may print an error message if an error occurs during output.
func (c *Command) PrintErrf(format string, i ...interface{}) {
	c.PrintErr(fmt.Sprintf(format, i...))
}

// CommandPath returns the full path to this command, including its parent commands if any. If the command has no parent, it returns the display name of the command.
// It recursively concatenates the names of parent commands with the current command's name and a space in between each name.
// Returns:
//   - The full path to this command as a string.
func (c *Command) CommandPath() string {
	if c.HasParent() {
		return c.Parent().CommandPath() + " " + c.Name()
	}
	return c.DisplayName()
}

// DisplayName returns the name to display in help text.
// It checks for the CommandDisplayNameAnnotation in the command's annotations.
// If found, it returns the annotation value; otherwise, it returns the result of calling c.Name().
func (c *Command) DisplayName() string {
	if displayName, ok := c.Annotations[CommandDisplayNameAnnotation]; ok {
		return displayName
	}
	return c.Name()
}

// UseLine returns the full usage for a given command, including parents. It replaces the command name with its display name and appends flags if necessary.
// Parameters:
// - c: The Command instance for which to generate the usage line.
// Returns:
// - A string representing the full usage of the command.
func (c *Command) UseLine() string {
	var useline string
	use := strings.Replace(c.Use, c.Name(), c.DisplayName(), 1)
	if c.HasParent() {
		useline = c.parent.CommandPath() + " " + use
	} else {
		useline = use
	}
	if c.DisableFlagsInUseLine {
		return useline
	}
	if c.HasAvailableFlags() && !strings.Contains(useline, "[flags]") {
		useline += " [flags]"
	}
	return useline
}

// DebugFlags prints debug information about the command and its flags, including both local and persistent flags.
func (c *Command) DebugFlags() {
	c.Println("DebugFlags called on", c.Name())
	var debugflags func(*Command)

	debugflags = func(x *Command) {
		if x.HasFlags() || x.HasPersistentFlags() {
			c.Println(x.Name())
		}
		if x.HasFlags() {
			x.flags.VisitAll(func(f *flag.Flag) {
				if x.HasPersistentFlags() && x.persistentFlag(f.Name) != nil {
					c.Println("  -"+f.Shorthand+",", "--"+f.Name, "["+f.DefValue+"]", "", f.Value, "  [LP]")
				} else {
					c.Println("  -"+f.Shorthand+",", "--"+f.Name, "["+f.DefValue+"]", "", f.Value, "  [L]")
				}
			})
		}
		if x.HasPersistentFlags() {
			x.pflags.VisitAll(func(f *flag.Flag) {
				if x.HasFlags() {
					if x.flags.Lookup(f.Name) == nil {
						c.Println("  -"+f.Shorthand+",", "--"+f.Name, "["+f.DefValue+"]", "", f.Value, "  [P]")
					}
				} else {
					c.Println("  -"+f.Shorthand+",", "--"+f.Name, "["+f.DefValue+"]", "", f.Value, "  [P]")
				}
			})
		}
		c.Println(x.flagErrorBuf)
		if x.HasSubCommands() {
			for _, y := range x.commands {
				debugflags(y)
			}
		}
	}

	debugflags(c)
}

// Name returns the command's name: the first word in the use line.
func (c *Command) Name() string {
	name := c.Use
	i := strings.Index(name, " ")
	if i >= 0 {
		name = name[:i]
	}
	return name
}

// HasAlias checks if the given string is an alias of the command.
//
// Parameters:
//   - s: The string to check for being an alias.
//
// Returns:
//   - bool: True if the string is an alias, false otherwise.
func (c *Command) HasAlias(s string) bool {
	for _, a := range c.Aliases {
		if commandNameMatches(a, s) {
			return true
		}
	}
	return false
}

// CalledAs returns the command name or alias that was used to invoke
// this command or an empty string if the command has not been called.
func (c *Command) CalledAs() string {
	if c.commandCalledAs.called {
		return c.commandCalledAs.name
	}
	return ""
}

// hasNameOrAliasPrefix checks if the command's Name or any of its aliases start with the given prefix.
//
// Parameters:
//   - prefix: The string prefix to check against the command's Name and Aliases.
//
// Returns:
//   - bool: true if any of the Name or Aliases start with the prefix, false otherwise.
func (c *Command) hasNameOrAliasPrefix(prefix string) bool {
	if strings.HasPrefix(c.Name(), prefix) {
		c.commandCalledAs.name = c.Name()
		return true
	}
	for _, alias := range c.Aliases {
		if strings.HasPrefix(alias, prefix) {
			c.commandCalledAs.name = alias
			return true
		}
	}
	return false
}

// NameAndAliases returns a string containing the command name followed by its aliases, separated by commas.
// It takes no parameters and returns the formatted string.
func (c *Command) NameAndAliases() string {
	return strings.Join(append([]string{c.Name()}, c.Aliases...), ", ")
}

// HasExample returns true if the command has an example.
func (c *Command) HasExample() bool {
	return len(c.Example) > 0
}

// Runnable checks whether the command can be executed directly. It returns true if either the Run or RunE method of the Command instance is non-nil, indicating that the command is runnable.
func (c *Command) Runnable() bool {
	return c.Run != nil || c.RunE != nil
}

// HasSubCommands returns true if the command has child commands, otherwise false.
func (c *Command) HasSubCommands() bool {
	return len(c.commands) > 0
}

// IsAvailableCommand determines if a command is available as a non-help command (this includes all non deprecated/hidden commands). It checks if the command has no deprecated or hidden flags, is not the help command of its parent, and either runnable itself or has available subcommands.
func (c *Command) IsAvailableCommand() bool {
	if len(c.Deprecated) != 0 || c.Hidden {
		return false
	}

	if c.HasParent() && c.Parent().helpCommand == c {
		return false
	}

	if c.Runnable() || c.HasAvailableSubCommands() {
		return true
	}

	return false
}

// IsAdditionalHelpTopicCommand determines if a command is an additional help topic command. An additional help topic command is defined as a command that is not runnable, deprecated, hidden, and has no sub-commands that meet these criteria.
// Concrete example: https://github.com/spf13/cobra/issues/393#issuecomment-282741924.
func (c *Command) IsAdditionalHelpTopicCommand() bool {
	// if a command is runnable, deprecated, or hidden it is not a 'help' command
	if c.Runnable() || len(c.Deprecated) != 0 || c.Hidden {
		return false
	}

	// if any non-help sub commands are found, the command is not a 'help' command
	for _, sub := range c.commands {
		if !sub.IsAdditionalHelpTopicCommand() {
			return false
		}
	}

	// the command either has no sub commands, or no non-help sub commands
	return true
}

// HasHelpSubCommands determines if a command has any available 'help' sub commands that need to be shown in the usage/help default template under 'additional help topics'.
// It iterates through all sub-commands of the current command and checks if any of them are marked as additional help topic commands.
// Returns true if at least one such sub-command is found, otherwise returns false.
func (c *Command) HasHelpSubCommands() bool {
	// return true on the first found available 'help' sub command
	for _, sub := range c.commands {
		if sub.IsAdditionalHelpTopicCommand() {
			return true
		}
	}

	// the command either has no sub commands, or no available 'help' sub commands
	return false
}

// HasAvailableSubCommands determines if a command has available subcommands that need to be shown in the usage/help default template under 'available commands'.
// It iterates through the subcommands of the command and returns true if any subcommand is available, considering non-deprecated, non-help, and non-hidden subcommands.
// If there are no subcommands or all subcommands are deprecated, help, or hidden, it returns false.
func (c *Command) HasAvailableSubCommands() bool {
	// return true on the first found available (non deprecated/help/hidden)
	// sub command
	for _, sub := range c.commands {
		if sub.IsAvailableCommand() {
			return true
		}
	}

	// the command either has no sub commands, or no available (non deprecated/help/hidden)
	// sub commands
	return false
}

// HasParent returns true if the command has a parent command, indicating that it is not a root command. Otherwise, it returns false.
func (c *Command) HasParent() bool {
	return c.parent != nil
}

// GlobalNormalizationFunc returns the global normalization function associated with the command, or nil if none is set.
func (c *Command) GlobalNormalizationFunc() func(f *flag.FlagSet, name string) flag.NormalizedName {
	return c.globNormFunc
}

// Flags returns the complete FlagSet that applies to this command (local and persistent declared here and by all parents).
func (c *Command) Flags() *flag.FlagSet {
	if c.flags == nil {
		c.flags = flag.NewFlagSet(c.DisplayName(), flag.ContinueOnError)
		if c.flagErrorBuf == nil {
			c.flagErrorBuf = new(bytes.Buffer)
		}
		c.flags.SetOutput(c.flagErrorBuf)
	}

	return c.flags
}

// LocalNonPersistentFlags returns a FlagSet containing flags specific to this command that will NOT persist to subcommands.
// This function does not modify the flags of the current command; its purpose is to return the current state.
func (c *Command) LocalNonPersistentFlags() *flag.FlagSet {
	persistentFlags := c.PersistentFlags()

	out := flag.NewFlagSet(c.DisplayName(), flag.ContinueOnError)
	c.LocalFlags().VisitAll(func(f *flag.Flag) {
		if persistentFlags.Lookup(f.Name) == nil {
			out.AddFlag(f)
		}
	})
	return out
}

// LocalFlags returns the local FlagSet specifically set in the current command.
//
// This function does not modify the flags of the current command; its purpose is to return the current state.
// It merges persistent flags, initializes the local flag set if necessary, sets output for errors, and adds flags from both the
// regular flags and persistent flags, ensuring that they are not parent PFlags or shadows a parent PFlag.
func (c *Command) LocalFlags() *flag.FlagSet {
	c.mergePersistentFlags()

	if c.lflags == nil {
		c.lflags = flag.NewFlagSet(c.DisplayName(), flag.ContinueOnError)
		if c.flagErrorBuf == nil {
			c.flagErrorBuf = new(bytes.Buffer)
		}
		c.lflags.SetOutput(c.flagErrorBuf)
	}
	c.lflags.SortFlags = c.Flags().SortFlags
	if c.globNormFunc != nil {
		c.lflags.SetNormalizeFunc(c.globNormFunc)
	}

	addToLocal := func(f *flag.Flag) {
		// Add the flag if it is not a parent PFlag, or it shadows a parent PFlag
		if c.lflags.Lookup(f.Name) == nil && f != c.parentsPflags.Lookup(f.Name) {
			c.lflags.AddFlag(f)
		}
	}
	c.Flags().VisitAll(addToLocal)
	c.PersistentFlags().VisitAll(addToLocal)
	return c.lflags
}

// InheritedFlags returns all flags which were inherited from parent commands.
// This function does not modify the flags of the current command; its purpose is to return the current state.
func (c *Command) InheritedFlags() *flag.FlagSet {
	c.mergePersistentFlags()

	if c.iflags == nil {
		c.iflags = flag.NewFlagSet(c.DisplayName(), flag.ContinueOnError)
		if c.flagErrorBuf == nil {
			c.flagErrorBuf = new(bytes.Buffer)
		}
		c.iflags.SetOutput(c.flagErrorBuf)
	}

	local := c.LocalFlags()
	if c.globNormFunc != nil {
		c.iflags.SetNormalizeFunc(c.globNormFunc)
	}

	c.parentsPflags.VisitAll(func(f *flag.Flag) {
		if c.iflags.Lookup(f.Name) == nil && local.Lookup(f.Name) == nil {
			c.iflags.AddFlag(f)
		}
	})
	return c.iflags
}

// Command represents a command in the application.
func (c *Command) NonInheritedFlags() *flag.FlagSet {
	return c.LocalFlags()
}

// PersistentFlags returns the persistent FlagSet specifically set in the current command.
// If the FlagSet has not been initialized, it creates a new one with ContinueOnError as error handling and sets its output to the flagErrorBuf of the command.
func (c *Command) PersistentFlags() *flag.FlagSet {
	if c.pflags == nil {
		c.pflags = flag.NewFlagSet(c.DisplayName(), flag.ContinueOnError)
		if c.flagErrorBuf == nil {
			c.flagErrorBuf = new(bytes.Buffer)
		}
		c.pflags.SetOutput(c.flagErrorBuf)
	}
	return c.pflags
}

// ResetFlags deletes all flags from the command, resetting its internal state.
func (c *Command) ResetFlags() {
	c.flagErrorBuf = new(bytes.Buffer)
	c.flagErrorBuf.Reset()
	c.flags = flag.NewFlagSet(c.DisplayName(), flag.ContinueOnError)
	c.flags.SetOutput(c.flagErrorBuf)
	c.pflags = flag.NewFlagSet(c.DisplayName(), flag.ContinueOnError)
	c.pflags.SetOutput(c.flagErrorBuf)

	c.lflags = nil
	c.iflags = nil
	c.parentsPflags = nil
}

// HasFlags returns true if the command or any of its parent commands contain any flags.
// It checks both local and persistent flags within the command's flag set.
func (c *Command) HasFlags() bool {
	return c.Flags().HasFlags()
}

// HasPersistentFlags returns true if the command contains any persistent flags.
func (c *Command) HasPersistentFlags() bool {
	return c.PersistentFlags().HasFlags()
}

// HasLocalFlags returns true if the command has flags that are specifically declared locally, false otherwise.
func (c *Command) HasLocalFlags() bool {
	return c.LocalFlags().HasFlags()
}

// HasInheritedFlags returns true if the command has flags that are inherited from its parent command.
func (c *Command) HasInheritedFlags() bool {
	return c.InheritedFlags().HasFlags()
}

// HasAvailableFlags returns true if the command contains any flags that are not hidden or deprecated.
func (c *Command) HasAvailableFlags() bool {
	return c.Flags().HasAvailableFlags()
}

// HasAvailablePersistentFlags determines whether the command has any persistent flags that are not marked as hidden or deprecated. Persistent flags are those that are defined in a parent command and can be used by all its subcommands. This method returns true if there is at least one such flag available, otherwise it returns false.
func (c *Command) HasAvailablePersistentFlags() bool {
	return c.PersistentFlags().HasAvailableFlags()
}

// HasAvailableLocalFlags returns true if the command has flags specifically declared locally which are not hidden or deprecated.
func (c *Command) HasAvailableLocalFlags() bool {
	return c.LocalFlags().HasAvailableFlags()
}

// HasAvailableInheritedFlags returns true if the command has flags inherited from its parent command that are
// not hidden or deprecated. It checks the inherited flags using the InheritedFlags method and then filters out
// any flags that are marked as hidden or deprecated before determining if there are available flags.
func (c *Command) HasAvailableInheritedFlags() bool {
	return c.InheritedFlags().HasAvailableFlags()
}

// Flag searches for a flag with the given name within the command and its persistent flags.
//
// It first looks up the flag in the current command's local flags using `Flags().Lookup`.
// If the flag is not found, it then checks the persistent flags using `persistentFlag`.
//
// Parameters:
//   - name: The name of the flag to search for.
//
// Returns:
//   - flag: A pointer to the found flag if it exists; otherwise, nil.
func (c *Command) Flag(name string) (flag *flag.Flag) {
	flag = c.Flags().Lookup(name)

	if flag == nil {
		flag = c.persistentFlag(name)
	}

	return
}

// persistentFlag recursively finds the matching persistent flag.
//
// It first checks if the current command has persistent flags and looks for the flag in them.
// If the flag is not found, it updates the parents' persistent flags and searches again.
//
// Parameters:
// - name: The name of the flag to find.
//
// Returns:
// - flag: A pointer to the matching flag if found; otherwise, nil.
func (c *Command) persistentFlag(name string) (flag *flag.Flag) {
	if c.HasPersistentFlags() {
		flag = c.PersistentFlags().Lookup(name)
	}

	if flag == nil {
		c.updateParentsPflags()
		flag = c.parentsPflags.Lookup(name)
	}
	return
}

// ParseFlags parses persistent flag tree and local flags, merging them into a single set and parsing the provided arguments.
//
// Parameters:
// - args: A slice of strings representing command-line arguments to be parsed.
//
// Returns:
// - error: An error if parsing fails, nil otherwise.
//
// Errors:
// - May return an error from the flag package or custom errors defined in the application.
func (c *Command) ParseFlags(args []string) error {
	if c.DisableFlagParsing {
		return nil
	}

	if c.flagErrorBuf == nil {
		c.flagErrorBuf = new(bytes.Buffer)
	}
	beforeErrorBufLen := c.flagErrorBuf.Len()
	c.mergePersistentFlags()

	// do it here after merging all flags and just before parse
	c.Flags().ParseErrorsWhitelist = flag.ParseErrorsWhitelist(c.FParseErrWhitelist)

	err := c.Flags().Parse(args)
	// Print warnings if they occurred (e.g. deprecated flag messages).
	if c.flagErrorBuf.Len()-beforeErrorBufLen > 0 && err == nil {
		c.Print(c.flagErrorBuf.String())
	}

	return err
}

// Parent returns the parent command of the current command. If there is no parent, it returns nil.
func (c *Command) Parent() *Command {
	return c.parent
}

// Command represents a command in the application.
func (c *Command) mergePersistentFlags() {
	c.updateParentsPflags()
	c.Flags().AddFlagSet(c.PersistentFlags())
	c.Flags().AddFlagSet(c.parentsPflags)
}

// updateParentsPflags updates the flag set of the command and its parents by adding new persistent flags.
// If no parent flags have been initialized, it creates a new flag set with the current command's display name and error output.
func (c *Command) updateParentsPflags() {
	if c.parentsPflags == nil {
		c.parentsPflags = flag.NewFlagSet(c.DisplayName(), flag.ContinueOnError)
		c.parentsPflags.SetOutput(c.flagErrorBuf)
		c.parentsPflags.SortFlags = false
	}

	if c.globNormFunc != nil {
		c.parentsPflags.SetNormalizeFunc(c.globNormFunc)
	}

	c.Root().PersistentFlags().AddFlagSet(flag.CommandLine)

	c.VisitParents(func(parent *Command) {
		c.parentsPflags.AddFlagSet(parent.PersistentFlags())
	})
}

// commandNameMatches checks if two command names are equal, taking into account case sensitivity according to the EnableCaseInsensitive global configuration. It returns true if the command names match, false otherwise.
func commandNameMatches(s string, t string) bool {
	if EnableCaseInsensitive {
		return strings.EqualFold(s, t)
	}

	return s == t
}

// tmplFunc holds a template and a function that will execute said template.
type tmplFunc struct {
	tmpl string
	fn   func(io.Writer, interface{}) error
}

var defaultUsageTemplate = `Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}{{$cmds := .Commands}}{{if eq (len .Groups) 0}}

Available Commands:{{range $cmds}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{else}}{{range $group := .Groups}}

{{.Title}}{{range $cmds}}{{if (and (eq .GroupID $group.ID) (or .IsAvailableCommand (eq .Name "help")))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if not .AllChildCommandsHaveGroup}}

Additional Commands:{{range $cmds}}{{if (and (eq .GroupID "") (or .IsAvailableCommand (eq .Name "help")))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`

// defaultUsageFunc is equivalent to executing defaultUsageTemplate. The two should be changed in sync.
//
// It takes an io.Writer and an interface{} as input, where the interface{} must be a pointer to a Command struct.
// It writes usage information to the io.Writer based on the properties of the Command.
//
// Parameters:
//   - w: An io.Writer to write the output to.
//   - in: An interface{} that should be a pointer to a Command struct.
//
// Returns:
//   - error: If an error occurs during the execution, it returns an error. Otherwise, it returns nil.
func defaultUsageFunc(w io.Writer, in interface{}) error {
	c := in.(*Command)
	fmt.Fprint(w, "Usage:")
	if c.Runnable() {
		fmt.Fprintf(w, "\n  %s", c.UseLine())
	}
	if c.HasAvailableSubCommands() {
		fmt.Fprintf(w, "\n  %s [command]", c.CommandPath())
	}
	if len(c.Aliases) > 0 {
		fmt.Fprintf(w, "\n\nAliases:\n")
		fmt.Fprintf(w, "  %s", c.NameAndAliases())
	}
	if c.HasExample() {
		fmt.Fprintf(w, "\n\nExamples:\n")
		fmt.Fprintf(w, "%s", c.Example)
	}
	if c.HasAvailableSubCommands() {
		cmds := c.Commands()
		if len(c.Groups()) == 0 {
			fmt.Fprintf(w, "\n\nAvailable Commands:")
			for _, subcmd := range cmds {
				if subcmd.IsAvailableCommand() || subcmd.Name() == helpCommandName {
					fmt.Fprintf(w, "\n  %s %s", rpad(subcmd.Name(), subcmd.NamePadding()), subcmd.Short)
				}
			}
		} else {
			for _, group := range c.Groups() {
				fmt.Fprintf(w, "\n\n%s", group.Title)
				for _, subcmd := range cmds {
					if subcmd.GroupID == group.ID && (subcmd.IsAvailableCommand() || subcmd.Name() == helpCommandName) {
						fmt.Fprintf(w, "\n  %s %s", rpad(subcmd.Name(), subcmd.NamePadding()), subcmd.Short)
					}
				}
			}
			if !c.AllChildCommandsHaveGroup() {
				fmt.Fprintf(w, "\n\nAdditional Commands:")
				for _, subcmd := range cmds {
					if subcmd.GroupID == "" && (subcmd.IsAvailableCommand() || subcmd.Name() == helpCommandName) {
						fmt.Fprintf(w, "\n  %s %s", rpad(subcmd.Name(), subcmd.NamePadding()), subcmd.Short)
					}
				}
			}
		}
	}
	if c.HasAvailableLocalFlags() {
		fmt.Fprintf(w, "\n\nFlags:\n")
		fmt.Fprint(w, trimRightSpace(c.LocalFlags().FlagUsages()))
	}
	if c.HasAvailableInheritedFlags() {
		fmt.Fprintf(w, "\n\nGlobal Flags:\n")
		fmt.Fprint(w, trimRightSpace(c.InheritedFlags().FlagUsages()))
	}
	if c.HasHelpSubCommands() {
		fmt.Fprintf(w, "\n\nAdditional help topics:")
		for _, subcmd := range c.Commands() {
			if subcmd.IsAdditionalHelpTopicCommand() {
				fmt.Fprintf(w, "\n  %s %s", rpad(subcmd.CommandPath(), subcmd.CommandPathPadding()), subcmd.Short)
			}
		}
	}
	if c.HasAvailableSubCommands() {
		fmt.Fprintf(w, "\n\nUse \"%s [command] --help\" for more information about a command.", c.CommandPath())
	}
	fmt.Fprintln(w)
	return nil
}

var defaultHelpTemplate = `{{with (or .Long .Short)}}{{. | trimTrailingWhitespaces}}

{{end}}{{if or .Runnable .HasSubCommands}}{{.UsageString}}{{end}}`

// defaultHelpFunc is equivalent to executing defaultHelpTemplate. The two should be changed in sync.
//
// It takes an io.Writer and an interface{} as parameters. The interface{} parameter is expected to be a *Command.
//
// The function retrieves the usage information from the Command, prioritizing Long over Short if both are provided.
// It trims any trailing spaces from the usage text.
//
// If the usage text is not empty, it writes the usage text followed by a newline to the io.Writer.
// If the Command is runnable or has subcommands, it appends the usage string of the Command to the io.Writer.
//
// The function returns nil if successful, or an error if any occurs during execution.
func defaultHelpFunc(w io.Writer, in interface{}) error {
	c := in.(*Command)
	usage := c.Long
	if usage == "" {
		usage = c.Short
	}
	usage = trimRightSpace(usage)
	if usage != "" {
		fmt.Fprintln(w, usage)
		fmt.Fprintln(w)
	}
	if c.Runnable() || c.HasSubCommands() {
		fmt.Fprint(w, c.UsageString())
	}
	return nil
}

var defaultVersionTemplate = `{{with .DisplayName}}{{printf "%s " .}}{{end}}{{printf "version %s" .Version}}
`

// defaultVersionFunc is equivalent to executing defaultVersionTemplate. The two should be changed in sync.
// It writes the version information of a command to the provided writer.
// Parameters:
//   - w: An io.Writer to which the version information will be written.
//   - in: An interface containing a pointer to a Command object from which the display name and version are retrieved.
// Returns:
//   - error: An error if writing to the writer fails, otherwise nil.
func defaultVersionFunc(w io.Writer, in interface{}) error {
	c := in.(*Command)
	_, err := fmt.Fprintf(w, "%s version %s\n", c.DisplayName(), c.Version)
	return err
}
