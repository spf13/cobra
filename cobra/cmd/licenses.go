// Copyright © 2015 Steve Francia <spf@spf13.com>.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Parts inspired by https://github.com/ryanuber/go-license

package cmd

import "strings"

//Licenses contains all possible licenses a user can chose from
var Licenses map[string]License

//License represents a software license agreement, containing the Name of
// the license, its possible matches (on the command line as given to cobra)
// the header to be used with each file on the file's creating, and the text
// of the license
type License struct {
	Name            string   // The type of license in use
	PossibleMatches []string // Similar names to guess
	Text            string   // License text data
	Header          string   // License header for source files
}

// given a license name (in), try to match the license indicated
func matchLicense(in string) string {
	for key, lic := range Licenses {
		for _, match := range lic.PossibleMatches {
			if strings.EqualFold(in, match) {
				return key
			}
		}
	}
	return ""
}

func init() {
	Licenses = make(map[string]License)

	// Allows a user to not use a license.
	Licenses["none"] = License{"None", []string{"none", "false"}, "", ""}

	// Allows a user to use config for a custom license.
	Licenses["custom"] = License{"Custom", []string{}, "", ""}

	initApache2()

	initMit()

	initBsdClause3()

	initBsdClause2()

	initGpl2()

	initGpl3()

	// Licenses["apache20"] = License{
	// 	Name:            "Apache 2.0",
	// 	PossibleMatches: []string{"apache", "apache20", ""},
	//   Header: `
	//   `,
	// 	Text: `
	//   `,
	// }
}
