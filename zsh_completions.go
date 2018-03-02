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
		"genZshFuncName":              generateZshCompletionFuncName,
		"extractFlags":                extractFlags,
		"genFlagEntryForZshArguments": genFlagEntryForZshArguments,
	}
	zshCompletionText = `
{{/* should accept Command (that contains subcommands) as parameter */}}
{{define "argumentsC" -}}
{{ $cmdPath := genZshFuncName .}}
function {{$cmdPath}} {
  local -a commands

  _arguments -C \{{- range extractFlags .}}
    {{genFlagEntryForZshArguments .}} \{{- end}}
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
    {{genFlagEntryForZshArguments . -}}
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

// genFlagEntryForZshArguments returns an entry that matches _arguments
// zsh-completion parameters. It's too complicated to generate in a template.
func genFlagEntryForZshArguments(f *pflag.Flag) string {
	if f.Name == "" || f.Shorthand == "" {
		return genFlagEntryForSingleOptionFlag(f)
	}
	return genFlagEntryForMultiOptionFlag(f)
}

func genFlagEntryForSingleOptionFlag(f *pflag.Flag) string {
	var option, multiMark, extras string

	if flagCouldBeSpecifiedMoreThenOnce(f) {
		multiMark = "*"
	}

	option = "--" + f.Name
	if option == "--" {
		option = "-" + f.Shorthand
	}
	extras = genZshFlagEntryExtras(f)

	return fmt.Sprintf(`'%s%s[%s]%s'`, multiMark, option, f.Usage, extras)
}

func genFlagEntryForMultiOptionFlag(f *pflag.Flag) string {
	var options, parenMultiMark, curlyMultiMark, extras string

	if flagCouldBeSpecifiedMoreThenOnce(f) {
		parenMultiMark = "*"
		curlyMultiMark = "\\*"
	}

	options = fmt.Sprintf(`'(%s-%s %s--%s)'{%s-%s,%s--%s}`,
		parenMultiMark, f.Shorthand, parenMultiMark, f.Name, curlyMultiMark, f.Shorthand, curlyMultiMark, f.Name)
	extras = genZshFlagEntryExtras(f)

	return fmt.Sprintf(`%s'[%s]%s'`, options, f.Usage, extras)
}

func genZshFlagEntryExtras(f *pflag.Flag) string {
	var extras string

	globs, pathSpecified := f.Annotations[BashCompFilenameExt]
	if pathSpecified {
		extras = ":filename:_files"
		for _, g := range globs {
			extras = extras + fmt.Sprintf(` -g "%s"`, g)
		}
	} else if f.NoOptDefVal == "" {
		extras = ":" // allow option variable without assisting
	}

	return extras
}

func flagCouldBeSpecifiedMoreThenOnce(f *pflag.Flag) bool {
	return strings.Contains(f.Value.Type(), "Slice") ||
		strings.Contains(f.Value.Type(), "Array")
}
