package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

func TestConfigFileDefaultFormat(t *testing.T) {
	var (
		defaultConfigName        = ".cobra"
		configFileTestSuffix     = ".test.bak"
		testKey                  = "aKey"
		testValue0               = "aValue0"
		testValue1               = "aValue1"
		defaultConfigFileContent = []byte(fmt.Sprintf(
			"---\n%s:\n  - %s\n  - %s\n",
			testKey, testValue0, testValue1,
		))
		home string
		err  error
	)

	if home, err = homedir.Dir(); err != nil {
		t.Fatalf("Error discovering home directory: %v", err)
	}

	configFilePath := filepath.Join(home, defaultConfigName)

	// if file already exists, move to backup
	if fileInfo, _ := os.Stat(configFilePath); fileInfo != nil {
		configFilePathBackup := configFilePath + configFileTestSuffix
		os.Remove(configFilePathBackup)
		if err = os.Rename(configFilePath, configFilePathBackup); err != nil {
			t.Fatalf("Error backing up existing config file '%v' :'%v'", configFilePath, err)
		}
		// defer moving backup back to original
		defer func() {
			os.Remove(configFilePath)
			os.Rename(configFilePathBackup, configFilePath)
		}()
	}

	// create test file
	os.Remove(configFilePath)
	ioutil.WriteFile(configFilePath, defaultConfigFileContent, 0666)
	if _, err = os.Stat(configFilePath); os.IsNotExist(err) {
		t.Fatalf("Expected to find file '%s', but got '%v'", configFilePath, err)
	}

	// attempt to use config file
	initConfig()

	result := viper.GetStringSlice(testKey)
	expected := []string{testValue0, testValue1}
	if len(result) != len(expected) {
		t.Fatalf("Expected '%T' with %d elements, but got '%v'", expected, len(expected), result)
	}
	for i := range result {
		if result[i] != expected[i] {
			t.Fatalf("Expected '%v', but got '%v'", expected, result)
		}
	}

	// delete test file
	os.Remove(configFilePath)
}
