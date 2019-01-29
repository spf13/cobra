package tpl

func MainTemplate() []byte {
	return []byte(`
/*
{{ .Copyright }}
{{if .Legal.Header}}{{ .Legal.Header }}{{end}}
*/
package main

import "{{ .PkgName }}/cmd"

func main() {
	cmd.Execute()
}
`)
}
