package cobra

import (
	"embed"
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

// localeFS points to an embedded filesystem of binary gettext translation files.
// For performance and smaller builds, only the binary MO files are included.
// Their sibling PO files should still be considered their authoritative source.
//
//go:embed locales/*/*.mo
var localeFS embed.FS

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
