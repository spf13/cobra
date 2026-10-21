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

//go:build windows
// +build windows

package cobra

import (
	"os"
	"testing"
)

func TestMousetrapApplies(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		helpText     string
		fromExplorer bool
		want         bool
	}{
		{"double-clicked", []string{"app.exe"}, MousetrapHelpText, true, true},
		{"explorer with arguments", []string{"app.exe", "uninstall", "--purge"}, MousetrapHelpText, true, false},
		{"not from explorer", []string{"app.exe"}, MousetrapHelpText, false, false},
		{"disabled", []string{"app.exe"}, "", true, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			defer func(args []string, helpText string, started func() bool) {
				os.Args, MousetrapHelpText, startedByExplorer = args, helpText, started
			}(os.Args, MousetrapHelpText, startedByExplorer)
			os.Args, MousetrapHelpText = tc.args, tc.helpText
			startedByExplorer = func() bool { return tc.fromExplorer }

			if got := mousetrapApplies(); got != tc.want {
				t.Errorf("mousetrapApplies() = %v, want %v", got, tc.want)
			}
		})
	}
}
