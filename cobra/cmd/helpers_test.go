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
		t.Errorf("Unexpected Project Path. \nGot: %q\nExpected: %q\nArg: %v\nDir: %v", projectPath, expected, input, wd)
	}

	reset()
}

func reset() {
	testWd = ""
	inputPath = ""
	projectPath = ""
}

func TestProjectPath(t *testing.T) {
	checkGuess(t, "", filepath.Join("github.com", "spf13", "hugo"), filepath.Join(getSrcPath(), "github.com", "spf13", "hugo"))
	checkGuess(t, "", filepath.Join("spf13", "hugo"), filepath.Join(getSrcPath(), "spf13", "hugo"))
	checkGuess(t, "", filepath.Join("/", "bar", "foo"), filepath.Join("/", "bar", "foo"))
	checkGuess(t, "/bar/foo", "baz", filepath.Join(getSrcPath(), "github.com", "baz"))
	checkGuess(t, "/bar/foo", "gopkg.in/baz", filepath.Join(getSrcPath(), "gopkg.in", "baz"))
	checkGuess(t, "/bar/foo", "./baz", filepath.Join("/", "bar", "foo", "baz"))
	checkGuess(t, "/bar/foo", "./gopkg.in/baz", filepath.Join("/", "bar", "foo", "gopkg.in", "baz"))
	checkGuess(t, "/bar/foo/cmd", "", filepath.Join("/", "bar", "foo"))
	checkGuess(t, "/bar/foo/command", "", filepath.Join("/", "bar", "foo"))
	checkGuess(t, "/bar/foo/commands", "", filepath.Join("/", "bar", "foo"))
	checkGuess(t, "github.com/spf13/hugo/../hugo", "", filepath.Join("github.com", "spf13", "hugo"))
}
