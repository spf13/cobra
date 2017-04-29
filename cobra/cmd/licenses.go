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

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

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

func init() {
	Licenses = make(map[string]License)

	// Allows a user to not use a license.
	Licenses["none"] = License{"None", []string{"none", "false"}, "", ""}

	initApache2()
	initMit()
	initBsdClause3()
	initBsdClause2()
	initGpl2()
	initGpl3()
	initLgpl()
	initAgpl()
}

// TODO: Inspect project for existing license
func getLicense() License {
	// If explicitly flagged, use that.
	if userLicense != "" {
		return findLicense(userLicense)
	}

	// If user wants to have custom license, use that.
	if viper.IsSet("license.header") || viper.IsSet("license.text") {
		return License{Header: viper.GetString("license.header"),
			Text: "license.text"}
	}

	// If user wants to have built-in license, use that.
	if viper.IsSet("license") {
		return findLicense(viper.GetString("license"))
	}

	// If user didn't set any license, use Apache 2.0 by default.
	fmt.Println("apache")
	return Licenses["apache"]
}

func copyrightLine() string {
	author := viper.GetString("author")
	year := time.Now().Format("2006")

	return "Copyright © " + year + " " + author
}

func findLicense(name string) License {
	found := matchLicense(name)
	if found == "" {
		er(fmt.Errorf("unknown license %q", name))
	}
	return Licenses[found]
}

// given a license name, try to match the license indicated
func matchLicense(name string) string {
	if name == "" {
		return ""
	}

	for key, lic := range Licenses {
		for _, match := range lic.PossibleMatches {
			if strings.EqualFold(name, match) {
				return key
			}
		}
	}

	return ""
}
