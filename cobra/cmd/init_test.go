package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func getProject() *Project {
	return &Project{
		AbsolutePath: fmt.Sprintf("%s/testproject", mustTempDir()),
		Legal:        getLicense(),
		Copyright:    copyrightLine(),
		AppName:      "testproject",
		PkgName:      "github.com/spf13/testproject",
		Viper:        true,
	}
}

func mustTempDir() string {
	dir, err := ioutil.TempDir("", "cobra_cli_test_")
	if err != nil {
		panic(err)
	}
	return dir
}

func TestGoldenInitCmd(t *testing.T) {
	project := getProject()
	defer os.RemoveAll(project.AbsolutePath)

	if err := project.Create(); err != nil {
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
