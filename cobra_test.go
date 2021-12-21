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

func TestCheckErr(t *testing.T) {
	tests := []struct {
		name  string
		msg   interface{}
		panic bool
	}{
		{
			name:  "no error",
			msg:   nil,
			panic: false,
		},
		{
			name:  "panic string",
			msg:   "test",
			panic: true,
		},
		{
			name:  "panic error",
			msg:   errors.New("test error"),
			panic: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if r != nil {
					if !tt.panic {
						t.Error("Didn't expect panic")
					}
				} else {
					if tt.panic {
						t.Error("Expected to panic")
					}
				}
			}()
			CheckErr(tt.msg)
		})
	}
}
