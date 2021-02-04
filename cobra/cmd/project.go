package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/spf13/cobra/cobra/tpl"
)

// Project contains name, license, descriptions, and paths to projects.
type Project struct {
	// v2
	PkgName      string
	PkgShortDesc string
	PkgLongDesc  string
	Copyright    string
	AbsolutePath string
	Legal        License
	Viper        bool
	AppName      string
}

// Command contains name, parent command, and descriptions for commands
type Command struct {
	CmdName      string
	CmdParent    string
	CmdShortDesc string
	CmdLongDesc  string
	*Project
}

// Create will create a new project
func (p *Project) Create() error {
	// check if AbsolutePath exists
	if _, err := os.Stat(p.AbsolutePath); os.IsNotExist(err) {
		// create directory
		if err := os.Mkdir(p.AbsolutePath, 0754); err != nil {
			return err
		}
	}

	// create main.go
	mainFile, err := os.Create(fmt.Sprintf("%s/main.go", p.AbsolutePath))
	if err != nil {
		return err
	}
	defer mainFile.Close()

	mainTemplate := template.Must(template.New("main").Parse(string(tpl.MainTemplate())))
	err = mainTemplate.Execute(mainFile, p)
	if err != nil {
		return err
	}

	// create cmd/root.go
	if _, err = os.Stat(fmt.Sprintf("%s/cmd", p.AbsolutePath)); os.IsNotExist(err) {
		os.Mkdir(fmt.Sprintf("%s/cmd", p.AbsolutePath), 0751)
	}
	rootFile, err := os.Create(fmt.Sprintf("%s/cmd/root.go", p.AbsolutePath))
	if err != nil {
		return err
	}
	defer rootFile.Close()

	rootTemplate := template.Must(template.New("root").Parse(string(tpl.RootTemplate())))
	err = rootTemplate.Execute(rootFile, p)
	if err != nil {
		return err
	}

	// create license
	return p.createLicenseFile()
}

func (p *Project) createLicenseFile() error {
	data := map[string]interface{}{
		"copyright": copyrightLine(),
	}
	licenseFile, err := os.Create(fmt.Sprintf("%s/LICENSE", p.AbsolutePath))
	if err != nil {
		return err
	}
	defer licenseFile.Close()

	licenseTemplate := template.Must(template.New("license").Parse(p.Legal.Text))
	return licenseTemplate.Execute(licenseFile, data)
}

// Create will create a new command for a project
func (c *Command) Create() error {
	cmdFileName := c.CmdName
	if c.CmdParent != "rootCmd" {

		cmdFileName = fmt.Sprintf("%s_%s", strings.Replace(validateCmdName(c.CmdParent), "Cmd", "", -1), c.CmdName)
	}
	cmdFile, err := os.Create(fmt.Sprintf("%s/cmd/%s.go", c.AbsolutePath, cmdFileName))
	if err != nil {
		return err
	}
	defer cmdFile.Close()

	commandTemplate := template.Must(template.New("sub").Parse(string(tpl.AddCommandTemplate())))
	err = commandTemplate.Execute(cmdFile, c)
	if err != nil {
		return err
	}
	return nil
}
