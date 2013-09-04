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
	"bytes"
	"fmt"
	flag "github.com/ogier/pflag"
	"os"
	"strings"
)

var _ = flag.ContinueOnError

type Flag interface {
	Args() []string
}

type Flags interface {
	Lookup(string) *Flag
	VisitAll(fn func(*Flag))
	Parse(arguments []string) error
}

// A Commander holds the configuration for the command line tool.
type Commander struct {
	// A Commander is also a Command for top level and global help & flags
	Command
	//ExitOnError, ContinueOnError or PanicOnError
	behavior flag.ErrorHandling

	args []string
}

func NewCommander() (c *Commander) {
	c = new(Commander)
	return
}

func (c *Commander) setFlagBehavior(b flag.ErrorHandling) error {
	if b == flag.ExitOnError || b == flag.ContinueOnError || b == flag.PanicOnError {
		c.behavior = b
		return nil
	}
	return fmt.Errorf("%v is not a valid behavior", b)
}

func (c *Commander) SetName(name string) {
	c.name = name
}

func (c *Commander) SetArgs(a []string) {
	c.args = a
}

func (c *Commander) Execute() (err error) {
	if len(c.args) == 0 {
		err = c.execute(os.Args[1:])
	} else {
		err = c.execute(c.args)
	}
	return
}

// Command is just that, a command for your application.
// eg.  'go run' ... 'run' is the command. Cobra requires
// you to define the usage and description as part of your command
// definition to ensure usability.
type Command struct {
	// Name is the command name, usually the executable's name.
	name string
	// The one-line usage message.
	Use string
	// The short description shown in the 'help' output.
	Short string
	// The long message shown in the 'help <this-command>' output.
	Long string
	// Set of flags specific to this command.
	flags *flag.FlagSet
	// Set of flags children commands will inherit
	pflags *flag.FlagSet
	// Run runs the command.
	// The args are the arguments after the command name.
	Run func(cmd *Command, args []string)
	// Commands is the list of commands supported by this Commander program.
	commands []*Command
	// Parent Command for this command
	parent *Command
	// Commander
	//cmdr *Commander
	flagErrorBuf *bytes.Buffer
}

// find the target command given the args and command tree
func (c *Command) Find(args []string) (cmd *Command, a []string, err error) {
	if c == nil {
		return nil, nil, fmt.Errorf("Called find() on a nil Command")
	}

	validSubCommand := false
	if len(args) > 1 && c.HasSubCommands() {
		for _, cmd := range c.commands {
			if cmd.Name() == args[0] {
				validSubCommand = true
				return cmd.Find(args[1:])
			}
		}
	}
	if !validSubCommand && c.Runnable() {
		return c, args, nil
	}

	return nil, nil, nil
}

// execute the command determined by args and the command tree
func (c *Command) execute(args []string) (err error) {
	err = fmt.Errorf("unknown subcommand %q\nRun 'help' for usage.\n", args[0])

	if c == nil {
		return fmt.Errorf("Called Execute() on a nil Command")
	}

	cmd, a, e := c.Find(args)
	if e == nil {
		err = cmd.ParseFlags(a)
		if err != nil {
			return err
		} else {
			argWoFlags := cmd.Flags().Args()
			cmd.Run(cmd, argWoFlags)
			return nil
		}
	}
	err = e
	return err
}

// Add one or many commands as children of this
func (c *Command) AddCommand(cmds ...*Command) {
	for i, x := range cmds {
		cmds[i].parent = c
		c.commands = append(c.commands, x)
	}
}

// The full usage for a given command (including parents)
func (c *Command) Usage(depth ...int) string {
	i := 0
	if len(depth) > 0 {
		i = depth[0]
	}

	if c.HasParent() {
		return c.parent.Usage(i+1) + " " + c.Use
	} else if i > 0 {
		return c.Name()
	} else {
		return c.Use
	}
}

// Usage prints the usage details to the standard output.
func (c *Command) PrintUsage() {
	if c.Runnable() {
		fmt.Printf("usage: %s\n\n", c.Usage())
	}

	fmt.Println(strings.Trim(c.Long, "\n"))
}

// Name returns the command's name: the first word in the use line.
func (c *Command) Name() string {
	if c.name != "" {
		return c.name
	}
	name := c.Use
	i := strings.Index(name, " ")
	if i >= 0 {
		name = name[:i]
	}
	return name
}

// Determine if the command is itself runnable
func (c *Command) Runnable() bool {
	return c.Run != nil
}

// Determine if the command has children commands
func (c *Command) HasSubCommands() bool {
	return len(c.commands) > 0
}

// Determine if the command has children commands
func (c *Command) HasParent() bool {
	return c.parent != nil
}

// Get the Commands FlagSet
func (c *Command) Flags() *flag.FlagSet {
	if c.flags == nil {
		c.flags = flag.NewFlagSet(c.Name(), flag.ContinueOnError)
		if c.flagErrorBuf == nil {
			c.flagErrorBuf = new(bytes.Buffer)
		}
		c.flags.SetOutput(c.flagErrorBuf)
	}
	return c.flags
}

// Get the Commands Persistent FlagSet
func (c *Command) PersistentFlags() *flag.FlagSet {
	if c.pflags == nil {
		c.pflags = flag.NewFlagSet(c.Name(), flag.ContinueOnError)
		if c.flagErrorBuf == nil {
			c.flagErrorBuf = new(bytes.Buffer)
		}
		c.pflags.SetOutput(c.flagErrorBuf)
	}
	return c.flags
}

// Intended for use in testing
func (c *Command) ResetFlags() {
	c.flagErrorBuf = new(bytes.Buffer)
	c.flags = flag.NewFlagSet(c.Name(), flag.ContinueOnError)
	c.flags.SetOutput(c.flagErrorBuf)
	c.pflags = flag.NewFlagSet(c.Name(), flag.ContinueOnError)
	c.pflags.SetOutput(c.flagErrorBuf)
}

func (c *Command) HasFlags() bool {
	return hasFlags(c.flags)
}

func (c *Command) HasPersistentFlags() bool {
	return hasFlags(c.pflags)
}

// Is this set of flags not empty
func hasFlags(f *flag.FlagSet) bool {
	if f == nil {
		return false
	}
	if f.NFlag() != 0 {
		return true
	}
	return false
}

// Climbs up the command tree looking for matching flag
func (c *Command) Flag(name string) (flag *flag.Flag) {
	flag = c.Flags().Lookup(name)

	if flag == nil {
		flag = c.persistentFlag(name)
	}

	return
}

// recursively find matching persistent flag
func (c *Command) persistentFlag(name string) (flag *flag.Flag) {
	if c.HasPersistentFlags() {
		flag = c.PersistentFlags().Lookup(name)
	}

	if flag == nil && c.HasParent() {
		flag = c.parent.persistentFlag(name)
	}
	return
}

// Parses persistent flag tree & local flags
func (c *Command) ParseFlags(args []string) (err error) {
	err = c.ParsePersistentFlags(args)
	if err != nil {
		return err
	}
	err = c.Flags().Parse(args)
	if err != nil {
		return err
	}
	return nil
}

// Climbs up the command tree parsing flags from top to bottom
func (c *Command) ParsePersistentFlags(args []string) (err error) {
	if !c.HasParent() || (c.parent.HasPersistentFlags() && c.parent.PersistentFlags().Parsed()) {
		if c.HasPersistentFlags() {
			err = c.PersistentFlags().Parse(args)
			if err != nil {
				return err
			}
		}
	} else {
		if c.HasParent() && c.parent.HasPersistentFlags() {
			err = c.parent.ParsePersistentFlags(args)
			if err != nil {
				return err
			}
		}
	}
	return
}
