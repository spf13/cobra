package cmd

import (
	"fmt"
	"os"
	"testing"
)

func TestGoldenAddCmd(t *testing.T) {

	wd, _ := os.Getwd()
	command := &Command{
		CmdName:   "test",
		CmdParent: parentName,
		Project: &Project{
			AbsolutePath: fmt.Sprintf("%s/testproject", wd),
			Legal:        getLicense(),
			Copyright:    copyrightLine(),

			// required to init
			AppName: "testproject",
			PkgName: "github.com/spf13/testproject",
			Viper:   true,
		},
	}

	// init project first
	command.Project.Create()
	defer func() {
		if _, err := os.Stat(command.AbsolutePath); err == nil {
			os.RemoveAll(command.AbsolutePath)
		}
	}()

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
