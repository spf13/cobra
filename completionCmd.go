package cobra

import (
	"fmt"
)

const (
	// Constants for the completion command
	compCmdName              = "completion"
	compCmdNoDescFlagName    = "no-descriptions"
	compCmdNoDescFlagDesc    = "disable completion descriptions"
	compCmdNoDescFlagDefault = false
	shortDesc                = "Generate the autocompletion script for %s"
)

// CompletionOptions are the options to control shell completion
type CompletionOptions struct {
	// DisableDefaultCmd prevents Cobra from creating a default 'completion' command
	DisableDefaultCmd bool
	// DisableNoDescFlag prevents Cobra from creating the '--no-descriptions' flag
	// for shells that support completion descriptions
	DisableNoDescFlag bool
	// DisableDescriptions turns off all completion descriptions for shells
	// that support them
	DisableDescriptions bool
}

var (
	CompletionCmd = &Command{
		Use:               compCmdName,
		Short:             "Generate the autocompletion script for the specified shell",
		Args:              NoArgs,
		ValidArgsFunction: NoFileCompletions,
	}
	BashCompletionCmd = &Command{
		Use:                   "bash",
		Short:                 fmt.Sprintf(shortDesc, "bash"),
		Args:                  NoArgs,
		DisableFlagsInUseLine: true,
		ValidArgsFunction:     NoFileCompletions,
	}
	ZshCompletionCmd = &Command{
		Use:                   "zsh",
		Short:                 fmt.Sprintf(shortDesc, "zsh"),
		Args:                  NoArgs,
		DisableFlagsInUseLine: true,
		ValidArgsFunction:     NoFileCompletions,
	}
	FishCompletionCmd = &Command{
		Use:                   "fish",
		Short:                 fmt.Sprintf(shortDesc, "fish"),
		Args:                  NoArgs,
		DisableFlagsInUseLine: true,
		ValidArgsFunction:     NoFileCompletions,
	}
	PwshCompletionCmd = &Command{
		Use:                   "powershell",
		Short:                 fmt.Sprintf(shortDesc, "powershell"),
		Args:                  NoArgs,
		DisableFlagsInUseLine: true,
		ValidArgsFunction:     NoFileCompletions,
	}
)

// initDefaultCompletionCmd adds a default 'completion' command to c.
// This function will do nothing if any of the following is true:
// 1- the feature has been explicitly disabled by the program,
// 2- c has no subcommands (to avoid creating one),
// 3- c already has a 'completion' command provided by the program.
func (c *Command) initDefaultCompletionCmd() {
	if c.CompletionOptions.DisableDefaultCmd || !c.HasSubCommands() {
		return
	}

	for _, cmd := range c.commands {
		if cmd.Name() == compCmdName || cmd.HasAlias(compCmdName) {
			// A completion command is already available
			return
		}
	}

	if CompletionCmd.Long == "" {
		CompletionCmd.Long = fmt.Sprintf(
			`Generate the autocompletion script for %[1]s for the specified shell.
See each sub-command's help for details on how to use the generated script.`, c.Root().Name())
	}

	c.RemoveCommand(CompletionCmd) // Tests can call this function multiple times in a row, so we must reset
	c.AddCommand(CompletionCmd)

	out := c.OutOrStdout()
	noDesc := c.CompletionOptions.DisableDescriptions
	haveNoDescFlag := !c.CompletionOptions.DisableNoDescFlag && !c.CompletionOptions.DisableDescriptions

	bash := BashCompletionCmd
	if bash.Long == "" {
		bash.Long = fmt.Sprintf(
			`Generate the autocompletion script for the bash shell.

This script depends on the 'bash-completion' package.
If it is not installed already, you can install it via your OS's package manager.

To load completions in your current shell session:
$ source <(%[1]s %[2]s %[3]s)

To load completions for every new session, execute once:
Linux:
$ %[1]s %[2]s %[3]s > /etc/bash_completion.d/%[1]s
MacOS:
$ %[1]s %[2]s %[3]s > /usr/local/etc/bash_completion.d/%[1]s

You will need to start a new shell for this setup to take effect.`,
			c.Root().Name(), CompletionCmd.Name(), BashCompletionCmd.Name())
	}
	bash.RunE = func(cmd *Command, args []string) error {
		return cmd.Root().GenBashCompletionV2(out, !noDesc)
	}

	bash.ResetFlags() // Tests can call this function multiple times in a row, so we must reset
	if haveNoDescFlag {
		bash.Flags().BoolVar(&noDesc, compCmdNoDescFlagName, compCmdNoDescFlagDefault, compCmdNoDescFlagDesc)
	}

	zsh := ZshCompletionCmd
	if zsh.Long == "" {
		zsh.Long = fmt.Sprintf(
			`Generate the autocompletion script for the zsh shell.

If shell completion is not already enabled in your environment you will need
to enable it.  You can execute the following once:

$ echo "autoload -U compinit; compinit" >> ~/.zshrc

To load completions for every new session, execute once:
# Linux:
$ %[1]s %[2]s %[3]s > "${fpath[1]}/_%[1]s"
# macOS:
$ %[1]s %[2]s %[3]s > /usr/local/share/zsh/site-functions/_%[1]s

You will need to start a new shell for this setup to take effect.`,
			c.Root().Name(), CompletionCmd.Name(), ZshCompletionCmd.Name())
	}
	zsh.RunE = func(cmd *Command, args []string) error {
		if noDesc {
			return cmd.Root().GenZshCompletionNoDesc(out)
		}
		return cmd.Root().GenZshCompletion(out)
	}

	zsh.ResetFlags() // Tests can call this function multiple times in a row, so we must reset
	if haveNoDescFlag {
		zsh.Flags().BoolVar(&noDesc, compCmdNoDescFlagName, compCmdNoDescFlagDefault, compCmdNoDescFlagDesc)
	}

	fish := FishCompletionCmd
	if fish.Long == "" {
		fish.Long = fmt.Sprintf(
			`Generate the autocompletion script for the fish shell.

To load completions in your current shell session:
$ %[1]s %[2]s %[3]s | source

To load completions for every new session, execute once:
$ %[1]s %[2]s %[3]s > ~/.config/fish/completions/%[1]s.fish

You will need to start a new shell for this setup to take effect.`,
			c.Root().Name(), CompletionCmd.Name(), FishCompletionCmd.Name())
	}
	fish.RunE = func(cmd *Command, args []string) error {
		return cmd.Root().GenFishCompletion(out, !noDesc)
	}

	fish.ResetFlags() // Tests can call this function multiple times in a row, so we must reset
	if haveNoDescFlag {
		fish.Flags().BoolVar(&noDesc, compCmdNoDescFlagName, compCmdNoDescFlagDefault, compCmdNoDescFlagDesc)
	}

	pwsh := PwshCompletionCmd
	if pwsh.Long == "" {
		pwsh.Long = fmt.Sprintf(
			`Generate the autocompletion script for powershell.

To load completions in your current shell session:
PS C:\> %[1]s %[2]s %[3]s | Out-String | Invoke-Expression

To load completions for every new session, add the output of the above command
to your powershell profile.`,
			c.Root().Name(), CompletionCmd.Name(), PwshCompletionCmd.Name())
	}
	pwsh.RunE = func(cmd *Command, args []string) error {
		if noDesc {
			return cmd.Root().GenPowerShellCompletion(out)
		}
		return cmd.Root().GenPowerShellCompletionWithDesc(out)
	}

	pwsh.ResetFlags() // Tests can call this function multiple times in a row, so we must reset
	if haveNoDescFlag {
		pwsh.Flags().BoolVar(&noDesc, compCmdNoDescFlagName, compCmdNoDescFlagDefault, compCmdNoDescFlagDesc)
	}

	CompletionCmd.RemoveCommand(bash, zsh, fish, pwsh) // Tests can call this function multiple times in a row, so we must reset
	CompletionCmd.AddCommand(bash, zsh, fish, pwsh)
}
