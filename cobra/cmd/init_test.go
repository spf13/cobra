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
		PkgName:      "github.com/spf13/testproject",
		Legal:        getLicense(),
		Copyright:    copyrightLine(),
		Viper:        true,
		AppName:      "testproject",
	}

	err := project.Create()
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if _, err := os.Stat(project.AbsolutePath); err == nil {
			os.RemoveAll(project.AbsolutePath)
		}
	}()

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
