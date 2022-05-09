package cobra

import (
	"errors"
	"testing"
	"text/template"
)

func assertNoErr(t *testing.T, e error) {
	if e != nil {
		t.Error(e)
	}
}

func TestAddTemplateFunctions(t *testing.T) {
	AddTemplateFunc("t", func() bool { return true })
	AddTemplateFuncs(template.FuncMap{
		"f": func() bool { return false },
		"h": func() string { return "Hello," },
		"w": func() string { return "world." }})

	c := &Command{}
	c.SetUsageTemplate(`{{if t}}{{h}}{{end}}{{if f}}{{h}}{{end}} {{w}}`)

	const expected = "Hello, world."
	if got := c.UsageString(); got != expected {
		t.Errorf("Expected UsageString: %v\nGot: %v", expected, got)
	}
}

func Test_OnInitialize(t *testing.T) {
	call := false
	c := &Command{Use: "c", Run: emptyRun}
	OnInitialize(func() {
		call = true
	})
	_, err := executeCommand(c)
	if err != nil {
		t.Error(err)
	}
	if !call {
		t.Error("expected OnInitialize func to be called")
	}
}

func Test_OnInitializeE(t *testing.T) {
	c := &Command{Use: "c", Run: emptyRun}
	e := errors.New("test error")
	OnInitializeE(func() error {
		return e
	})
	_, err := executeCommand(c)
	if err != e {
		t.Error("expected error: %w", e)
	}
}
