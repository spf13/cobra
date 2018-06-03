package cobra

import (
	"bytes"
	"regexp"
	"strings"
	"testing"
)

func TestGenZshCompletion(t *testing.T) {
	var debug bool
	var option string

	tcs := []struct {
		name                string
		root                *Command
		expectedExpressions []string
		invocationArgs      []string
		skip                string
	}{
		{
			name: "simple command",
			root: func() *Command {
				r := &Command{
					Use:  "mycommand",
					Long: "My Command long description",
					Run:  emptyRun,
				}
				r.Flags().BoolVar(&debug, "debug", debug, "description")
				return r
			}(),
			expectedExpressions: []string{
				`(?s)function _mycommand {\s+_arguments \\\s+'--debug\[description\]'.*--help.*}`,
				"#compdef _mycommand mycommand",
			},
		},
		{
			name: "flags with both long and short flags",
			root: func() *Command {
				r := &Command{
					Use:  "testcmd",
					Long: "long description",
					Run:  emptyRun,
				}
				r.Flags().BoolVarP(&debug, "debug", "d", debug, "debug description")
				return r
			}(),
			expectedExpressions: []string{
				`'\(-d --debug\)'{-d,--debug}'\[debug description\]'`,
			},
		},
		{
			name: "command with subcommands and flags with values",
			root: func() *Command {
				r := &Command{
					Use:  "rootcmd",
					Long: "Long rootcmd description",
				}
				d := &Command{
					Use:   "subcmd1",
					Short: "Subcmd1 short description",
					Run:   emptyRun,
				}
				e := &Command{
					Use:  "subcmd2",
					Long: "Subcmd2 short description",
					Run:  emptyRun,
				}
				r.PersistentFlags().BoolVar(&debug, "debug", debug, "description")
				d.Flags().StringVarP(&option, "option", "o", option, "option description")
				r.AddCommand(d, e)
				return r
			}(),
			expectedExpressions: []string{
				`commands=\(\n\s+"help:.*\n\s+"subcmd1:.*\n\s+"subcmd2:.*\n\s+\)`,
				`_arguments \\\n.*'--debug\[description]'`,
				`_arguments -C \\\n.*'--debug\[description]'`,
				`function _rootcmd_subcmd1 {`,
				`function _rootcmd_subcmd1 {`,
				`_arguments \\\n.*'\(-o --option\)'{-o,--option}'\[option description]:' \\\n`,
			},
		},
		{
			name: "filename completion with and without globs",
			root: func() *Command {
				var file string
				r := &Command{
					Use:   "mycmd",
					Short: "my command short description",
					Run:   emptyRun,
				}
				r.Flags().StringVarP(&file, "config", "c", file, "config file")
				r.MarkFlagFilename("config")
				r.Flags().String("output", "", "output file")
				r.MarkFlagFilename("output", "*.log", "*.txt")
				return r
			}(),
			expectedExpressions: []string{
				`\n +'\(-c --config\)'{-c,--config}'\[config file]:filename:_files'`,
				`:_files -g "\*.log" -g "\*.txt"`,
			},
		},
		{
			name: "repeated variables both with and without value",
			root: func() *Command {
				r := genTestCommand("mycmd", true)
				_ = r.Flags().BoolSliceP("debug", "d", []bool{}, "debug usage")
				_ = r.Flags().StringArray("option", []string{}, "options")
				return r
			}(),
			expectedExpressions: []string{
				`'\*--option\[options]`,
				`'\(\*-d \*--debug\)'{\\\*-d,\\\*--debug}`,
			},
		},
		{
			name: "generated flags --help and --version should be created even when not executing root cmd",
			root: func() *Command {
				r := &Command{
					Use:     "mycmd",
					Short:   "mycmd short description",
					Version: "myversion",
				}
				s := genTestCommand("sub1", true)
				r.AddCommand(s)
				return s
			}(),
			expectedExpressions: []string{
				"--version",
				"--help",
			},
			invocationArgs: []string{
				"sub1",
			},
			skip: "--version and --help are currently not generated when not running on root command",
		},
		{
			name: "zsh generation should run on root command",
			root: func() *Command {
				r := genTestCommand("root", false)
				s := genTestCommand("sub1", true)
				r.AddCommand(s)
				return s
			}(),
			expectedExpressions: []string{
				"function _root {",
			},
		},
		{
			name: "flag description with single quote (') shouldn't break quotes in completion file",
			root: func() *Command {
				r := genTestCommand("root", true)
				r.Flags().Bool("private", false, "Don't show public info")
				return r
			}(),
			expectedExpressions: []string{
				`--private\[Don'\\''t show public info]`,
			},
		},
		{
			name: "argument completion for file with and without patterns",
			root: func() *Command {
				r := genTestCommand("root", true)
				r.MarkZshCompPositionalArgumentFile(1, "*.log")
				r.MarkZshCompPositionalArgumentFile(2)
				return r
			}(),
			expectedExpressions: []string{
				`'1: :_files -g "\*.log"' \\\n\s+'2: :_files`,
			},
		},
		{
			name: "argument zsh completion for words",
			root: func() *Command {
				r := genTestCommand("root", true)
				r.MarkZshCompPositionalArgumentWords(1, "word1", "word2")
				return r
			}(),
			expectedExpressions: []string{
				`'1: :\("word1" "word2"\)`,
			},
		},
		{
			name: "argument completion for words with spaces",
			root: func() *Command {
				r := genTestCommand("root", true)
				r.MarkZshCompPositionalArgumentWords(1, "single", "multiple words")
				return r
			}(),
			expectedExpressions: []string{
				`'1: :\("single" "multiple words"\)'`,
			},
		},
		{
			name: "argument completion when command has ValidArgs and no annotation for argument completion",
			root: func() *Command {
				r := genTestCommand("root", true)
				r.ValidArgs = []string{"word1", "word2"}
				return r
			}(),
			expectedExpressions: []string{
				`'1: :\("word1" "word2"\)'`,
			},
		},
		{
			name: "argument completion when command has ValidArgs and no annotation for argument at argPosition 1",
			root: func() *Command {
				r := genTestCommand("root", true)
				r.ValidArgs = []string{"word1", "word2"}
				r.MarkZshCompPositionalArgumentFile(2)
				return r
			}(),
			expectedExpressions: []string{
				`'1: :\("word1" "word2"\)' \\`,
			},
		},
		{
			name: "directory completion for flag",
			root: func() *Command {
				r := genTestCommand("root", true)
				r.Flags().String("test", "", "test")
				r.PersistentFlags().String("ptest", "", "ptest")
				r.MarkFlagDirname("test")
				r.MarkPersistentFlagDirname("ptest")
				return r
			}(),
			expectedExpressions: []string{
				`--test\[test]:filename:_files -g "-\(/\)"`,
				`--ptest\[ptest]:filename:_files -g "-\(/\)"`,
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			if tc.skip != "" {
				t.Skip(tc.skip)
			}
			tc.root.Root().SetArgs(tc.invocationArgs)
			tc.root.Execute()
			buf := new(bytes.Buffer)
			if err := tc.root.GenZshCompletion(buf); err != nil {
				t.Error(err)
			}
			output := buf.Bytes()

			for _, expr := range tc.expectedExpressions {
				rgx, err := regexp.Compile(expr)
				if err != nil {
					t.Errorf("error compiling expression (%s): %v", expr, err)
				}
				if !rgx.Match(output) {
					t.Errorf("expected completion (%s) to match '%s'", buf.String(), expr)
				}
			}
		})
	}
}

func TestGenZshCompletionHidden(t *testing.T) {
	tcs := []struct {
		name                string
		root                *Command
		expectedExpressions []string
	}{
		{
			name: "hidden commands",
			root: func() *Command {
				r := &Command{
					Use:   "main",
					Short: "main short description",
				}
				s1 := &Command{
					Use:    "sub1",
					Hidden: true,
					Run:    emptyRun,
				}
				s2 := &Command{
					Use:   "sub2",
					Short: "short sub2 description",
					Run:   emptyRun,
				}
				r.AddCommand(s1, s2)

				return r
			}(),
			expectedExpressions: []string{
				"sub1",
			},
		},
		{
			name: "hidden flags",
			root: func() *Command {
				var hidden string
				r := &Command{
					Use:   "root",
					Short: "root short description",
					Run:   emptyRun,
				}
				r.Flags().StringVarP(&hidden, "hidden", "H", hidden, "hidden usage")
				if err := r.Flags().MarkHidden("hidden"); err != nil {
					t.Errorf("Error setting flag hidden: %v\n", err)
				}
				return r
			}(),
			expectedExpressions: []string{
				"--hidden",
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tc.root.Execute()
			buf := new(bytes.Buffer)
			if err := tc.root.GenZshCompletion(buf); err != nil {
				t.Error(err)
			}
			output := buf.String()

			for _, expr := range tc.expectedExpressions {
				if strings.Contains(output, expr) {
					t.Errorf("Expected completion (%s) not to contain '%s' but it does", output, expr)
				}
			}
		})
	}
}

func TestMarkZshCompPositionalArgumentFile(t *testing.T) {
	t.Run("Doesn't allow overwriting existing positional argument", func(t *testing.T) {
		c := &Command{}
		if err := c.MarkZshCompPositionalArgumentFile(1, "*.log"); err != nil {
			t.Errorf("Received error when we shouldn't have: %v\n", err)
		}
		if err := c.MarkZshCompPositionalArgumentFile(1); err == nil {
			t.Error("Didn't receive an error when trying to overwrite argument position")
		}
	})

	t.Run("Refuses to accept argPosition less then 1", func(t *testing.T) {
		c := &Command{}
		err := c.MarkZshCompPositionalArgumentFile(0, "*")
		if err == nil {
			t.Fatal("Error was not thrown when indicating argument position 0")
		}
		if !strings.Contains(err.Error(), "position") {
			t.Errorf("expected error message '%s' to contain 'position'", err.Error())
		}
	})
}

func TestMarkZshCompPositionalArgumentWords(t *testing.T) {
	t.Run("Doesn't allow overwriting existing positional argument", func(t *testing.T) {
		c := &Command{}
		if err := c.MarkZshCompPositionalArgumentFile(1, "*.log"); err != nil {
			t.Errorf("Received error when we shouldn't have: %v\n", err)
		}
		if err := c.MarkZshCompPositionalArgumentWords(1, "hello"); err == nil {
			t.Error("Didn't receive an error when trying to overwrite argument position")
		}
	})

	t.Run("Doesn't allow calling without words", func(t *testing.T) {
		c := &Command{}
		if err := c.MarkZshCompPositionalArgumentWords(0); err == nil {
			t.Error("Should not allow saving empty word list for annotation")
		}
	})

	t.Run("Refuses to accept argPosition less then 1", func(t *testing.T) {
		c := &Command{}
		err := c.MarkZshCompPositionalArgumentWords(0, "word")
		if err == nil {
			t.Fatal("Should not allow setting argument position less then 1")
		}
		if !strings.Contains(err.Error(), "position") {
			t.Errorf("Expected error '%s' to contain 'position' but didn't", err.Error())
		}
	})
}

func BenchmarkMediumSizeConstruct(b *testing.B) {
	root := constructLargeCommandHierarchy()
	// if err := root.GenZshCompletionFile("_mycmd"); err != nil {
	// 	b.Error(err)
	// }

	for i := 0; i < b.N; i++ {
		buf := new(bytes.Buffer)
		err := root.GenZshCompletion(buf)
		if err != nil {
			b.Error(err)
		}
	}
}

func TestExtractFlags(t *testing.T) {
	var debug, cmdc, cmdd bool
	c := &Command{
		Use:  "cmdC",
		Long: "Command C",
	}
	c.PersistentFlags().BoolVarP(&debug, "debug", "d", debug, "debug mode")
	c.Flags().BoolVar(&cmdc, "cmd-c", cmdc, "Command C")
	d := &Command{
		Use:  "CmdD",
		Long: "Command D",
	}
	d.Flags().BoolVar(&cmdd, "cmd-d", cmdd, "Command D")
	c.AddCommand(d)

	resC := zshCompExtractFlag(c)
	resD := zshCompExtractFlag(d)

	if len(resC) != 2 {
		t.Errorf("expected Command C to return 2 flags, got %d", len(resC))
	}
	if len(resD) != 2 {
		t.Errorf("expected Command D to return 2 flags, got %d", len(resD))
	}
}

func constructLargeCommandHierarchy() *Command {
	var config, st1, st2 string
	var long, debug bool
	var in1, in2 int
	var verbose []bool

	r := genTestCommand("mycmd", false)
	r.PersistentFlags().StringVarP(&config, "config", "c", config, "config usage")
	if err := r.MarkPersistentFlagFilename("config", "*"); err != nil {
		panic(err)
	}
	s1 := genTestCommand("sub1", true)
	s1.Flags().BoolVar(&long, "long", long, "long description")
	s1.Flags().BoolSliceVar(&verbose, "verbose", verbose, "verbose description")
	s1.Flags().StringArray("option", []string{}, "various options")
	s2 := genTestCommand("sub2", true)
	s2.PersistentFlags().BoolVar(&debug, "debug", debug, "debug description")
	s3 := genTestCommand("sub3", true)
	s3.Hidden = true
	s1_1 := genTestCommand("sub1sub1", true)
	s1_1.Flags().StringVar(&st1, "st1", st1, "st1 description")
	s1_1.Flags().StringVar(&st2, "st2", st2, "st2 description")
	s1_2 := genTestCommand("sub1sub2", true)
	s1_3 := genTestCommand("sub1sub3", true)
	s1_3.Flags().IntVar(&in1, "int1", in1, "int1 description")
	s1_3.Flags().IntVar(&in2, "int2", in2, "int2 description")
	s1_3.Flags().StringArrayP("option", "O", []string{}, "more options")
	s2_1 := genTestCommand("sub2sub1", true)
	s2_2 := genTestCommand("sub2sub2", true)
	s2_3 := genTestCommand("sub2sub3", true)
	s2_4 := genTestCommand("sub2sub4", true)
	s2_5 := genTestCommand("sub2sub5", true)

	s1.AddCommand(s1_1, s1_2, s1_3)
	s2.AddCommand(s2_1, s2_2, s2_3, s2_4, s2_5)
	r.AddCommand(s1, s2, s3)
	r.Execute()
	return r
}

func genTestCommand(name string, withRun bool) *Command {
	r := &Command{
		Use:   name,
		Short: name + " short description",
		Long:  "Long description for " + name,
	}
	if withRun {
		r.Run = emptyRun
	}

	return r
}
