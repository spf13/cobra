// Copyright 2013-2023 The Cobra Authors
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
	"fmt"
	"strings"
	"testing"
)

// InvalidArgCountError

func TestInvalidArgCountError_GetCommand(t *testing.T) {
	expected := &Command{}
	err := InvalidArgCountError{cmd: expected}

	got := err.GetCommand()
	if got != expected {
		t.Errorf("expected %v, got %v",
			getCommandName(expected), getCommandName(got))
	}
}

func TestInvalidArgCountError_GetArgs(t *testing.T) {
	expected := []string{"a", "b", "c"}
	err := InvalidArgCountError{args: expected}

	got := err.GetArguments()
	if strings.Join(expected, " ") != strings.Join(got, " ") {
		t.Fatalf("expected %v, got %v", expected, got)
	}
}

func TestInvalidArgCountError_GetMinArgumentCount(t *testing.T) {
	expected := 1
	err := InvalidArgCountError{atLeast: expected}

	got := err.GetMinArgumentCount()
	if got != expected {
		t.Fatalf("expected %v, got %v", expected, got)
	}
}

func TestInvalidArgCountError_GetMaxArgumentCount(t *testing.T) {
	expected := 1
	err := InvalidArgCountError{atMost: expected}

	got := err.GetMaxArgumentCount()
	if got != expected {
		t.Fatalf("expected %v, got %v", expected, got)
	}
}

// InvalidArgValueError

func TestInvalidArgValueError_GetCommand(t *testing.T) {
	expected := &Command{}
	err := InvalidArgValueError{cmd: expected}

	got := err.GetCommand()
	if got != expected {
		t.Errorf("expected %v, got %v",
			getCommandName(expected), getCommandName(got))
	}
}

func TestInvalidArgValueError_GetArgument(t *testing.T) {
	expected := "a"
	err := InvalidArgValueError{arg: expected}

	got := err.GetArgument()
	if got != expected {
		t.Fatalf("expected %v, got %v", expected, got)
	}
}

func TestInvalidArgValueError_GetSuggestions(t *testing.T) {
	expected := []string{"a", "b"}
	err := InvalidArgValueError{suggestions: expected}

	got := err.GetSuggestions()
	expectedString := fmt.Sprintf("%#v", expected)
	gotString := fmt.Sprintf("%#v", got)
	if expectedString != gotString {
		t.Fatalf("expected %v, got %v", expected, got)
	}
}

// UnknownSubcommandError

func TestUnknownSubcommandError_GetCommand(t *testing.T) {
	expected := &Command{}
	err := UnknownSubcommandError{cmd: expected}

	got := err.GetCommand()
	if got != expected {
		t.Errorf("expected %v, got %v",
			getCommandName(expected), getCommandName(got))
	}
}

func TestUnknownSubcommandError_GetSubcommand(t *testing.T) {
	expected := "a"
	err := UnknownSubcommandError{subcmd: expected}

	got := err.GetSubcommand()
	if got != expected {
		t.Fatalf("expected %v, got %v", expected, got)
	}
}

func TestUnknownSubcommandError_GetSuggestions(t *testing.T) {
	expected := []string{"a", "b"}
	err := UnknownSubcommandError{suggestions: expected}

	got := err.GetSuggestions()
	expectedString := fmt.Sprintf("%#v", expected)
	gotString := fmt.Sprintf("%#v", got)
	if expectedString != gotString {
		t.Fatalf("expected %v, got %v", expected, got)
	}
}

// RequiredFlagError

func TestRequiredFlagError_GetCommand(t *testing.T) {
	expected := &Command{}
	err := UnknownSubcommandError{cmd: expected}

	got := err.GetCommand()
	if got != expected {
		t.Errorf("expected %v, got %v",
			getCommandName(expected), getCommandName(got))
	}
}

func TestRequiredFlagError_GetFlags(t *testing.T) {
	expected := []string{"a", "b", "c"}
	err := RequiredFlagError{missingFlagNames: expected}

	got := err.GetFlags()
	if strings.Join(expected, " ") != strings.Join(got, " ") {
		t.Fatalf("expected %v, got %v", expected, got)
	}
}

// FlagGroupError

func TestFlagGroupError_GetCommand(t *testing.T) {
	expected := &Command{}
	err := FlagGroupError{cmd: expected}

	got := err.GetCommand()
	if got != expected {
		t.Errorf("expected %v, got %v",
			getCommandName(expected), getCommandName(got))
	}
}

func TestFlagGroupError_GetFlags(t *testing.T) {
	expected := []string{"a", "b", "c"}
	err := FlagGroupError{flagList: "a b c"}

	got := err.GetFlags()
	if strings.Join(expected, " ") != strings.Join(got, " ") {
		t.Fatalf("expected %v, got %v", expected, got)
	}
}

func TestFlagGroupError_GetProblemFlags(t *testing.T) {
	expected := []string{"a", "b", "c"}
	err := FlagGroupError{problemFlags: expected}

	got := err.GetProblemFlags()
	if strings.Join(expected, " ") != strings.Join(got, " ") {
		t.Fatalf("expected %v, got %v", expected, got)
	}
}

func TestFlagGroupError_Unwrap(t *testing.T) {
	expected := ErrFlagsAreMutuallyExclusive
	err := FlagGroupError{err: expected}

	got := err.Unwrap()
	if got != expected {
		t.Fatalf("expected %v, got %v", expected, got)
	}
}
