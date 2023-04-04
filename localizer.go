package cobra

import (
	"embed"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"os"
)

var defaultLanguage = language.English

// envVariablesHoldingLocale is sorted by decreasing priority
// These environment variables are expected to hold a parsable locale (fr_FR, es, en-US, …)
var envVariablesHoldingLocale = []string{
	"LANGUAGE",
	"LC_ALL",
	"LANG",
}

// localeFS points to an embedded filesystem of TOML translation files
//
//go:embed translations/*.toml
var localeFS embed.FS

// Localizer can be used to fetch localized messages
var localizer *i18n.Localizer

func i18nError() string {
	return localizeMessage(&i18n.Message{
		ID:          "Error",
		Description: "prefix of error messages",
		Other:       "Error",
	})
}

func i18nRunHelpTip() string {
	return localizeMessage(&i18n.Message{
		ID:          "RunHelpTip",
		Description: "tip shown when a command fails (command path)",
		Other:       "Run '%v --help' for usage.",
	})
}

func i18nExclusiveFlagsValidationError() string {
	return localizeMessage(&i18n.Message{
		ID:          "ExclusiveFlagsValidationError",
		Description: "error shown when multiple exclusive flags are provided (group flags, offending flags)",
		Other:       "if any flags in the group [%v] are set none of the others can be; %v were all set",
	})
}

// … lots more translations here

func localizeMessage(message *i18n.Message) string {
	localizedValue, err := localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: message,
	})
	if err != nil {
		return message.Other
	}

	return localizedValue
}

func loadTranslationFiles(bundle *i18n.Bundle, langs []string) {
	for _, lang := range langs {
		_, _ = bundle.LoadMessageFileFS(localeFS, fmt.Sprintf("translations/active.%s.toml", lang))
	}
}

func detectLangs() []string {
	var detectedLangs []string
	for _, envKey := range envVariablesHoldingLocale {
		lang := os.Getenv(envKey)
		if lang != "" {
			detectedLang := language.Make(lang)
			appendLang(&detectedLangs, detectedLang)
		}
	}
	appendLang(&detectedLangs, defaultLanguage)

	return detectedLangs
}

func appendLang(langs *[]string, lang language.Tag) {
	langString := lang.String()
	langBase, _ := lang.Base()
	*langs = append(*langs, langString)
	*langs = append(*langs, langBase.ISO3())
	*langs = append(*langs, langBase.String())
}

func init() {
	bundle := i18n.NewBundle(defaultLanguage)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	detectedLangs := detectLangs()
	//fmt.Println("Detected languages", detectedLangs)
	loadTranslationFiles(bundle, detectedLangs)
	localizer = i18n.NewLocalizer(bundle, detectedLangs...)
}
