package cobra

import (
	"fmt"
	"io"
	"os"
	"text/template"

	"github.com/spf13/pflag"
)

var (
	funcMap = template.FuncMap{
		"genZshFuncName": generateZshCompletionFuncName,
		"extractFlags":   extractFlags,
		"simpleFlag":     simpleFlag,
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
{{ $cmdPath := genZshFuncName .}}
function {{$cmdPath}} {
  local -a commands

  _arguments -C \{{- range extractFlags .}}
    {{if simpleFlag .}}{{template "simpleFlag" .}}{{else}}{{template "complexFlag" .}}{{end}} \{{- end}}
    "1: :->cmnds" \
    "*::arg:->args"

  case $state in
  cmnds)
    commands=({{range .Commands}}{{if not .Hidden}}
      "{{.Name}}:{{.Short}}"{{end}}{{end}}
    )
    _describe "command" commands
    ;;
  esac

  case "$words[1]" in {{- range .Commands}}{{if not .Hidden}}
  {{.Name}})
    {{$cmdPath}}_{{.Name}}
    ;;{{end}}{{end}}
  esac
}
{{range .Commands}}{{if not .Hidden}}
{{template "selectCmdTemplate" .}}
{{- end}}{{end}}
{{- end}}

{{/* should accept Command without subcommands as parameter */}}
{{define "arguments" -}}
function {{genZshFuncName .}} {
{{"  _arguments"}}{{range extractFlags .}} \
    {{if simpleFlag .}}{{template "simpleFlag" .}}{{else}}{{template "complexFlag" .}}{{end -}}
{{end}}
}
{{end}}

{{/* dispatcher for commands with or without subcommands */}}
{{define "selectCmdTemplate" -}}
{{if .Hidden}}{{/* ignore hidden*/}}{{else -}}
{{if .Commands}}{{template "argumentsC" .}}{{else}}{{template "arguments" .}}{{end}}
{{- end}}
{{- end}}

{{/* template entry point */}}
{{define "Main" -}}
#compdef _{{.Name}} {{.Name}}

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

func generateZshCompletionFuncName(c *Command) string {
	if c.HasParent() {
		return generateZshCompletionFuncName(c.Parent()) + "_" + c.Name()
	}
	return "_" + c.Name()
}

func extractFlags(c *Command) []*pflag.Flag {
	var flags []*pflag.Flag
	c.LocalFlags().VisitAll(func(f *pflag.Flag) {
		if !f.Hidden {
			flags = append(flags, f)
		}
	})
	c.InheritedFlags().VisitAll(func(f *pflag.Flag) {
		if !f.Hidden {
			flags = append(flags, f)
		}
	})
	return flags
}

func simpleFlag(p *pflag.Flag) bool {
	return p.Name == "" || p.Shorthand == ""
}
