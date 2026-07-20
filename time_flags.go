// Copyright 2013-2026 The Cobra Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cobra

import (
	"time"

	"github.com/spf13/pflag"
)

// defaultTimeLayout is the layout used by the Time / TimeP / TimeVar /
// TimeVarP helpers that do not take an explicit layout. It is RFC 3339 with
// nano precision, e.g. "2024-06-01T12:00:00Z".
const defaultTimeLayout = time.RFC3339Nano

// timeValue implements pflag.Value for a time.Time flag. The layout controls
// how values are parsed (Set) and formatted (String). See issue #742.
type timeValue struct {
	v      *time.Time
	layout string
}

// NewTimeValue returns a pflag.Value backed by p, initialised to value and
// using layout to parse/format flag strings. Use it with (*Command).Flags().Var
// or VarP for full control, or with the Time / TimeLayout helpers below.
func NewTimeValue(value time.Time, layout string, p *time.Time) pflag.Value {
	*p = value
	return &timeValue{v: p, layout: layout}
}

func (t *timeValue) Set(s string) error {
	parsed, err := time.Parse(t.layout, s)
	if err != nil {
		return err
	}
	*t.v = parsed
	return nil
}

func (t *timeValue) Type() string { return "time" }

func (t *timeValue) String() string {
	if t.v == nil || t.v.IsZero() {
		return ""
	}
	return t.v.Format(t.layout)
}

// Time defines a time.Time flag with the given name, default value and usage
// string, parsed/formatted as RFC 3339. It returns a pointer to the value.
func (c *Command) Time(name string, value time.Time, usage string) *time.Time {
	return c.TimeP(name, "", value, usage)
}

// TimeP is like Time but accepts a shorthand letter.
func (c *Command) TimeP(name, shorthand string, value time.Time, usage string) *time.Time {
	p := new(time.Time)
	c.TimeVarP(p, name, shorthand, value, usage)
	return p
}

// TimeVar defines a time.Time flag bound to p.
func (c *Command) TimeVar(p *time.Time, name string, value time.Time, usage string) {
	c.TimeVarP(p, name, "", value, usage)
}

// TimeVarP is like TimeVar but accepts a shorthand letter.
func (c *Command) TimeVarP(p *time.Time, name, shorthand string, value time.Time, usage string) {
	c.Flags().VarP(NewTimeValue(value, defaultTimeLayout, p), name, shorthand, usage)
}

// TimeLayout is like Time but parses and formats values using the given layout
// (a reference layout string as accepted by time.Parse, e.g.
// "2006-01-02 15:04:05").
func (c *Command) TimeLayout(name, layout string, value time.Time, usage string) *time.Time {
	return c.TimeLayoutP(name, "", layout, value, usage)
}

// TimeLayoutP is like TimeLayout but accepts a shorthand letter.
func (c *Command) TimeLayoutP(name, shorthand, layout string, value time.Time, usage string) *time.Time {
	p := new(time.Time)
	c.TimeLayoutVarP(p, name, shorthand, layout, value, usage)
	return p
}

// TimeLayoutVar defines a time.Time flag bound to p, parsed/formatted using
// the given layout.
func (c *Command) TimeLayoutVar(p *time.Time, name, layout string, value time.Time, usage string) {
	c.TimeLayoutVarP(p, name, "", layout, value, usage)
}

// TimeLayoutVarP is like TimeLayoutVar but accepts a shorthand letter.
func (c *Command) TimeLayoutVarP(p *time.Time, name, shorthand, layout string, value time.Time, usage string) {
	c.Flags().VarP(NewTimeValue(value, layout, p), name, shorthand, usage)
}