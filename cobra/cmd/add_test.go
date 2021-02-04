package cmd

import (
	"fmt"
	"os"
	"testing"
)

func TestBaseGoldenAddCmd(t *testing.T) {
	command := &Command{
		CmdName:   "testbase",
		CmdParent: parentName,
		Project:   getProject(),
	}
	defer os.RemoveAll(command.AbsolutePath)

	command.Project.Create()
	if err := command.Create(); err != nil {
		t.Fatal(err)
	}

	generatedFile := fmt.Sprintf("%s/cmd/%s.go", command.AbsolutePath, command.CmdName)
	goldenFile := fmt.Sprintf("testdata/%s.go.golden", command.CmdName)
	err := compareFiles(generatedFile, goldenFile)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGoldenAddCmd(t *testing.T) {
	command := &Command{
		CmdName:      "test",
		CmdParent:    parentName,
		CmdShortDesc: "A brief description of your command",
		CmdLongDesc: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
		Project: getProject(),
	}
	defer os.RemoveAll(command.AbsolutePath)

	command.Project.Create()
	if err := command.Create(); err != nil {
		t.Fatal(err)
	}

	generatedFile := fmt.Sprintf("%s/cmd/%s.go", command.AbsolutePath, command.CmdName)
	goldenFile := fmt.Sprintf("testdata/%s.go.golden", command.CmdName)
	err := compareFiles(generatedFile, goldenFile)
	if err != nil {
		t.Fatal(err)
	}
}

func TestValidateCmdName(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"cmdName", "cmdName"},
		{"cmd_name", "cmdName"},
		{"cmd-name", "cmdName"},
		{"cmd______Name", "cmdName"},
		{"cmd------Name", "cmdName"},
		{"cmd______name", "cmdName"},
		{"cmd------name", "cmdName"},
		{"cmdName-----", "cmdName"},
		{"cmdname-", "cmdname"},
	}

	for _, testCase := range testCases {
		got := validateCmdName(testCase.input)
		if testCase.expected != got {
			t.Errorf("Expected %q, got %q", testCase.expected, got)
		}
	}
}
