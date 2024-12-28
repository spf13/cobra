// Copyright 2013-2024 The Cobra Authors
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

//go:build !go1.21
// +build !go1.21

package cobra

import (
	"os"
	"strings"
)

// based on golang.org/x/mod/internal/lazyregexp: https://cs.opensource.google/go/x/mod/+/refs/tags/v0.19.0:internal/lazyregexp/lazyre.go;l=66
// For a non-go-test program which still has a name ending with ".test[.exe]", it will need to either:
// 1- Use go >= 1.21, or
// 2- call "rootCmd.SetArgs(os.Args[1:])" before calling "rootCmd.Execute()"
var inTest = len(os.Args) > 0 && strings.HasSuffix(strings.TrimSuffix(os.Args[0], ".exe"), ".test")

func isTesting() bool {
	return inTest
}
