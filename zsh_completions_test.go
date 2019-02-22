package cobra

import (
	"bytes"
	"strings"
	"testing"
)

const (
	expectedTail = `BASH_COMPLETION_EOF
}

__cobra_bash_source <(__trivialapp_convert_bash_to_zsh)
_complete trivialapp 2>/dev/null
`
	expectedHead = `#compdef trivialapp`
)

func TestZshCompletion(t *testing.T) {
	root := &Command{Use: "trivialapp"}

	buf := &bytes.Buffer{}
	err := root.GenZshCompletion(buf)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	actual := buf.String()

	if !strings.HasPrefix(actual, expectedHead) {
		t.Error("Unexpected head")
	}
	if !strings.HasSuffix(actual, expectedTail) {
		t.Error("Unexpected tail")
	}
}
