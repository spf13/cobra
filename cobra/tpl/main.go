package tpl

func MainTemplate() []byte {
	return []byte(`
/*
{{ .Copyright }}
{{ if .Legal.Header }}{{ .Legal.Header }}{{ end }}
*/
package main

import "{{ .PkgName }}/cmd"

func main() {
  cmd.Execute()
}
`)
}

func RootTemplate() []byte {
	return []byte(`
/*
{{ .Copyright }}
{{ if .Legal.Header }}{{ .Legal.Header }}{{ end }}
*/
package cmd

import (
  "fmt"
  "os"
  "github.com/spf13/cobra"
{{ if .Viper }}
  homedir "github.com/mitchellh/go-homedir"
  "github.com/spf13/viper"
{{ end }}
)

{{ if .Viper }}
var cfgFile string
{{ end }}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
  Use:   "{{ .AppName }}",
  Short: "A brief description of your application",
  Long: ` + "`" + `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.` + "`" + `,
  // Uncomment the following line if your bare application
  // has an action associated with it:
  //	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
  if err := rootCmd.Execute(); err != nil {
    fmt.Println(err)
    os.Exit(1)
  }
}

func init() {
{{- if .Viper }}
  cobra.OnInitialize(initConfig)
{{ end }}
  // Here you will define your flags and configuration settings.
  // Cobra supports persistent flags, which, if defined here,
  // will be global for your application.
{{ if .Viper }}
  rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.{{ .AppName }}.yaml)")
{{ else }}
  // rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.{{ .AppName }}.yaml)")
{{ end }}

  // Cobra also supports local flags, which will only run
  // when this action is called directly.
  rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

{{ if .Viper }}
// initConfig reads in config file and ENV variables if set.
func initConfig() {
  if cfgFile != "" {
    // Use config file from the flag.
    viper.SetConfigFile(cfgFile)
  } else {
    // Find home directory.
    home, err := homedir.Dir()
    if err != nil {
      fmt.Println(err)
      os.Exit(1)
    }

    // Search config in home directory with name ".{{ .AppName }}" (without extension).
    viper.AddConfigPath(home)
    viper.SetConfigName(".{{ .AppName }}")
  }

  viper.AutomaticEnv() // read in environment variables that match

  // If a config file is found, read it in.
  if err := viper.ReadInConfig(); err == nil {
    fmt.Println("Using config file:", viper.ConfigFileUsed())
  }
}
{{ end }}
`)
}
