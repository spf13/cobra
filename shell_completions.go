package cobra

import (
	"fmt"

	"github.com/spf13/pflag"
)

// MarkFlagRequired adds the BashCompOneRequiredFlag annotation to the named flag if it exists,
// and causes your command to report an error if invoked without the flag.
func (c *Command) MarkFlagRequired(name string) error {
	return MarkFlagRequired(c.Flags(), name)
}

// MarkPersistentFlagRequired adds the BashCompOneRequiredFlag annotation to the named persistent flag if it exists,
// and causes your command to report an error if invoked without the flag.
func (c *Command) MarkPersistentFlagRequired(name string) error {
	return MarkFlagRequired(c.PersistentFlags(), name)
}

// MarkFlagRequired adds the BashCompOneRequiredFlag annotation to the named flag if it exists,
// and causes your command to report an error if invoked without the flag.
func MarkFlagRequired(flags *pflag.FlagSet, name string) error {
	return flags.SetAnnotation(name, BashCompOneRequiredFlag, []string{"true"})
}

// MarkFlagFilename adds the BashCompFilenameExt annotation to the named flag, if it exists.
// Generated bash autocompletion will select filenames for the flag, limiting to named extensions if provided.
func (c *Command) MarkFlagFilename(name string, extensions ...string) error {
	return MarkFlagFilename(c.Flags(), name, extensions...)
}

// MarkFlagCustom adds the BashCompCustom annotation to the named flag, if it exists.
// Generated bash autocompletion will call the bash function f for the flag.
func (c *Command) MarkFlagCustom(name string, f string) error {
	return MarkFlagCustom(c.Flags(), name, f)
}

// MarkPersistentFlagFilename instructs the various shell completion
// implementations to limit completions for this persistent flag to the
// specified extensions (patterns).
//
// Shell Completion compatibility matrix: bash, zsh
func (c *Command) MarkPersistentFlagFilename(name string, extensions ...string) error {
	return MarkFlagFilename(c.PersistentFlags(), name, extensions...)
}

// MarkFlagFilename instructs the various shell completion implementations to
// limit completions for this flag to the specified extensions (patterns).
//
// Shell Completion compatibility matrix: bash, zsh
func MarkFlagFilename(flags *pflag.FlagSet, name string, extensions ...string) error {
	return flags.SetAnnotation(name, BashCompFilenameExt, extensions)
}

// MarkFlagCustom instructs the various shell completion implementations to
// limit completions for this flag to the specified extensions (patterns).
//
// Shell Completion compatibility matrix: bash, zsh
func MarkFlagCustom(flags *pflag.FlagSet, name string, f string) error {
	return flags.SetAnnotation(name, BashCompCustom, []string{f})
}

// MarkFlagDirname instructs the various shell completion implementations to
// complete only directories with this named flag.
//
// Shell Completion compatibility matrix: zsh
func (c *Command) MarkFlagDirname(name string) error {
	return MarkFlagDirname(c.Flags(), name)
}

// MarkPersistentFlagDirname instructs the various shell completion
// implementations to complete only directories with this persistent named flag.
//
// Shell Completion compatibility matrix: zsh
func (c *Command) MarkPersistentFlagDirname(name string) error {
	return MarkFlagDirname(c.PersistentFlags(), name)
}

// MarkFlagDirname instructs the various shell completion implementations to
// complete only directories with this specified flag.
//
// Shell Completion compatibility matrix: zsh
func MarkFlagDirname(flags *pflag.FlagSet, name string) error {
	zshPattern := "-(/)"
	return flags.SetAnnotation(name, zshCompDirname, []string{zshPattern})
}

type DynamicFlagCompletion func(currentValue string) (suggestedValues []string, err error)

// MarkDynamicFlagCompletion provides cobra a function to dynamically suggest values to the user during tab completion
// for this flag. All (Persistent)PreRun(E) functions will be run accordingly before the provided function is called if
// RunPreRunsDuringCompletion is set to true. All flags other than the flag currently being completed will be parsed
// according to their type. The flag being completed is parsed as a raw string with no format requirements
//
// Shell Completion compatibility matrix: zsh
func (c *Command) MarkDynamicFlagCompletion(name string, completion DynamicFlagCompletion) error {
	flag := c.Flag(name)
	if flag == nil {
		return fmt.Errorf("no such flag %s", name)
	}
	if flag.NoOptDefVal != "" {
		return fmt.Errorf("%s takes no parameters", name)
	}

	if c.flagCompletions == nil {
		c.flagCompletions = make(map[*pflag.Flag]DynamicFlagCompletion)
	}
	c.flagCompletions[flag] = completion
	if flag.Annotations == nil {
		flag.Annotations = map[string][]string{}
	}
	flag.Annotations[zshCompDynamicCompletion] = []string{zshCompGenFlagCompletionFuncName(c)}
	return nil
}
