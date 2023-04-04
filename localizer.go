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
			SectionUsage:                i18nSectionUsage(),
			SectionAliases:              i18nSectionAliases(),
			SectionExamples:             i18nSectionExamples(),
			SectionAvailableCommands:    i18nSectionAvailableCommands(),
			SectionAdditionalCommands:   i18nSectionAdditionalCommands(),
			SectionFlags:                i18nSectionFlags(),
			SectionGlobalFlags:          i18nSectionGlobalFlags(),
			SectionAdditionalHelpTopics: i18nSectionAdditionalHelpTopics(),
			Use:                         i18nUse(),
			ForInfoAboutCommand:         i18nForInfoAboutCommand(),
		}
	}
	return commonCommandGlossary
}

func i18nSectionUsage() string {
	return localizeMessage(&i18n.Message{
		ID:          "SectionUsage",
		Description: "title of the section in the usage template",
		Other:       "Usage",
	})
}

func i18nSectionAliases() string {
	return localizeMessage(&i18n.Message{
		ID:          "SectionAliases",
		Description: "title of the section in the usage template",
		Other:       "Aliases",
	})
}

func i18nSectionExamples() string {
	return localizeMessage(&i18n.Message{
		ID:          "SectionExamples",
		Description: "title of the section in the usage template",
		Other:       "Examples",
	})
}

func i18nSectionAvailableCommands() string {
	return localizeMessage(&i18n.Message{
		ID:          "SectionAvailableCommands",
		Description: "title of the section in the usage template",
		Other:       "Available Commands",
	})
}

func i18nSectionAdditionalCommands() string {
	return localizeMessage(&i18n.Message{
		ID:          "SectionAdditionalCommands",
		Description: "title of the section in the usage template",
		Other:       "Additional Commands",
	})
}

func i18nSectionFlags() string {
	return localizeMessage(&i18n.Message{
		ID:          "SectionFlags",
		Description: "title of the section in the usage template",
		Other:       "Flags",
	})
}

func i18nSectionGlobalFlags() string {
	return localizeMessage(&i18n.Message{
		ID:          "SectionGlobalFlags",
		Description: "title of the section in the usage template",
		Other:       "Global Flags",
	})
}

func i18nSectionAdditionalHelpTopics() string {
	return localizeMessage(&i18n.Message{
		ID:          "SectionAdditionalHelpTopics",
		Description: "title of the section in the usage template",
		Other:       "Additional Help Topics",
	})
}

func i18nUse() string {
	return localizeMessage(&i18n.Message{
		ID:          "Use",
		Description: "beginning of a sentence like 'Use <this> to do <that>'",
		Other:       "Use",
	})
}

func i18nForInfoAboutCommand() string {
	return localizeMessage(&i18n.Message{
		ID:          "ForInfoAboutCommand",
		Description: "end of a sentence",
		Other:       "for more information about a command",
	})
}

func i18nError() string {
	return localizeMessage(&i18n.Message{
		ID:          "Error",
		Description: "prefix of error messages",
		Other:       "Error",
	})
}

func i18nLegacyArgsValidationError() string {
	return localizeMessage(&i18n.Message{
		ID:          "LegacyArgsValidationError",
		Description: "error shown when args are not understood (subcmd, cmd, suggestion)",
		Other:       "unknown command %q for %q%s",
	})
}

func i18nNoArgsValidationError() string {
	return localizeMessage(&i18n.Message{
		ID:          "NoArgsValidationError",
		Description: "error shown when args are present but should not (subcmd, cmd)",
		Other:       "unknown command %q for %q",
	})
}

func i18nOnlyValidArgsValidationError() string {
	return localizeMessage(&i18n.Message{
		ID:          "OnlyValidArgsValidationError",
		Description: "error shown when arg is invalid (arg, cmd, suggestion)",
		Other:       "invalid argument %q for %q%s",
	})
}

func i18nMinimumNArgsValidationError(amountRequired int) string {
	return localizeMessageWithPlural(&i18n.Message{
		ID:          "MinimumNArgsValidationError",
		Description: "error shown when arg count is too low (expected amount, actual amount)",
		Other:       "requires at least %d args, only received %d",
		One:         "requires at least %d arg, only received %d",
	}, amountRequired)
}

func i18nMaximumNArgsValidationError(amountRequired int) string {
	return localizeMessageWithPlural(&i18n.Message{
		ID:          "MaximumNArgsValidationError",
		Description: "error shown when arg count is too low (expected amount, actual amount)",
		Other:       "accepts at most %d args, received %d",
		One:         "accepts at most %d arg, received %d",
	}, amountRequired)
}

func i18nExactArgsValidationError(amountRequired int) string {
	return localizeMessageWithPlural(&i18n.Message{
		ID:          "ExactArgsValidationError",
		Description: "error shown when arg count is not exact (expected amount, actual amount)",
		Other:       "accepts %d args, received %d",
		One:         "accepts %d arg, received %d",
	}, amountRequired)
}

func i18nRangeArgsValidationError(amountMax int) string {
	return localizeMessageWithPlural(&i18n.Message{
		ID:          "RangeArgsValidationError",
		Description: "error shown when arg count is not in range (expected min, expected max, actual amount)",
		Other:       "accepts between %d and %d args, received %d",
		One:         "accepts between %d and %d arg, received %d",
	}, amountMax)
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

func localizeMessageWithPlural(message *i18n.Message, pluralCount int) string {
	localizedValue, err := localizer.Localize(&i18n.LocalizeConfig{
		PluralCount:    pluralCount,
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

func setupLocalizer() {
	bundle := i18n.NewBundle(defaultLanguage)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	detectedLangs := detectLangs()
	//fmt.Println("Detected languages", detectedLangs)
	loadTranslationFiles(bundle, detectedLangs)
	localizer = i18n.NewLocalizer(bundle, detectedLangs...)
}

func init() {
	setupLocalizer() // FIXME: perhaps hook this somewhere else?  (not init)
}
