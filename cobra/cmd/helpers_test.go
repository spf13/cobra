package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

var _ = fmt.Println
var _ = os.Stderr

func checkGuess(t *testing.T, wd, input, expected string) {
	testWd = wd
	inputPath = input
	guessProjectPath()

	if projectPath != expected {
		t.Errorf("Unexpected Project Path. \n Got: %q\nExpected: %q\n", projectPath, expected)
	}

	reset()
}

func reset() {
	testWd = ""
	inputPath = ""
	projectPath = ""
}

func TestProjectPath(t *testing.T) {
	wd, _ := os.Getwd()

	checkGuess(t, "", "github.com/spf13/hugo", filepath.Join(wd, "github.com", "spf13", "hugo"))
	checkGuess(t, "", "spf13/hugo", filepath.Join(wd, "spf13", "hugo"))
	checkGuess(t, "", "/bar/foo", filepath.Join(wd, "bar", "foo"))
	checkGuess(t, "/bar/foo", "baz", filepath.Join("/", "bar", "foo", "baz"))
	checkGuess(t, "/bar/foo/cmd", "", filepath.Join("/", "bar", "foo"))
	checkGuess(t, "/bar/foo/command", "", filepath.Join("/", "bar", "foo"))
	checkGuess(t, "/bar/foo/commands", "", filepath.Join("/", "bar", "foo"))
	checkGuess(t, "github.com/spf13/hugo/../hugo", "", filepath.Join("github.com", "spf13", "hugo"))
}
