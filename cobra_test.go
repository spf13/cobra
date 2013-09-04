package cobra_test

import (
	. "cobra"
	"strings"
	"testing"
)

var tp, te, tt, t1 []string
var flagb1, flagb2, flagb3 bool
var flags1, flags2, flags3 string
var flagi1, flagi2, flagi3 int
var globalFlag1 bool
var flagEcho bool

var cmdPrint = &Command{
	Use:   "print [string to print]",
	Short: "Print anything to the screen",
	Long:  `an utterly useless command for testing.`,
	Run: func(cmd *Command, args []string) {
		tp = args
	},
}

var cmdEcho = &Command{
	Use:   "echo [string to echo]",
	Short: "Echo anything to the screen",
	Long:  `an utterly useless command for testing.`,
	Run: func(cmd *Command, args []string) {
		te = args
	},
}

var cmdTimes = &Command{
	Use:   "times [string to echo]",
	Short: "Echo anything to the screen more times",
	Long:  `an slightly useless command for testing.`,
	Run:   timesRunner,
}

func timesRunner(cmd *Command, args []string) {
	tt = args
}

func flagInit() {
	cmdEcho.ResetFlags()
	cmdPrint.ResetFlags()
	cmdTimes.ResetFlags()
	cmdEcho.Flags().IntVarP(&flagi1, "intone", "i", 123, "help message for flag intone")
	cmdTimes.Flags().IntVarP(&flagi2, "inttwo", "j", 234, "help message for flag inttwo")
	cmdPrint.Flags().IntVarP(&flagi3, "intthree", "i", 345, "help message for flag intthree")
	cmdEcho.PersistentFlags().StringVarP(&flags1, "strone", "s", "one", "help message for flag strone")
	cmdTimes.PersistentFlags().StringVarP(&flags2, "strtwo", "t", "two", "help message for flag strtwo")
	cmdPrint.PersistentFlags().StringVarP(&flags3, "strthree", "s", "three", "help message for flag strthree")
	cmdEcho.Flags().BoolVarP(&flagb1, "boolone", "b", true, "help message for flag boolone")
	cmdTimes.Flags().BoolVarP(&flagb2, "booltwo", "c", false, "help message for flag booltwo")
	cmdPrint.Flags().BoolVarP(&flagb3, "boolthree", "b", true, "help message for flag boolthree")
}

func commandInit() {
	cmdEcho.ResetCommands()
	cmdPrint.ResetCommands()
	cmdTimes.ResetCommands()
}

func initialize() *Commander {
	tt, tp, te = nil, nil, nil
	var c = NewCommander()
	c.SetName("cobra test")
	flagInit()
	commandInit()
	return c
}

func TestSingleCommand(t *testing.T) {
	c := initialize()
	c.AddCommand(cmdPrint, cmdEcho)
	c.SetArgs(strings.Split("print one two", " "))
	c.Execute()

	if te != nil || tt != nil {
		t.Error("Wrong command called")
	}
	if tp == nil {
		t.Error("Wrong command called")
	}
	if strings.Join(tp, " ") != "one two" {
		t.Error("Command didn't parse correctly")
	}
}

func TestChildCommand(t *testing.T) {
	c := initialize()
	cmdEcho.AddCommand(cmdTimes)
	c.AddCommand(cmdPrint, cmdEcho)
	c.SetArgs(strings.Split("echo times one two", " "))
	c.Execute()

	if te != nil || tp != nil {
		t.Error("Wrong command called")
	}
	if tt == nil {
		t.Error("Wrong command called")
	}
	if strings.Join(tt, " ") != "one two" {
		t.Error("Command didn't parse correctly")
	}
}

func TestFlagLong(t *testing.T) {
	c := initialize()
	c.AddCommand(cmdPrint, cmdEcho, cmdTimes)
	c.SetArgs(strings.Split("echo --intone=13 something here", " "))
	c.Execute()

	if strings.Join(te, " ") != "something here" {
		t.Errorf("flags didn't leave proper args remaining..%s given", te)
	}
	if flagi1 != 13 {
		t.Errorf("int flag didn't get correct value, had %d", flagi1)
	}
	if flagi2 != 234 {
		t.Errorf("default flag value changed, 234 expected, %d given", flagi2)
	}
}

func TestFlagShort(t *testing.T) {
	c := initialize()
	c.AddCommand(cmdPrint, cmdEcho, cmdTimes)
	c.SetArgs(strings.Split("echo -i13 something here", " "))
	c.Execute()

	if strings.Join(te, " ") != "something here" {
		t.Errorf("flags didn't leave proper args remaining..%s given", te)
	}
	if flagi1 != 13 {
		t.Errorf("int flag didn't get correct value, had %d", flagi1)
	}
	if flagi2 != 234 {
		t.Errorf("default flag value changed, 234 expected, %d given", flagi2)
	}

	c = initialize()
	c.AddCommand(cmdPrint, cmdEcho, cmdTimes)
	c.SetArgs(strings.Split("echo -i 13 something here", " "))
	c.Execute()

	if strings.Join(te, " ") != "something here" {
		t.Errorf("flags didn't leave proper args remaining..%s given", te)
	}
	if flagi1 != 13 {
		t.Errorf("int flag didn't get correct value, had %d", flagi1)
	}
	if flagi2 != 234 {
		t.Errorf("default flag value changed, 234 expected, %d given", flagi2)
	}

	// Testing same shortcode, different command
	c = initialize()
	c.AddCommand(cmdPrint, cmdEcho, cmdTimes)
	c.SetArgs(strings.Split("print -i99 one two", " "))
	c.Execute()

	if strings.Join(tp, " ") != "one two" {
		t.Errorf("flags didn't leave proper args remaining..%s given", tp)
	}
	if flagi3 != 99 {
		t.Errorf("int flag didn't get correct value, had %d", flagi3)
	}
	if flagi1 != 123 {
		t.Errorf("default flag value changed on different comamnd with same shortname, 234 expected, %d given", flagi2)
	}
}

func TestChildCommandFlags(t *testing.T) {
	c := initialize()
	cmdEcho.AddCommand(cmdTimes)
	c.AddCommand(cmdPrint, cmdEcho)
	c.SetArgs(strings.Split("echo times -j 99 one two", " "))
	c.Execute()

	if strings.Join(tt, " ") != "one two" {
		t.Errorf("flags didn't leave proper args remaining..%s given", tt)
	}

	//c = initialize()
	//cmdEcho.AddCommand(cmdTimes)
	//c.AddCommand(cmdPrint, cmdEcho)
	//c.SetArgs(strings.Split("echo times -j 99 -i 77 one two", " "))
	//c.Execute()

	//if strings.Join(tt, " ") != "one two" {
	//t.Errorf("flags didn't leave proper args remaining..%s given", tt)
	//}
}

func TestPersistentFlags(t *testing.T) {
	c := initialize()
	cmdEcho.AddCommand(cmdTimes)
	c.AddCommand(cmdPrint, cmdEcho)
	flagInit()
	c.SetArgs(strings.Split("echo -s something more here", " "))
	c.Execute()

	// persistentFlag should act like normal flag on it's own command
	if strings.Join(te, " ") != "more here" {
		t.Errorf("flags didn't leave proper args remaining..%s given", te)
	}

	// persistentFlag should act like normal flag on it's own command
	if flags1 != "something" {
		t.Errorf("string flag didn't get correct value, had %v", flags1)
	}

}
