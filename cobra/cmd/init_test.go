package cmd

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func TestGoldenInitCmd(t *testing.T) {
	wd, _ := os.Getwd()
	project := &Project{
		AbsolutePath: wd,
		PkgName:      "github.com/spf13/testproject",
		Legal:        getLicense(),
		Viper:        true,
		AppName:      "testproject",
	}

	err := project.Create()
	if err != nil {
		t.Fatal(err)
	}

	//expectedFiles := []string{"LICENSE", "main.go", "cmd/root.go"}
	//for _, f := range expectedFiles {
	//	// read each file and compare with corresponding golden file
	//}
}

// TestGoldenInitCmd initializes the project "github.com/spf13/testproject"
// in GOPATH and compares the content of files in initialized project with
// appropriate golden files ("testdata/*.golden").
// Use -update to update existing golden files.
func TTestGoldenInitCmd(t *testing.T) {
	projectName := "github.com/spf13/testproject"
	project := NewProject(projectName)
	defer os.RemoveAll(project.AbsPath())

	viper.Set("author", "NAME HERE <EMAIL ADDRESS>")
	viper.Set("license", "apache")
	viper.Set("year", 2017)
	defer viper.Set("author", nil)
	defer viper.Set("license", nil)
	defer viper.Set("year", nil)

	os.Args = []string{"cobra", "init", projectName}
	if err := rootCmd.Execute(); err != nil {
		t.Fatal("Error by execution:", err)
	}

	expectedFiles := []string{".", "cmd", "LICENSE", "main.go", "cmd/root.go"}
	gotFiles := []string{}

	// Check project file hierarchy and compare the content of every single file
	// with appropriate golden file.
	err := filepath.Walk(project.AbsPath(), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Make path relative to project.AbsPath().
		// E.g. path = "/home/user/go/src/github.com/spf13/testproject/cmd/root.go"
		// then it returns just "cmd/root.go".
		relPath, err := filepath.Rel(project.AbsPath(), path)
		if err != nil {
			return err
		}
		relPath = filepath.ToSlash(relPath)
		gotFiles = append(gotFiles, relPath)
		goldenPath := filepath.Join("testdata", filepath.Base(path)+".golden")

		switch relPath {
		// Known directories.
		case ".", "cmd":
			return nil
		// Known files.
		case "LICENSE", "main.go", "cmd/root.go":
			if *update {
				got, err := ioutil.ReadFile(path)
				if err != nil {
					return err
				}
				if err := ioutil.WriteFile(goldenPath, got, 0644); err != nil {
					t.Fatal("Error while updating file:", err)
				}
			}
			return compareFiles(path, goldenPath)
		}
		// Unknown file.
		return errors.New("unknown file: " + path)
	})
	if err != nil {
		t.Fatal(err)
	}

	// Check if some files lack.
	if err := checkLackFiles(expectedFiles, gotFiles); err != nil {
		t.Fatal(err)
	}
}
