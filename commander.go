// Copyright © 2013 Steve Francia <spf@spf13.com>.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Commands similar to git, go tools and other modern CLI tools
// inspired by go, go-Commander, gh and subcommand

package cobra

import (
	"fmt"
	"io"
	"os"
)

// A Commander holds the configuration for the command line tool.
type Commander struct {
	// A Commander is also a Command for top level and global help & flags
	Command

	args          []string
	output        *io.Writer               // nil means stderr; use out() accessor
	UsageFunc     func(*Command) error     // Usage can be defined by application
	UsageTemplate string                   // Can be defined by Application
	HelpTemplate  string                   // Can be defined by Application
	HelpFunc      func(*Command, []string) // Help can be defined by application
	HelpCommand   *Command
}

// Provide the user with a new commander.
func NewCommander() (c *Commander) {
	c = new(Commander)
	c.cmdr = c
	c.UsageFunc = c.defaultUsage
	c.HelpFunc = c.defaultHelp
	c.initTemplates()
	return
}

func (c *Commander) initHelp() {
	if c.HelpCommand == nil {
		c.HelpCommand = &Command{
			Use:   "help [command to learn about]",
			Short: "Help about any command",
			Long: `Help provides help for any command in the application.
    Simply type ` + c.Name() + ` help [path to command] for full details.`,
			Run: c.HelpFunc,
		}
	}
	c.AddCommand(c.HelpCommand)
}

// Name for commander, should match application name
func (c *Commander) SetName(name string) {
	c.name = name
}

// os.Args[1:] by default, if desired, can be overridden
// particularly useful when testing.
func (c *Commander) SetArgs(a []string) {
	c.args = a
}

// Call execute to use the args (os.Args[1:] by default)
// and run through the command tree finding appropriate matches
// for commands and then corresponding flags.
func (c *Commander) Execute() (err error) {
	// initialize help as the last point possible to allow for user
	// overriding
	c.initHelp()
	if len(c.args) == 0 {
		err = c.execute(os.Args[1:])
	} else {
		err = c.execute(c.args)
	}
	return
}

func (c *Commander) out() io.Writer {
	if c.output == nil {
		return os.Stderr
	}
	return *c.output
}

func (cmdr *Commander) defaultUsage(c *Command) error {
	err := tmpl(cmdr.out(), cmdr.UsageTemplate, c)
	return err
}

func (cmdr *Commander) defaultHelp(c *Command, args []string) {
	cmd, _, e := c.Root().Find(args)
	if cmd == nil {
		cmdr.Printf("Unknown help topic %#q.  Run '%v help'.\n", args, cmdr.Name())
		return
	}
	if e != nil {
		cmdr.Printf("Unknown help topic %#q.  Run '%v help'.\n", args, cmdr.Name())
	} else {
		err := tmpl(cmdr.out(), cmdr.HelpTemplate, cmd)
		if err != nil {
			c.Println(err)
		}
	}
}

//Print to out
func (c *Commander) PrintOut(i ...interface{}) {
	fmt.Fprint(c.out(), i...)
}

// SetOutput sets the destination for usage and error messages.
// If output is nil, os.Stderr is used.
func (c *Commander) SetOutput(output io.Writer) {
	c.output = &output
}

func (c *Commander) initTemplates() {
	c.UsageTemplate = `{{ $cmd := . }}
Usage: {{if .Runnable}}
  {{.UseLine}}{{if .HasFlags}} [flags]{{end}}{{end}}{{if .HasSubCommands}}
  {{ .CommandPath}} [command]{{end}}
{{ if .HasSubCommands}}
Available Commands: {{range .Commands}}{{if .Runnable}}
  {{.Use | printf "%-11s"}} :: {{.Short}}{{end}}{{end}}
{{end}}
{{ if .HasFlags}} Available Flags:
{{.Flags.FlagUsages}}{{end}}{{if and (gt .Commands 0) (gt .Parent.Commands 1) }}
Additional help topics: {{if gt .Commands 0 }}{{range .Commands}}{{if not .Runnable}}
  {{.CommandPath | printf "%-11s"}} :: {{.Short}}{{end}}{{end}}{{end}}{{if gt .Parent.Commands 1 }}{{range .Parent.Commands}}{{if .Runnable}}{{if not (eq .Name $cmd.Name) }}{{end}}
  {{.CommandPath | printf "%-11s"}} :: {{.Short}}{{end}}{{end}}{{end}}

{{end}}
Use "{{.Commander.Name}} help [command]" for more information about that command.
`

	c.HelpTemplate = `{{.Name}}
{{.Long | trim}}
{{if .Runnable}}{{.Usage}}{{end}}
`
}
