package cobra

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

var _ = fmt.Println

var tp, te, tt, t1 []string
var flagb1, flagb2, flagb3, flagbr bool
var flags1, flags2, flags3 string
var flagi1, flagi2, flagi3, flagir int
var globalFlag1 bool
var flagEcho, rootcalled bool

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
	Use:   "times [# times] [string to echo]",
	Short: "Echo anything to the screen more times",
	Long:  `an slightly useless command for testing.`,
	Run:   timesRunner,
}

var cmdRootNoRun = &Command{
	Use:   "cobra-test",
	Short: "The root can run it's own function",
	Long:  "The root description for help",
}

var cmdRootWithRun = &Command{
	Use:   "cobra-test",
	Short: "The root can run it's own function",
	Long:  "The root description for help",
	Run: func(cmd *Command, args []string) {
		rootcalled = true
	},
}

func timesRunner(cmd *Command, args []string) {
	tt = args
}

func flagInit() {
	cmdEcho.ResetFlags()
	cmdPrint.ResetFlags()
	cmdTimes.ResetFlags()
	cmdRootNoRun.ResetFlags()
	cmdRootWithRun.ResetFlags()
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

func initialize() *Command {
	tt, tp, te = nil, nil, nil
	var c = cmdRootNoRun
	flagInit()
	commandInit()
	return c
}

func initializeWithRootCmd() *Command {
	cmdRootWithRun.ResetCommands()
	tt, tp, te, rootcalled = nil, nil, nil, false
	flagInit()
	cmdRootWithRun.Flags().BoolVarP(&flagbr, "boolroot", "b", false, "help message for flag boolroot")
	cmdRootWithRun.Flags().IntVarP(&flagir, "introot", "i", 321, "help message for flag intthree")
	commandInit()
	return cmdRootWithRun
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
		t.Errorf("default flag value changed on different command with same shortname, 234 expected, %d given", flagi2)
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

	buf := new(bytes.Buffer)
	// Testing with flag that shouldn't be persistent
	c = initialize()
	c.SetOutput(buf)
	// define children
	c.AddCommand(cmdPrint, cmdEcho)
	// define grandchild
	cmdEcho.AddCommand(cmdTimes)
	c.SetArgs(strings.Split("echo times -j 99 -i77 one two", " "))
	e := c.Execute()

	if e == nil {
		t.Errorf("invalid flag should generate error")
	}

	if !strings.Contains(buf.String(), "unknown shorthand") {
		t.Errorf("Wrong error message displayed, \n %s", buf.String())
	}

	if flagi2 != 99 {
		t.Errorf("flag value should be 99, %d given", flagi2)
	}

	if flagi1 != 123 {
		t.Errorf("unset flag should have default value, expecting 123, given %d", flagi1)
	}

	// Testing with flag only existing on child
	buf.Reset()
	c = initialize()
	c.SetOutput(buf)
	cmdEcho.AddCommand(cmdTimes)
	c.AddCommand(cmdPrint, cmdEcho)
	c.SetArgs(strings.Split("echo -j 99 -i77 one two", " "))
	err := c.Execute()

	if err == nil {
		t.Errorf("invalid flag should generate error")
	}

	if !strings.Contains(buf.String(), "intone=123") {
		t.Errorf("Wrong error message displayed, \n %s", buf.String())
	}
}

func TestTrailingCommandFlags(t *testing.T) {
	buf := new(bytes.Buffer)
	c := initialize()
	c.SetOutput(buf)
	cmdEcho.AddCommand(cmdTimes)
	c.AddCommand(cmdPrint, cmdEcho)
	c.SetArgs(strings.Split("echo two -x", " "))
	e3 := c.Execute()

	if e3 == nil {
		t.Errorf("invalid flag should generate error")
	}
}

func TestPersistentFlags(t *testing.T) {
	c := initialize()
	cmdEcho.AddCommand(cmdTimes)
	c.AddCommand(cmdPrint, cmdEcho)
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

	c = initialize()
	cmdEcho.AddCommand(cmdTimes)
	c.AddCommand(cmdPrint, cmdEcho)
	c.SetArgs(strings.Split("echo times -s again -c test here", " "))
	c.Execute()

	if strings.Join(tt, " ") != "test here" {
		t.Errorf("flags didn't leave proper args remaining..%s given", tt)
	}

	if flags1 != "again" {
		t.Errorf("string flag didn't get correct value, had %v", flags1)
	}

	if flagb2 != true {
		t.Errorf("local flag not parsed correctly. Expected false, had %v", flagb2)
	}
}

func TestHelpCommand(t *testing.T) {
	buf := new(bytes.Buffer)
	c := initialize()
	cmdEcho.AddCommand(cmdTimes)
	c.AddCommand(cmdPrint, cmdEcho)
	c.SetArgs(strings.Split("help echo", " "))
	c.SetOutput(buf)
	c.Execute()

	if !strings.Contains(buf.String(), cmdEcho.Long) {
		t.Errorf("Wrong error message displayed, \n %s", buf.String())
	}

	buf.Reset()
	c = initialize()
	cmdEcho.AddCommand(cmdTimes)
	c.AddCommand(cmdPrint, cmdEcho)
	c.SetArgs(strings.Split("help echo times", " "))
	c.SetOutput(buf)
	c.Execute()

	if !strings.Contains(buf.String(), cmdTimes.Long) {
		t.Errorf("Wrong error message displayed, \n %s", buf.String())
	}
}

func TestRunnableRootCommand(t *testing.T) {
	c := initializeWithRootCmd()
	c.AddCommand(cmdPrint, cmdEcho)
	c.SetArgs([]string(nil))
	c.Execute()

	if rootcalled != true {
		t.Errorf("Root Function was not called")
	}
}

func TestRootFlags(t *testing.T) {
	c := initializeWithRootCmd()
	c.AddCommand(cmdPrint, cmdEcho)
	c.SetArgs(strings.Split("-i 17 -b", " "))
	c.Execute()

	if flagbr != true {
		t.Errorf("flag value should be true, %v given", flagbr)
	}

	if flagir != 17 {
		t.Errorf("flag value should be 17, %d given", flagir)
	}

}
