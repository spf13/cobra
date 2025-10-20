package cobra_test

import (
	"bufio"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestRunInteractive_NoSubCmds(t *testing.T) {
	rootCmd := &cobra.Command{
		Use:   "main",
		Short: "first level",
		Run: func(cmd *cobra.Command, args []string) {
			_, err := cobra.RunInteractive(cmd)
			if err == nil {
				t.Errorf("expected error, because root cmd doesn't contain sub commands")
			}
		},
	}
	rootCmd.Execute()
}

func TestRunInteractive_SubCmdChain(t *testing.T) {
	tests := []struct {
		rootWasCalled  bool
		oneWasCalled   bool
		twoWasCalled   bool
		threeWasCalled bool
		commandToCall  string
	}{
		{
			rootWasCalled:  true,
			oneWasCalled:   true,
			twoWasCalled:   false,
			threeWasCalled: false,
			commandToCall:  "one",
		},
		{
			rootWasCalled:  true,
			oneWasCalled:   false,
			twoWasCalled:   true,
			threeWasCalled: false,
			commandToCall:  "two",
		},
		{
			rootWasCalled:  true,
			oneWasCalled:   false,
			twoWasCalled:   false,
			threeWasCalled: true,
			commandToCall:  "three",
		},
	}
	origSelectionFactory := cobra.SelectionFactory
	defer func() {
		cobra.SelectionFactory = origSelectionFactory
	}()

	for _, test := range tests {
		rootWasCalled := false
		oneWasCalled := false
		twoWasCalled := false
		threeWasCalled := false

		*cobra.SelectionFactory = func(options []string) (string, error) {
			return test.commandToCall, nil
		}

		rootCmd := &cobra.Command{
			Use:   "main",
			Short: "first level",
			Run: func(cmd *cobra.Command, args []string) {
				rootWasCalled = true
				nextCmd, err := cobra.RunInteractive(cmd)
				if err != nil {
					t.Errorf("expected no error, got: %v", err)
				}
				nextCmd.Run(nextCmd, args)
			},
		}
		oneCmd := &cobra.Command{
			Use:   "one",
			Short: "second level",
			Run: func(cmd *cobra.Command, args []string) {
				oneWasCalled = true
			},
		}
		twoCmd := &cobra.Command{
			Use:   "two",
			Short: "second level",
			Run: func(cmd *cobra.Command, args []string) {
				twoWasCalled = true
			},
		}
		threeCmd := &cobra.Command{
			Use:   "three",
			Short: "second level",
			Run: func(cmd *cobra.Command, args []string) {
				threeWasCalled = true
			},
		}
		rootCmd.AddCommand(oneCmd)
		rootCmd.AddCommand(twoCmd)
		rootCmd.AddCommand(threeCmd)

		rootCmd.Run(rootCmd, []string{})

		if rootWasCalled != test.rootWasCalled {
			t.Errorf("%s - unexpected root cmd call: expected: %v, got: %v", test.commandToCall, test.rootWasCalled, rootWasCalled)
		}
		if oneWasCalled != test.oneWasCalled {
			t.Errorf("%s - unexpected one cmd call: expected: %v, got: %v", test.commandToCall, test.oneWasCalled, oneWasCalled)
		}
		if twoWasCalled != test.twoWasCalled {
			t.Errorf("%s - unexpected two cmd call: expected: %v, got: %v", test.commandToCall, test.twoWasCalled, twoWasCalled)
		}
		if threeWasCalled != test.threeWasCalled {
			t.Errorf("%s - unexpected three cmd call: expected: %v, got: %v", test.commandToCall, test.threeWasCalled, threeWasCalled)
		}
	}
}

type TestData struct {
	rootWasCalled bool
	oneWasCalled  bool
	twoWasCalled  bool
	commandToCall string
	hasFlags      bool
	flagsInput    string
	flagVars      struct {
		pBool        bool
		pInt         int
		pString      string
		pStringArray []string
		fBool        bool
		fInt         int
		fString      string
		fStringArray []string
	}
	initFlagsFunc func(testData *TestData, cmd *cobra.Command)
	checkFunc     func(testData *TestData, t *testing.T) bool
}

func noFlags(testData *TestData, t *testing.T) bool {
	ret := true
	if testData.flagVars.pBool != false {
		ret = false
		t.Error("seems pBool was set")
	}
	if testData.flagVars.pInt != 0 {
		ret = false
		t.Error("seems pInt was set")
	}
	if testData.flagVars.pString != "" {
		ret = false
		t.Error("seems pString was set")

	}
	if len(testData.flagVars.pStringArray) != 0 {
		ret = false
		t.Error("seems pStringArray was set")
	}
	if testData.flagVars.fBool != false {
		ret = false
		t.Error("seems fBool was set")
	}
	if testData.flagVars.fInt != 0 {
		ret = false
		t.Error("seems fInt was set")
	}
	if testData.flagVars.fString != "" {
		ret = false
		t.Error("seems fString was set")
	}
	if len(testData.flagVars.fStringArray) != 0 {
		ret = false
		t.Error("seems fStringArray was set")
	}
	return ret
}

func TestRunInteractive_InitFlags(t *testing.T) {
	tests := []TestData{
		// {
		// 	rootWasCalled: true,
		// 	oneWasCalled:  true,
		// 	twoWasCalled:  false,
		// 	commandToCall: "one",
		// 	hasFlags:      false,
		// 	flagsInput:    "",
		// 	checkFunc:     noFlags,
		// },
		// {
		// 	rootWasCalled: true,
		// 	oneWasCalled:  false,
		// 	twoWasCalled:  true,
		// 	commandToCall: "two",
		// 	hasFlags:      true,
		// 	flagsInput:    "\n\n\n\n\n\n\n\n",

		// 	initFlagsFunc: func(testData *TestData, cmd *cobra.Command) {
		// 		rootCmd := cmd.Parent()
		// 		rootCmd.PersistentFlags().BoolVar(&testData.flagVars.pBool, "a", false, "i am a bool var")
		// 		rootCmd.PersistentFlags().StringVar(&testData.flagVars.pString, "b", "", "i am a string var")
		// 		rootCmd.PersistentFlags().IntVar(&testData.flagVars.pInt, "c", 0, "i am an int var")
		// 		rootCmd.PersistentFlags().StringArrayVar(&testData.flagVars.pStringArray, "d", []string{}, "i am an string array var")

		// 		cmd.Flags().BoolVar(&testData.flagVars.fBool, "e", false, "i am a bool var")
		// 		cmd.Flags().StringVar(&testData.flagVars.fString, "f", "", "i am a string var")
		// 		cmd.Flags().IntVar(&testData.flagVars.fInt, "g", 0, "i am an int var")
		// 		cmd.Flags().StringArrayVar(&testData.flagVars.fStringArray, "h", []string{}, "i am an string array var")
		// 	},
		// 	checkFunc: noFlags,
		// },
		{
			rootWasCalled: true,
			oneWasCalled:  false,
			twoWasCalled:  true,
			commandToCall: "two",
			hasFlags:      true,
			// the flags are queried in the order, that first the normal flags are queried and then persistence flags
			flagsInput: "false\ntest content\n540\nanother value\n\ntrue\nI am ä sentence\n666\nää üüü 00\n\n",

			initFlagsFunc: func(testData *TestData, cmd *cobra.Command) {
				rootCmd := cmd.Parent()
				rootCmd.PersistentFlags().BoolVar(&testData.flagVars.pBool, "a", false, "i am a persistent bool var")
				rootCmd.PersistentFlags().StringVar(&testData.flagVars.pString, "b", "", "i am a persistent string var")
				rootCmd.PersistentFlags().IntVar(&testData.flagVars.pInt, "c", 0, "i am an persistent int var")
				rootCmd.PersistentFlags().StringArrayVar(&testData.flagVars.pStringArray, "d", []string{}, "i am an persistent string array var")

				cmd.Flags().BoolVar(&testData.flagVars.fBool, "e", false, "i am a bool var")
				cmd.Flags().StringVar(&testData.flagVars.fString, "f", "", "i am a string var")
				cmd.Flags().IntVar(&testData.flagVars.fInt, "g", 0, "i am an int var")
				cmd.Flags().StringArrayVar(&testData.flagVars.fStringArray, "h", []string{}, "i am an string array var")
			},
			checkFunc: func(testData *TestData, t *testing.T) bool {
				ret := true
				if !testData.flagVars.pBool {
					ret = false
					t.Error("seems pBool was not set")
				}
				if testData.flagVars.pInt != 666 {
					ret = false
					t.Errorf("seems pInt has wrong value: %v", testData.flagVars.pInt)
				}
				if testData.flagVars.pString != "I am ä sentence" {
					ret = false
					t.Errorf("seems pString has wrong value: %v", testData.flagVars.pString)
				}
				if len(testData.flagVars.pStringArray) != 1 {
					ret = false
					t.Error("seems pStringArray wasn't set")
				} else if testData.flagVars.pStringArray[0] != "ää üüü 00" {
					ret = false
					t.Error("seems pStringArray has wrong content: " + testData.flagVars.pStringArray[0])
				}
				if testData.flagVars.fBool != false {
					ret = false
					t.Error("seems fBool was set")
				}
				if testData.flagVars.fInt != 540 {
					ret = false
					t.Errorf("seems fInt has wrong content: %v", testData.flagVars.fInt)
				}
				if testData.flagVars.fString != "test content" {
					ret = false
					t.Errorf("seems fString has wrong value: %v", testData.flagVars.fString)
				}
				if len(testData.flagVars.fStringArray) != 1 {
					ret = false
					t.Error("seems fStringArray wasn't set")
				} else if testData.flagVars.fStringArray[0] != "another value" {
					ret = false
					t.Error("seems fStringArray has wrong content: " + testData.flagVars.fStringArray[0])
				}
				return ret
			},
		},
	}
	origSelectionFactory := cobra.SelectionFactory
	origReaderFactory := cobra.ReaderFactory
	defer func() {
		cobra.SelectionFactory = origSelectionFactory
		cobra.ReaderFactory = origReaderFactory
	}()

	for _, test := range tests {
		rootWasCalled := false
		oneWasCalled := false
		twoWasCalled := false

		*cobra.SelectionFactory = func(options []string) (string, error) {
			return test.commandToCall, nil
		}
		*cobra.ReaderFactory = func() *bufio.Reader {
			r := strings.NewReader(test.flagsInput)
			return bufio.NewReader(r)
		}

		rootCmd := &cobra.Command{
			Use:   "main",
			Short: "first level",
			Run: func(cmd *cobra.Command, args []string) {
				rootWasCalled = true
				nextCmd, err := cobra.RunInteractive(cmd)
				if err != nil {
					t.Errorf("expected no error, got: %v", err)
				}
				nextCmd.Run(nextCmd, args)
			},
		}
		oneCmd := &cobra.Command{
			Use:   "one",
			Short: "second level",
			Run: func(cmd *cobra.Command, args []string) {
				oneWasCalled = true
			},
		}
		twoCmd := &cobra.Command{
			Use:   "two",
			Short: "second level",
			Run: func(cmd *cobra.Command, args []string) {
				twoWasCalled = true
			},
		}
		rootCmd.AddCommand(oneCmd)
		rootCmd.AddCommand(twoCmd)
		if test.hasFlags {
			test.initFlagsFunc(&test, twoCmd)
		}
		rootCmd.Run(rootCmd, []string{})
		if rootWasCalled != test.rootWasCalled {
			t.Errorf("%s - unexpected root cmd call: expected: %v, got: %v", test.commandToCall, test.rootWasCalled, rootWasCalled)
		}
		if oneWasCalled != test.oneWasCalled {
			t.Errorf("%s - unexpected one cmd call: expected: %v, got: %v", test.commandToCall, test.oneWasCalled, oneWasCalled)
		}
		if twoWasCalled != test.twoWasCalled {
			t.Errorf("%s - unexpected two cmd call: expected: %v, got: %v", test.commandToCall, test.twoWasCalled, twoWasCalled)
		}
		if twoWasCalled != test.twoWasCalled {
			t.Errorf("%s - unexpected two cmd call: expected: %v, got: %v", test.commandToCall, test.twoWasCalled, twoWasCalled)
		}
		if !test.checkFunc(&test, t) {
			t.Error("error in checkFuncs")
		}
	}
}
