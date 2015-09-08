package cobra

import (
	"reflect"
	"testing"
)

// test to ensure hidden flags run as intended
func TestHiddenFlagExecutes(t *testing.T) {
	var out string
	var secretFlag bool

	boringCmd := &Command{
		Use:   "boring",
		Short: "Do something boring...",
		Long:  `Not a flashy command, just does boring stuff`,
		Run: func(cmd *Command, args []string) {
			out = "boring output"

			if secretFlag {
				out = "super secret NOT boring output!"
			}
		},
	}

	//
	boringCmd.Flags().BoolVarP(&secretFlag, "secret", "s", false, "makes commands run in super secret mode")
	boringCmd.Flags().MarkHidden("secret")

	//
	boringCmd.Execute()

	if out != "boring output" {
		t.Errorf("Command with hidden flag failed to run!")
	}

	//
	boringCmd.execute([]string{"-s"})

	if out != "super secret NOT boring output!" {
		t.Errorf("Hidden flag failed to run!")
	}
}

// test to ensure hidden flags do not show up in usage/help text
func TestHiddenFlagsAreHidden(t *testing.T) {
	var out string
	var secretFlag bool
	var persistentSecretFlag bool

	boringCmd := &Command{
		Use:   "boring",
		Short: "Do something boring...",
		Long:  `Not a flashy command, just does boring stuff`,
		Run: func(cmd *Command, args []string) {
			out = "boring output"

			if secretFlag {
				out = "super secret NOT boring output!"
			}

			if persistentSecretFlag {
				out = "you have no idea what you're getting yourself into!"
			}
		},
	}

	//
	boringCmd.Flags().BoolVarP(&secretFlag, "secret", "s", false, "run this command in super secret mode")
	boringCmd.Flags().MarkHidden("secret")

	// if a command has local flags, they will appear in usage/help text
	if boringCmd.HasLocalFlags() {
		t.Errorf("Hidden flag found!")
	}

	//
	boringCmd.PersistentFlags().BoolVarP(&persistentSecretFlag, "Secret", "S", false, "run any sub command in super secret mode")
	boringCmd.Flags().MarkHidden("Secret")

	// if a command has inherited flags, they will appear in usage/help text
	if boringCmd.HasInheritedFlags() {
		t.Errorf("Hidden flag found!")
	}
}

func TestStripFlags(t *testing.T) {
	tests := []struct {
		input  []string
		output []string
	}{
		{
			[]string{"foo", "bar"},
			[]string{"foo", "bar"},
		},
		{
			[]string{"foo", "--bar", "-b"},
			[]string{"foo"},
		},
		{
			[]string{"-b", "foo", "--bar", "bar"},
			[]string{},
		},
		{
			[]string{"-i10", "echo"},
			[]string{"echo"},
		},
		{
			[]string{"-i=10", "echo"},
			[]string{"echo"},
		},
		{
			[]string{"--int=100", "echo"},
			[]string{"echo"},
		},
		{
			[]string{"-ib", "echo", "-bfoo", "baz"},
			[]string{"echo", "baz"},
		},
		{
			[]string{"-i=baz", "bar", "-i", "foo", "blah"},
			[]string{"bar", "blah"},
		},
		{
			[]string{"--int=baz", "-bbar", "-i", "foo", "blah"},
			[]string{"blah"},
		},
		{
			[]string{"--cat", "bar", "-i", "foo", "blah"},
			[]string{"bar", "blah"},
		},
		{
			[]string{"-c", "bar", "-i", "foo", "blah"},
			[]string{"bar", "blah"},
		},
		{
			[]string{"--persist", "bar"},
			[]string{"bar"},
		},
		{
			[]string{"-p", "bar"},
			[]string{"bar"},
		},
	}

	cmdPrint := &Command{
		Use:   "print [string to print]",
		Short: "Print anything to the screen",
		Long:  `an utterly useless command for testing.`,
		Run: func(cmd *Command, args []string) {
			tp = args
		},
	}

	var flagi int
	var flagstr string
	var flagbool bool
	cmdPrint.PersistentFlags().BoolVarP(&flagbool, "persist", "p", false, "help for persistent one")
	cmdPrint.Flags().IntVarP(&flagi, "int", "i", 345, "help message for flag int")
	cmdPrint.Flags().StringVarP(&flagstr, "bar", "b", "bar", "help message for flag string")
	cmdPrint.Flags().BoolVarP(&flagbool, "cat", "c", false, "help message for flag bool")

	for _, test := range tests {
		output := stripFlags(test.input, cmdPrint)
		if !reflect.DeepEqual(test.output, output) {
			t.Errorf("expected: %v, got: %v", test.output, output)
		}
	}
}
