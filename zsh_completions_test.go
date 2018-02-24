package cobra

import (
	"bytes"
	"regexp"
	"testing"
)

func TestGenZshCompletion(t *testing.T) {
	var debug bool
	var option string

	tcs := []struct {
		name                string
		root                *Command
		expectedExpressions []string
	}{
		{
			name: "simple command",
			root: func() *Command {
				r := &Command{
					Use:  "mycommand",
					Long: "My Command long description",
				}
				r.Flags().BoolVar(&debug, "debug", debug, "description")
				return r
			}(),
			expectedExpressions: []string{
				`function _mycommand {\s+_arguments \\\s+"--debug\[description\]"\s+}`,
				"#compdef _mycommand mycommand",
			},
		},
		{
			name: "flags with both long and short flags",
			root: func() *Command {
				r := &Command{
					Use:  "testcmd",
					Long: "long description",
				}
				r.Flags().BoolVarP(&debug, "debug", "d", debug, "debug description")
				return r
			}(),
			expectedExpressions: []string{
				`"\(-d --debug\)"{-d,--debug}"\[debug description\]"`,
			},
		},
		{
			name: "command with subcommands",
			root: func() *Command {
				r := &Command{
					Use:  "rootcmd",
					Long: "Long rootcmd description",
				}
				d := &Command{
					Use:   "subcmd1",
					Short: "Subcmd1 short descrition",
				}
				e := &Command{
					Use:  "subcmd2",
					Long: "Subcmd2 short description",
				}
				r.PersistentFlags().BoolVar(&debug, "debug", debug, "description")
				d.Flags().StringVarP(&option, "option", "o", option, "option description")
				r.AddCommand(d, e)
				return r
			}(),
			expectedExpressions: []string{
				`\\\n\s+"1: :\(subcmd1 subcmd2\)" \\\n`,
				`_arguments \\\n.*"--debug\[description]"`,
				`_arguments -C \\\n.*"--debug\[description]"`,
				`function _rootcmd_subcmd1 {`,
				`function _rootcmd_subcmd1 {`,
				`_arguments \\\n.*"\(-o --option\)"{-o,--option}"\[option description]" \\\n`,
			},
		},
		{
			name: "filename completion",
			root: func() *Command {
				var file string
				r := &Command{
					Use:   "mycmd",
					Short: "my command short description",
				}
				r.Flags().StringVarP(&file, "config", "c", file, "config file")
				r.MarkFlagFilename("config", "ext")
				return r
			}(),
			expectedExpressions: []string{
				`\n +"\(-c --config\)"{-c,--config}"\[config file]:filename:_files"`,
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			tc.root.GenZshCompletion(buf)
			output := buf.Bytes()

			for _, expr := range tc.expectedExpressions {
				rgx, err := regexp.Compile(expr)
				if err != nil {
					t.Errorf("error compiling expression (%s): %v", expr, err)
				}
				if !rgx.Match(output) {
					t.Errorf("expeced completion (%s) to match '%s'", buf.String(), expr)
				}
			}
		})
	}

}

func BenchmarkConstructPath(b *testing.B) {
	c := &Command{
		Use:   "main",
		Long:  "main long description which is very long",
		Short: "main short description",
	}
	d := &Command{
		Use: "hello",
	}
	e := &Command{
		Use: "world",
	}
	c.AddCommand(d)
	d.AddCommand(e)
	for i := 0; i < b.N; i++ {
		res := constructPath(e)
		if res != "_main_hello_world" {
			b.Errorf("expeced path to be '_main_hello_world', got %s", res)
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

	resC := extractFlags(c)
	resD := extractFlags(d)

	if len(resC) != 2 {
		t.Errorf("expected Command C to return 2 flags, got %d", len(resC))
	}
	if len(resD) != 2 {
		t.Errorf("expected Command D to return 2 flags, got %d", len(resD))
	}
}
