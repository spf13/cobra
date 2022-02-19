package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"text/template"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/cobra/tpl"
)

// Project contains name, license and paths to projects.
type Project struct {
	// v2
	PkgName      string
	Copyright    string
	AbsolutePath string
	Legal        License
	Viper        bool
	AppName      string
}

type Command struct {
	CmdName   string
	CmdParent string
	*Project
}

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
		cobra.CheckErr(os.Mkdir(fmt.Sprintf("%s/cmd", p.AbsolutePath), 0751))
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
	if p.Legal.Name != "None" {
		return p.createLicenseFile()
	}
	return nil
}

func (p *Project) createLicenseFile() error {
	data := map[string]interface{}{
		"copyright": copyrightLine(),
	}
	licensesExist := []string{}
	err := filepath.Walk(p.AbsolutePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && filepath.Ext(path) != ".txt" && filepath.Ext(path) != ".md" && filepath.Ext(path) != "" {
			return nil
		}
		reg := regexp.MustCompile(`(?i).*license\.?.*`)
		if reg.MatchString(info.Name()) {
			licensesExist = append(licensesExist, info.Name())
		}
		return nil
	})
	if err != nil {
		return err
	}
	if len(licensesExist) > 0 {
		fmt.Println("Licenses already exist in the project")
		fmt.Println("Licenses found:")
		for _, license := range licensesExist {
			fmt.Printf("  %s\n", license)
		}
		fmt.Print("Would you like still to add a license? [Y/n] ")
		var answer string
		fmt.Scanln(&answer)
		if !(answer == "y" || answer == "Y") {
			return nil
		}
		licenseFound := false
		for _, license := range licensesExist {
			if license == "LICENSE" {
				licenseFound = true
			}
		}
		if licenseFound {
			fmt.Print("LICENSE exists. Would you like to overwrite it? [Y/n] ")
			fmt.Scanln(&answer)
			if !(answer == "y" || answer == "Y") {
				return nil
			}
		}
	}
	licenseFile, err := os.Create(fmt.Sprintf("%s/LICENSE", p.AbsolutePath))
	if err != nil {
		return err
	}
	defer licenseFile.Close()

	licenseTemplate := template.Must(template.New("license").Parse(p.Legal.Text))
	return licenseTemplate.Execute(licenseFile, data)
}

func (c *Command) Create() error {
	cmdFile, err := os.Create(fmt.Sprintf("%s/cmd/%s.go", c.AbsolutePath, c.CmdName))
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
