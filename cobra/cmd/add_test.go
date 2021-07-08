package cmd

import (
	"fmt"
	"os"
	"testing"
)

func TestGoldenAddCmd(t *testing.T) {
	command := &Command{
		CmdName:     "test",
		CmdUse:      "test",
		CmdFileName: "test",
		CmdParent:   parentName,
		Project:     getProject(),
	}
	defer os.RemoveAll(command.AbsolutePath)

	assertNoErr(t, command.Project.Create())
	assertNoErr(t, command.Create())

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

func TestGenerateCmdFileName(t *testing.T) {
	testCases := []struct {
		inputCmd     string
		inputParent  string
		expectedCmd  string
		expectedFile string
	}{
		{"cmdname", "parent", "parentCmdname", "parent_cmdname"},
		{"cmdname", "parentCmd", "parentCmdname", "parent_cmdname"},
		{"CmdName", "ParentCmd", "parentCmdName", "parent_cmd_name"},
		{"cmdname", "granpaParentCmd", "granpaParentCmdname", "granpa_parent_cmdname"},
		{"test", "granpaParentCmd", "granpaParentTest", "granpa_parent_testcmd"},
		{"cmdname", "rootCmd", "cmdname", "cmdname"},
	}

	for _, testCase := range testCases {
		got1, got2 := generateCmdFileName(testCase.inputCmd, testCase.inputParent)
		if testCase.expectedCmd != got1 || testCase.expectedFile != got2 {
			t.Errorf(
				"Expected %q and %q, got %q and %q",
				testCase.expectedCmd, testCase.expectedFile,
				got1, got2,
			)
		}
	}
}
