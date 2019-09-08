package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestGoldenInitCmd(t *testing.T) {

	wd, _ := os.Getwd()
	project := &Project{
		AbsolutePath: fmt.Sprintf("%s/testproject", wd),
		Legal:        getLicense(),
		Copyright:    copyrightLine(),

		// required to init
		AppName: "testproject",
		PkgName: "github.com/spf13/testproject",
		Viper:   true,
	}
	defer os.RemoveAll(project.AbsolutePath)

	// init project first
	err := project.Create()
	if err != nil {
		t.Fatal(err)
	}

	expectedFiles := []string{"LICENSE", "main.go", "cmd/root.go"}
	for _, f := range expectedFiles {
		generatedFile := fmt.Sprintf("%s/%s", project.AbsolutePath, f)
		goldenFile := fmt.Sprintf("testdata/%s.golden", filepath.Base(f))
		err := compareFiles(generatedFile, goldenFile)
		if err != nil {
			t.Fatal(err)
		}
	}
}
