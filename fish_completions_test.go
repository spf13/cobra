package cobra

import (
	"bytes"
	"testing"
)

func TestFishCompletions(t *testing.T) {
	rootCmd := &Command{
		Use:        "root",
		ArgAliases: []string{"pods", "nodes", "services", "replicationcontrollers", "po", "no", "svc", "rc"},
		ValidArgs:  []string{"pod", "node", "service", "replicationcontroller"},
		Run:        emptyRun,
	}
	rootCmd.Flags().IntP("introot", "i", -1, "help's message for flag introot")
	rootCmd.MarkFlagRequired("introot")

	// Filename.
	rootCmd.Flags().String("filename", "", "Enter a filename")
	rootCmd.MarkFlagFilename("filename", "json", "yaml", "yml")

	// Persistent filename.
	rootCmd.PersistentFlags().String("persistent-filename", "", "Enter a filename")
	rootCmd.MarkPersistentFlagFilename("persistent-filename")
	rootCmd.MarkPersistentFlagRequired("persistent-filename")

	// Filename extensions.
	rootCmd.Flags().String("filename-ext", "", "Enter a filename (extension limited)")
	rootCmd.MarkFlagFilename("filename-ext")
	rootCmd.Flags().String("custom", "", "Enter a filename (extension limited)")
	rootCmd.MarkFlagCustom("custom", "__complete_custom")

	// Subdirectories in a given directory.
	rootCmd.Flags().String("theme", "", "theme to use (located in /themes/THEMENAME/)")

	echoCmd := &Command{
		Use:     "echo [string to echo]",
		Aliases: []string{"say"},
		Short:   "Echo anything's to the screen",
		Long:    "an utterly useless command for testing.",
		Example: "Just run cobra-test echo",
		Run:     emptyRun,
	}

	echoCmd.Flags().String("filename", "", "Enter a filename")
	echoCmd.MarkFlagFilename("filename", "json", "yaml", "yml")
	echoCmd.Flags().String("config", "", "config to use (located in /config/PROFILE/)")

	printCmd := &Command{
		Use:   "print [string to print]",
		Args:  MinimumNArgs(1),
		Short: "Print anything to the screen",
		Long:  "an absolutely utterly useless command for testing.",
		Run:   emptyRun,
	}

	deprecatedCmd := &Command{
		Use:        "deprecated [can't do anything here]",
		Args:       NoArgs,
		Short:      "A command which is deprecated",
		Long:       "an absolutely utterly useless command for testing deprecation!.",
		Deprecated: "Please use echo instead",
		Run:        emptyRun,
	}

	colonCmd := &Command{
		Use: "cmd:colon",
		Run: emptyRun,
	}

	timesCmd := &Command{
		Use:        "times [# times] [string to echo]",
		SuggestFor: []string{"counts"},
		Args:       OnlyValidArgs,
		ValidArgs:  []string{"one", "two", "three", "four"},
		Short:      "Echo anything to the screen more times",
		Long:       "a slightly useless command for testing.",
		Run:        emptyRun,
	}

	echoCmd.AddCommand(timesCmd)
	rootCmd.AddCommand(echoCmd, printCmd, deprecatedCmd, colonCmd)

	buf := new(bytes.Buffer)
	rootCmd.GenFishCompletion(buf)
	output := buf.String()

	// check for preamble helper functions
	check(t, output, "__fish_root_no_subcommand")
	check(t, output, "__fish_root_seen_subcommand_path")
	check(t, output, "__fish_seen_argument")

	// check for subcommands
	check(t, output, "-a echo")
	check(t, output, "-a print")
	checkOmit(t, output, "-a deprecated")
	check(t, output, "-a cmd:colon")

	// check for nested subcommands
	checkRegex(t, output, `-n '__fish_root_seen_subcommand_path echo(; and[^']*)?' -a times`)

	// check for flags that will take arguments
	check(t, output, "-n '__fish_root_no_subcommand' -r -s i -l introot")
	check(t, output, "-n '__fish_root_no_subcommand' -r  -l filename")
	check(t, output, "-n '__fish_root_no_subcommand' -r  -l persistent-filename")
	check(t, output, "-n '__fish_root_no_subcommand' -r  -l theme")
	check(t, output, "-n '__fish_root_seen_subcommand_path echo' -r  -l config")
	check(t, output, "-n '__fish_root_seen_subcommand_path echo' -r  -l filename")

	// checks escape of description in flags
	check(t, output, "-n '__fish_root_no_subcommand' -r -s i -l introot -d 'help\\'s message for flag introot'")

	// check for persistent flags that will take arguments
	check(t, output, "-n '__fish_root_seen_subcommand_path cmd:colon' -r  -l persistent-filename")
	check(t, output, "-n '__fish_root_seen_subcommand_path echo' -r  -l persistent-filename")
	check(t, output, "-n '__fish_root_seen_subcommand_path echo times' -r  -l persistent-filename")
	check(t, output, "-n '__fish_root_seen_subcommand_path print' -r  -l persistent-filename")

	// check for local non-persistent flags
	checkRegex(t, output, `; and not __fish_seen_argument -l custom[^']*' -a echo`)
	checkRegex(t, output, `; and not __fish_seen_argument -l filename[^']*' -a echo`)
	checkRegex(t, output, `; and not __fish_seen_argument -l filename-ext[^']*' -a echo`)
	checkRegex(t, output, `; and not __fish_seen_argument -s i -l introot[^']*' -a echo`)
	checkRegex(t, output, `; and not __fish_seen_argument -l theme[^']*' -a echo`)

	// check for positional arguments to a command
	checkRegex(t, output, `-n '__fish_root_no_subcommand(; and[^']*)?' -a pod`)
	checkRegex(t, output, `-n '__fish_root_no_subcommand(; and[^']*)?' -a node`)
	checkRegex(t, output, `-n '__fish_root_no_subcommand(; and[^']*)?' -a service`)
	checkRegex(t, output, `-n '__fish_root_no_subcommand(; and[^']*)?' -a replicationcontroller`)

	// check for aliases to positional arguments for a command
	checkRegex(t, output, `-n '__fish_root_no_subcommand(; and[^']*)?' -a pods`)
	checkRegex(t, output, `-n '__fish_root_no_subcommand(; and[^']*)?' -a nodes`)
	checkRegex(t, output, `-n '__fish_root_no_subcommand(; and[^']*)?' -a services`)
	checkRegex(t, output, `-n '__fish_root_no_subcommand(; and[^']*)?' -a replicationcontrollers`)
	checkRegex(t, output, `-n '__fish_root_no_subcommand(; and[^']*)?' -a po`)
	checkRegex(t, output, `-n '__fish_root_no_subcommand(; and[^']*)?' -a no`)
	checkRegex(t, output, `-n '__fish_root_no_subcommand(; and[^']*)?' -a svc`)
	checkRegex(t, output, `-n '__fish_root_no_subcommand(; and[^']*)?' -a rc`)
}
