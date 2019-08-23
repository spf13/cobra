package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
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

func TestInitNoLicense(t *testing.T) {
	project := getProject()
	project.Legal = noLicense
	defer os.RemoveAll(project.AbsolutePath)

	err := project.Create()
	if err != nil {
		t.Fatal(err)
	}

	root := project.AbsolutePath

	want := []string{"main.go", "cmd/root.go"}
	var got []string
	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		relpath, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		got = append(got, relpath)
		return nil
	})
	if err != nil {
		t.Fatalf("walking path %q: %v", root, err)
	}
	sort.StringSlice(got).Sort()
	sort.StringSlice(want).Sort()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf(
			"In %s, found:\n  %s\nwant:\n  %s",
			root,
			strings.Join(got, ", "),
			strings.Join(want, ", "),
		)
	}
}
