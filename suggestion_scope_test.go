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
	"strings"
	"testing"
)

func TestOnlyValidArgsSuggestionScope(t *testing.T) {
	for _, tt := range []struct {
		name     string
		disabled bool
		distance int
		valid    []string
		args     []string
		want     string
	}{
		{"described argument", false, 0, []string{"one", "two\tsecond value"}, []string{"one", "twe"}, "invalid argument \"twe\" for \"root\"\n\nDid you mean this?\n\tone\n\ttwo\n"},
		{"no command suggestion", false, 0, []string{"one", "remote"}, []string{"statux"}, "invalid argument \"statux\" for \"root\""},
		{"disabled", true, 0, []string{"two"}, []string{"twe"}, "invalid argument \"twe\" for \"root\""},
		{"distance", false, 1, []string{"three"}, []string{"twe"}, "invalid argument \"twe\" for \"root\""},
	} {
		t.Run(tt.name, func(t *testing.T) {
			c := &Command{Use: "root", Args: OnlyValidArgs, ValidArgs: tt.valid, DisableSuggestions: tt.disabled, SuggestionsMinimumDistance: tt.distance, Run: emptyRun}
			c.AddCommand(&Command{Use: "status", Run: emptyRun})
			_, err := executeCommand(c, tt.args...)
			if err == nil || err.Error() != tt.want {
				t.Fatalf("got %v; want %q", err, tt.want)
			}
		})
	}
}

func TestUnknownCommandDoesNotSuggestValidArgs(t *testing.T) {
	c := &Command{Use: "root", ValidArgs: []string{"remote\targument description"}}
	c.AddCommand(&Command{Use: "status", Run: emptyRun})
	_, err := executeCommand(c, "remot")
	if err == nil || strings.Contains(err.Error(), "Did you mean") {
		t.Fatalf("argument leaked into command suggestions: %v", err)
	}
}
