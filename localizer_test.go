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
	"github.com/leonelquinteros/gotext"
	"os"
	"testing"
)

// resetLocalization resets to the vendor defaults
// Ideally this would be done using gotext.SetStorage(nil)
func resetLocalization() {
	locale := gotext.NewLocale("/usr/local/share/locale", "en_US")
	locale.AddDomain("default")
	locale.SetDomain("default")
	gotext.SetStorage(locale)
}

func TestLocalization(t *testing.T) {
	tests := []struct {
		rule                string
		env                 map[string]string
		expectedLanguage    string
		message             string
		expectedTranslation string
	}{
		{
			rule:             "default language is english",
			expectedLanguage: "en",
		},
		{
			rule: "section example (en)",
			env: map[string]string{
				"LANGUAGE": "en",
			},
			expectedLanguage:    "en",
			message:             "SectionExamples",
			expectedTranslation: "Examples",
		},
		{
			rule: "section example (fr)",
			env: map[string]string{
				"LANGUAGE": "fr",
			},
			expectedLanguage:    "fr",
			message:             "SectionExamples",
			expectedTranslation: "Exemples",
		},
		{
			rule:                "untranslated string stays as-is",
			message:             "AtelophobiacCoder",
			expectedTranslation: "AtelophobiacCoder",
		},
		{
			rule: "fr_FR falls back to fr",
			env: map[string]string{
				"LANGUAGE": "fr_FR",
			},
			expectedLanguage: "fr",
		},
		{
			rule: "fr-FR falls back to fr",
			env: map[string]string{
				"LANGUAGE": "fr-FR",
			},
			expectedLanguage: "fr",
		},
		{
			rule: "fr_FR@UTF-8 falls back to fr",
			env: map[string]string{
				"LANGUAGE": "fr_FR@UTF-8",
			},
			expectedLanguage: "fr",
		},
		{
			rule: "fr_FR.UTF-8 falls back to fr",
			env: map[string]string{
				"LANGUAGE": "fr_FR.UTF-8",
			},
			expectedLanguage: "fr",
		},
		{
			rule: "LANGUAGE > LC_ALL",
			env: map[string]string{
				"LANGUAGE":    "fr",
				"LC_ALL":      "en",
				"LC_MESSAGES": "en",
				"LANG":        "en",
			},
			expectedLanguage: "fr",
		},
		{
			rule: "LC_ALL > LC_MESSAGES",
			env: map[string]string{
				"LC_ALL":      "fr",
				"LC_MESSAGES": "en",
				"LANG":        "en",
			},
			expectedLanguage: "fr",
		},
		{
			rule: "LC_MESSAGES > LANG",
			env: map[string]string{
				"LC_MESSAGES": "fr",
				"LANG":        "en",
			},
			expectedLanguage: "fr",
		},
		{
			rule: "LANG is supported",
			env: map[string]string{
				"LANG": "fr",
			},
			expectedLanguage: "fr",
		},
		{
			rule: "Fall back to another env if a language is not supported",
			env: map[string]string{
				"LANGUAGE": "xx",
				"LC_ALL":   "fr",
			},
			expectedLanguage: "fr",
		},
	}
	for _, tt := range tests {
		t.Run(tt.rule, func(t *testing.T) {
			// I. Prepare the environment
			os.Clearenv()
			if tt.env != nil {
				for envKey, envValue := range tt.env {
					err := os.Setenv(envKey, envValue)
					if err != nil {
						t.Errorf("os.Setenv() failed for %s=%s", envKey, envValue)
						return
					}
				}
			}

			// II. Run the initialization of localization
			resetLocalization()
			setupLocalization()

			// III. Assert that language was detected correctly
			if tt.expectedLanguage != "" {
				actualLanguage := gotext.GetLanguage()
				if actualLanguage != tt.expectedLanguage {
					t.Errorf("Expected language `%v' but got `%v'.", tt.expectedLanguage, actualLanguage)
					return
				}
			}

			// IV. Assert that the message was translated adequately
			if tt.message != "" {
				actualTranslation := gotext.Get(tt.message)
				if actualTranslation != tt.expectedTranslation {
					t.Errorf("Expected translation `%v' but got `%v'.", tt.expectedTranslation, actualTranslation)
					return
				}
			}
		})
	}
}
