package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
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

// GenZshCompletion generates a zsh completion file and writes to the passed writer.
func (c *Command) GenZshCompletion(w io.Writer) error {
	buf := new(bytes.Buffer)

	writeHeader(buf, c)
	maxDepth := maxDepth(c)
	writeLevelMapping(buf, maxDepth)
	writeLevelCases(buf, maxDepth, c)

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

func writeLevelMapping(w io.Writer, numLevels int) {
	fmt.Fprintln(w, "local -a cmd_options")
	fmt.Fprintln(w, `_arguments -C  \`)
	fmt.Fprintln(w, `  $jamf_pro_options \`)
	for i := 1; i <= numLevels; i++ {
		fmt.Fprintf(w, `  '%d: :->level%d' \`, i, i)
		fmt.Fprintln(w)
	}
	fmt.Fprintf(w, `  '%d: :%s'`, numLevels+1, "_files")
	fmt.Fprintln(w)
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

func writeLevel(w io.Writer, root *Command, level int) {
	fmt.Fprintf(w, "  level%d)\n", level)
	fmt.Fprintf(w, "    case $words[%d] in\n", level)
	for _, c := range filterByLevel(root, level) {
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
        fmt.Fprintln(w, "        cmd_options=(")
		for _, flag := range flags {
			fmt.Fprintf(w, "            %s\n", flag)
		}
		fmt.Fprintln(w, "\n       )")
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
		ns[i] = fmt.Sprintf("%s[%s]", c.Name(), c.Short)
	}
	return ns
}

func commandFlags(command *Command) []string {
	flags := command.Flags()
	ns := make([]string, 0)
	flags.VisitAll(func(flag *pflag.Flag) {
		if len(flag.Shorthand) > 0 {
			ns = append(ns, fmt.Sprintf("{-%s,--%s}'[%s]'", flag.Shorthand, flag.Name, flag.Usage))
		} else {
			ns = append(ns, fmt.Sprintf("--%s'[%s]'", flag.Name, flag.Usage))
		}
	})
	return ns
}
