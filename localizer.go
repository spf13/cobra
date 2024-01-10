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

package cobra

import (
	"fmt"
	"os"

	"github.com/leonelquinteros/gotext"
	"golang.org/x/text/language"
)

var defaultLanguage = language.English

// envVariablesHoldingLocale is sorted by decreasing priority.
// These environment variables are expected to hold a parsable locale (fr_FR, es, en-US, …)
var envVariablesHoldingLocale = []string{
	"LANGUAGE",
	"LC_ALL",
	"LC_MESSAGES",
	"LANG",
}

// availableLocalizationDomains holds all the domains used in localization.
// Each domain MUST have its own locales/<domain>.pot file and locales/<domain>/ dir.
// Therefore, please only use short, ^[a-z]+$ strings as domains.
var availableLocalizationDomains = []string{
	"default",
}

// i18nCommandGlossary wraps the translated strings passed to the command usage template.
// This is used in CommandUsageTemplateData.
type i18nCommandGlossary struct {
	SectionUsage                string
	SectionAliases              string
	SectionExamples             string
	SectionAvailableCommands    string
	SectionAdditionalCommands   string
	SectionFlags                string
	SectionGlobalFlags          string
	SectionAdditionalHelpTopics string
	Use                         string
	ForInfoAboutCommand         string
}

var commonCommandGlossary *i18nCommandGlossary

func getCommandGlossary() *i18nCommandGlossary {
	if commonCommandGlossary == nil {
		commonCommandGlossary = &i18nCommandGlossary{
			SectionUsage:                gotext.Get("SectionUsage"),
			SectionAliases:              gotext.Get("SectionAliases"),
			SectionExamples:             gotext.Get("SectionExamples"),
			SectionAvailableCommands:    gotext.Get("SectionAvailableCommands"),
			SectionAdditionalCommands:   gotext.Get("SectionAdditionalCommands"),
			SectionFlags:                gotext.Get("SectionFlags"),
			SectionGlobalFlags:          gotext.Get("SectionGlobalFlags"),
			SectionAdditionalHelpTopics: gotext.Get("SectionAdditionalHelpTopics"),
			Use:                         gotext.Get("Use"),
			ForInfoAboutCommand:         gotext.Get("ForInfoAboutCommand"),
		}
	}
	return commonCommandGlossary
}

func setupLocalization() {
	for _, localeIdentifier := range detectLangs() {
		locale := gotext.NewLocale("", localeIdentifier)

		allDomainsFound := true
		for _, domain := range availableLocalizationDomains {

			//localeFilepath := fmt.Sprintf("locales/%s/%s.po", domain, localeIdentifier)
			localeFilepath := fmt.Sprintf("locales/%s/%s.mo", domain, localeIdentifier)
			localeFile, err := localeFS.ReadFile(localeFilepath)
			if err != nil {
				allDomainsFound = false
				break
			}

			//translator := gotext.NewPo()
			translator := gotext.NewMo()
			translator.Parse(localeFile)

			locale.AddTranslator(domain, translator)
		}

		if !allDomainsFound {
			continue
		}

		gotext.SetStorage(locale)
		break
	}
}

func detectLangs() []string {
	var detectedLangs []string

	// From environment
	for _, envKey := range envVariablesHoldingLocale {
		lang := os.Getenv(envKey)
		if lang != "" {
			detectedLang := language.Make(lang)
			appendLang(&detectedLangs, detectedLang)
		}
	}

	// Lastly, from defaults
	appendLang(&detectedLangs, defaultLanguage)

	return detectedLangs
}

func appendLang(langs *[]string, lang language.Tag) {
	if lang.IsRoot() {
		return
	}

	langString := lang.String()
	*langs = append(*langs, langString)

	langBase, confidentInBase := lang.Base()
	if confidentInBase != language.No {
		*langs = append(*langs, langBase.ISO3())
		*langs = append(*langs, langBase.String())
	}
}

func init() {
	setupLocalization()
}
