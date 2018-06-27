package cobra

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	flag "github.com/spf13/pflag"
)

// GenZshCompletionFile generates zsh completion file.
func (c *Command) GenZshCompletionFile(filename string) error {
	outFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer outFile.Close()

	return c.GenZshCompletion(outFile)
}

func argName(cmd *Command) string {
	for cmd.HasParent() {
		cmd = cmd.Parent()
	}
	name := fmt.Sprintf("%s_cmd_args", cmd.Name())
	return strings.Replace(name, "-", "_",-1)
}
// GenZshCompletion generates a zsh completion file and writes to the passed writer.
func (c *Command) GenZshCompletion(w io.Writer) error {
	buf := new(bytes.Buffer)

	writeHeader(buf, c)
	maxDepth := maxDepth(c)
	fmt.Fprintf(buf, "_%s() {\n", c.Name())
	writeLevelMapping(buf, maxDepth, c)
	writeLevelCases(buf, maxDepth, c)
	fmt.Fprintf(buf, "}\n_%s \"$@\"\n", c.Name())

	_, err := buf.WriteTo(w)
	return err
}

func writeHeader(w io.Writer, cmd *Command) {
	fmt.Fprintf(w, "#compdef %s\n\n", cmd.Name())
}

func maxDepth(c *Command) int {
	if len(c.Commands()) == 0 {
		return 0
	}
	maxDepthSub := 0
	for _, s := range c.Commands() {
		subDepth := maxDepth(s)
		if subDepth > maxDepthSub {
			maxDepthSub = subDepth
		}
	}
	return 1 + maxDepthSub
}

func writeLevelMapping(w io.Writer, numLevels int, root *Command) {
	fmt.Fprintln(w, `local context curcontext="$curcontext" state line`)
	fmt.Fprintln(w, `typeset -A opt_args`)
	fmt.Fprintln(w, `_arguments -C  \`)
	for i := 1; i <= numLevels; i++ {
		fmt.Fprintf(w, "  '%d: :->level%d' \\\n", i, i)
	}
	fmt.Fprintf(w, "  $%s \\\n", argName(root))
	fmt.Fprintln(w, `  '*: :_files'`)
}

func writeLevelCases(w io.Writer, maxDepth int, root *Command) {
	fmt.Fprintln(w, "case $state in")
	for i := 1; i <= maxDepth; i++ {
		writeLevel(w, root, i)
	}
	fmt.Fprintln(w, "  *)")
	fmt.Fprintln(w, "    _arguments '*: :_files'")
	fmt.Fprintln(w, "  ;;")
	fmt.Fprintln(w, "esac")
}

func writeLevel(w io.Writer, root *Command, l int) {
	fmt.Fprintf(w, "  level%d)\n", l)
	fmt.Fprintf(w, "    case $words[%d] in\n", l)
	for _, c := range filterByLevel(root, l) {
		writeCommandArgsBlock(w, c)
	}
	fmt.Fprintln(w, "      *)")
	fmt.Fprintln(w, "        _arguments '*: :_files'")
	fmt.Fprintln(w, "      ;;")
	fmt.Fprintln(w, "    esac")
	fmt.Fprintln(w, "  ;;")
}

func writeCommandArgsBlock(w io.Writer, c *Command) {
	names := commandNames(c)
	flags := commandFlags(c)
	if len(names) > 0 || len(flags) > 0 {
		fmt.Fprintf(w, "      %s)\n", c.Name())
		defer fmt.Fprintln(w, "      ;;")
	}
	if len(flags) > 0 {
        fmt.Fprintf(w, "        %s=(\n", argName(c))
		for _, flag := range flags {
			fmt.Fprintf(w, "            %s\n", flag)
		}
		fmt.Fprintln(w, "        )")
	}
	if len(names) > 0 {
		fmt.Fprintf(w, "        _values 'command' '%s'\n", strings.Join(names, "' '"))
	}
}

func filterByLevel(c *Command, l int) []*Command {
	commands := []*Command{c}
	for i := 1; i < l; i++ {
		var nextLevel []*Command
		for _, c := range commands {
			if c.HasSubCommands() {
				nextLevel = append(nextLevel, c.Commands()...)
			}
		}
		commands = nextLevel
	}

	return commands
}

func commandNames(command *Command) []string {
	commands := command.Commands()
	ns := make([]string, len(commands))
	for i, c := range commands {
        commandMsg := c.Name()
        if len(c.Short) > 0 {
            commandMsg += fmt.Sprintf("[%s]", c.Short)
        }
		ns[i] = commandMsg
	}
	return ns
}

func commandFlags(command *Command) []string {
	flags := command.Flags()
	ns := make([]string, 0)
	flags.VisitAll(func(flag *flag.Flag) {
        var flagMsg string
		if len(flag.Shorthand) > 0 {
			flagMsg = fmt.Sprintf("{-%s,--%s}", flag.Shorthand, flag.Name)
		} else {
			flagMsg = fmt.Sprintf("--%s", flag.Name)
		}
        if len(flag.Usage) > 0 {
            flagMsg += fmt.Sprintf("'[%s]'", flag.Usage)
        }
        ns = append(ns, flagMsg)
	})
	return ns
}
