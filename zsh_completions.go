package main

import (
	"os"
	"io"
	"bytes"
	"fmt"
	"github.com/spf13/cobra"
"github.com/spf13/pflag"
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
	fmt.Fprintln(w, `_arguments \`)
	for i := 1; i <= numLevels; i++ {
		fmt.Fprintf(w, `  '%d: :->level%d' \`, i, i)
		fmt.Fprintln(w)
	}
	fmt.Fprintf(w, `  '%d: :%s'`, numLevels+1, "_files")
	fmt.Fprintln(w)
}

func writeLevelCases(w io.Writer, maxDepth int, root *Command) {
	fmt.Fprintln(w, "case $state in")
	defer fmt.Fprintln(w, "esac")

	for i := 1; i <= maxDepth; i++ {
		fmt.Fprintf(w, "  level%d)\n", i)
		writeLevel(w, root, i)
		fmt.Fprintln(w, "  ;;")
	}
	fmt.Fprintln(w, "  *)")
	fmt.Fprintln(w, "    _arguments '*: :_files'")
	fmt.Fprintln(w, "  ;;")
}

func writeLevel(w io.Writer, root *Command, i int) {
	fmt.Fprintf(w, "    case $words[%d] in\n", i)
	defer fmt.Fprintln(w, "    esac")

	commands := filterByLevel(root, i)
	byParent := groupByParent(commands)

	for p, c := range byParent {
		names := names(c)
		//flags := flags(p)
		fmt.Fprintf(w, "      %s)\n", p.Name())
		fmt.Fprintf(w, "        _values '%s command' ", p.Name())
		fmt.Fprintf(w, "'%s'\n", strings.Join(names, "' '"))
		fmt.Fprintln(w, "      ;;")
	}
	fmt.Fprintln(w, "      *)")
	fmt.Fprintln(w, "        _arguments '*: :_files'")
	fmt.Fprintln(w, "      ;;")

}

func filterByLevel(c *Command, l int) []*Command {
	cs := make([]*Command, 0)
	if l == 0 {
		cs = append(cs, c)
		return cs
	}
	for _, s := range c.Commands() {
		cs = append(cs, filterByLevel(s, l-1)...)
	}
	return cs
}

func groupByParent(commands []*Command) map[*Command][]*Command {
	m := make(map[*Command][]*Command)
	for _, c := range commands {
		parent := c.Parent()
		if parent == nil {
			continue
		}
		m[parent] = append(m[parent], c)
	}
	return m
}

func names(commands []*Command) []string {
	ns := make([]string, len(commands))
	for i, c := range commands {
		ns[i] = fmt.Sprintf("%s[%s]", c.Name(), c.Short)
	}
	return ns
}

func flags(command *Command) []string {
	flags := command.Flags()
	ns := make([]string, flags.NArg())
	if flags.NArg() > 0 {
		i := 0
		flags.VisitAll(func(f *pflag.Flag) {
			ns[i] = fmt.Sprintf("%s[%s]", f.Name, f.Usage)
			i += 1
		})
	}
	return ns
}
