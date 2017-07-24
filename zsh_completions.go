package cobra

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

type fWriter struct {
	io.Writer
}

func (fw *fWriter) fWriteLn(format string, a ...interface{}) (int, error) {
	return io.WriteString(fw, fmt.Sprintf(format+"\n", a...))
}

// GenZshCompletion generates a zsh completion file and writes to the passed writer.
func (cmd *Command) GenZshCompletion(w io.Writer) error {
	buf := new(bytes.Buffer)
	fw := &fWriter{buf}

	writeHeader(fw, cmd)
	maxDepth := maxDepth(cmd)
	writeLevelMapping(fw, maxDepth)
	writeLevelCases(fw, maxDepth, cmd)

	_, err := buf.WriteTo(w)
	return err
}

func writeHeader(fw *fWriter, cmd *Command) {
	fw.fWriteLn("#compdef %s", cmd.Name())
	fw.fWriteLn("")
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

func writeLevelMapping(fw *fWriter, numLevels int) {
	fw.fWriteLn(`_arguments \`)
	for i := 1; i <= numLevels; i++ {
		fw.fWriteLn(`  '%d: :->level%d' \`, i, i)
	}
	fw.fWriteLn(`  '%d: :%s'`, numLevels+1, "_files")
	fw.fWriteLn("")
}

func writeLevelCases(fw *fWriter, maxDepth int, root *Command) {
	fw.fWriteLn("case $state in")
	defer fw.fWriteLn("esac")

	for i := 1; i <= maxDepth; i++ {
		fw.fWriteLn("  level%d)", i)
		writeLevel(fw, root, i)
		fw.fWriteLn("  ;;")
	}
	fw.fWriteLn("  *)")
	fw.fWriteLn("    _arguments '*: :_files'")
	fw.fWriteLn("  ;;")
}

func writeLevel(fw *fWriter, root *Command, i int) {
	fw.fWriteLn(fmt.Sprintf("    case $words[%d] in", i))
	defer fw.fWriteLn("    esac")

	commands := filterByLevel(root, i)
	byParent := groupByParent(commands)

	for p, c := range byParent {
		names := names(c)
		fw.fWriteLn(fmt.Sprintf("      %s)", p))
		fw.fWriteLn(fmt.Sprintf("        _arguments '%d: :(%s)'", i, strings.Join(names, " ")))
		fw.fWriteLn(fmt.Sprintf("      ;;"))
	}
	fw.fWriteLn("      *)")
	fw.fWriteLn("        _arguments '*: :_files'")
	fw.fWriteLn("      ;;")

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

func groupByParent(commands []*Command) map[string][]*Command {
	m := make(map[string][]*Command)
	for _, c := range commands {
		parent := c.Parent()
		if parent == nil {
			continue
		}
		m[parent.Name()] = append(m[parent.Name()], c)
	}
	return m
}

func names(commands []*Command) []string {
	ns := make([]string, len(commands))
	for i, c := range commands {
		ns[i] = c.Name()
	}
	return ns
}
