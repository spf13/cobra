package cmd

import (
	"errors"
	"fmt"
	"strings"
)

var specialFlagTypes = []string{"stringMapFlag", "genericMapFlag"}

// Indexes for flag definition string after split {name}:{type}:{description}
const (
	FlagDefLength = 3
	FlagDefSeparator = ":"
	FlagDefault = ` "",`
	FlagNameIndex = 0
	FlagTypeIndex = 1
	FlagDescriptionIndex = 2
)

type flagDefinition struct {
	Name string
	FlagType string
	FlagDescription string
	CreateFn string
	VarName string
	Default string
	CmdName string
}

func printFlagVars(iFlags interface{}) string {
	flags := toFlagArray(iFlags)
	varDefs := ""
	for _, f := range flags {
		varDefs += fmt.Sprintf("var %s %s\r\n", f.VarName, f.FlagType)
	}
	return varDefs
}

func printFlagCreates(iFlags interface{}, persistent bool) string {
	flags := toFlagArray(iFlags)
	createStrs := ""
	flagsFn := "Flags"
	if persistent {
		flagsFn = "PersistentFlags"
	}
	for _, f := range flags {
		createStrs += fmt.Sprintf( `
%sCmd.%s().%s(&%s, "%s",%s "%s")`,
			f.CmdName,
			flagsFn,
			f.CreateFn,
			f.VarName,
			f.Name,
			f.Default,
			f.FlagDescription)
	}
	return createStrs
}

func toFlagArray(value interface{}) []flagDefinition {
	switch v := value.(type) {
	case []flagDefinition:
		return v
	// Add whatever other types you need
	default:
		return []flagDefinition{}
	}
}

func buildFlag(in, cmdName string) (flagDefinition, error) {
	var def flagDefinition
	values := strings.Split(in, FlagDefSeparator)
	if len(values) != FlagDefLength {
		return def, errors.New("invalid flag definition, make sure to specify flags as 'name:type:definition' they are all required")
	}
	name, flagType, desc := values[FlagNameIndex], values[FlagTypeIndex],values[FlagDescriptionIndex]
	def = flagDefinition{
		Name: name,
		FlagType: flagType,
		FlagDescription: desc,
		CreateFn: fmt.Sprintf("%s%s", strings.Title(flagType), "Var"),
		VarName: fmt.Sprintf("%s%sFlag", cmdName, name),
		Default:FlagDefault,
		CmdName: cmdName,
	}

	return def, nil
}
