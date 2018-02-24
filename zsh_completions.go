package cobra

import (
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"

	"github.com/spf13/pflag"
)

var (
	funcMap = template.FuncMap{
		"constructPath": constructPath,
		"subCmdList":    subCmdList,
		"extractFlags":  extractFlags,
		"simpleFlag":    simpleFlag,
	}
	zshCompletionText = `
{{/* for pflag.Flag (specifically annotations) */}}
{{define "flagAnnotations" -}}
{{with index .Annotations "cobra_annotation_bash_completion_filename_extensions"}}:filename:_files{{end}}
{{- end}}

{{/* for pflag.Flag with short and long options */}}
{{define "complexFlag" -}}
"(-{{.Shorthand}} --{{.Name}})"{-{{.Shorthand}},--{{.Name}}}"[{{.Usage}}]{{template "flagAnnotations" .}}"
{{- end}}

{{/* for pflag.Flag with either short or long options */}}
{{define "simpleFlag" -}}
"{{with .Name}}--{{.}}{{else}}-{{.Shorthand}}{{end}}[{{.Usage}}]{{template "flagAnnotations" .}}"
{{- end}}

{{/* should accept Command (that contains subcommands) as parameter */}}
{{define "argumentsC" -}}
function {{constructPath .}} {
  local line

  _arguments -C \
{{range extractFlags . -}}
{{"    "}}{{if simpleFlag .}}{{template "simpleFlag" .}}{{else}}{{template "complexFlag" .}}{{end}} \
{{end}}    "1: :({{subCmdList .}})" \
    "*::arg:->args"

    case $line[1] in {{- range .Commands}}
        {{.Use}})
            {{constructPath .}}
            ;;
{{end}}    esac
}
{{range .Commands}}
{{template "selectCmdTemplate" .}}
{{- end}}
{{- end}}

{{/* should accept Command without subcommands as parameter */}}
{{define "arguments" -}}
function {{constructPath .}} {
{{with extractFlags . -}}
{{ "  _arguments" -}}
{{range .}} \
    {{if simpleFlag .}}{{template "simpleFlag" .}}{{else}}{{template "complexFlag" .}}{{end -}}
{{end}}
{{end -}}
}
{{- end}}

{{define "selectCmdTemplate" -}}
{{if .Commands}}{{template "argumentsC" .}}{{else}}{{template "arguments" .}}{{end}}
{{- end}}

{{define "Main" -}}
#compdef _{{.Use}} {{.Use}}

{{template "selectCmdTemplate" .}}
{{end}}
`
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
	tmpl, err := template.New("Main").Funcs(funcMap).Parse(zshCompletionText)
	if err != nil {
		return fmt.Errorf("error creating zsh completion template: %v", err)
	}
	return tmpl.Execute(w, c)
}

func constructPath(c *Command) string {
	var path []string
	tmpCmd := c
	path = append(path, tmpCmd.Use)

	for {
		if !tmpCmd.HasParent() {
			break
		}
		tmpCmd = tmpCmd.Parent()
		path = append(path, tmpCmd.Use)
	}

	// reverse path
	for left, right := 0, len(path)-1; left < right; left, right = left+1, right-1 {
		path[left], path[right] = path[right], path[left]
	}

	return "_" + strings.Join(path, "_")
}

// subCmdList returns a space separated list of subcommands names
func subCmdList(c *Command) string {
	var subCmds []string

	for _, cmd := range c.Commands() {
		subCmds = append(subCmds, cmd.Use)
	}

	return strings.Join(subCmds, " ")
}

func extractFlags(c *Command) []*pflag.Flag {
	var flags []*pflag.Flag
	c.LocalFlags().VisitAll(func(f *pflag.Flag) {
		flags = append(flags, f)
	})
	c.InheritedFlags().VisitAll(func(f *pflag.Flag) {
		flags = append(flags, f)
	})
	return flags
}

func simpleFlag(p *pflag.Flag) bool {
	return p.Name == "" || p.Shorthand == ""
}
