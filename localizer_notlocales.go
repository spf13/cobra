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

//go:build !locales
// +build !locales

package cobra

import (
	"embed"
)

// localeFS points to an embedded filesystem of binary gettext translation files,
// but only for the default (english) language, as the locales build tag was not set.
// For performance and smaller builds, only the binary MO files are included.
// Their sibling PO files should still be considered their authoritative source.
//
//go:embed locales/*/en.mo
var localeFS embed.FS
