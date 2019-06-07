package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func getProject() *Project {
	wd, _ := os.Getwd()
	return &Project{
		AbsolutePath: fmt.Sprintf("%s/testproject", wd),
		Legal:        getLicense(),
		Copyright:    copyrightLine(),
		AppName:      "testproject",
		PkgName:      "github.com/spf13/testproject",
		Viper:        true,
	}
}

func TestGoldenInitCmd(t *testing.T) {

	dir, err := ioutil.TempDir("", "cobra-init")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	tests := []struct {
		name      string
		args      []string
		pkgName   string
		expectErr bool
	}{
		{
			name:      "successfully creates a project with name",
			args:      []string{"testproject"},
			pkgName:   "github.com/spf13/testproject",
			expectErr: false,
		},
		{
			name:      "returns error when passing an absolute path for project",
			args:      []string{dir},
			pkgName:   "github.com/spf13/testproject",
			expectErr: true,
		},
		{
			name:      "returns error when passing an relative path for project",
			args:      []string{"github.com/spf13/testproject"},
			pkgName:   "github.com/spf13/testproject",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			assertNoErr(t, initCmd.Flags().Set("pkg-name", tt.pkgName))
			viper.Set("useViper", true)
			projectPath, err := initializeProject(tt.args)
			defer func() {
				if projectPath != "" {
					os.RemoveAll(projectPath)
				}
			}()

			if !tt.expectErr && err != nil {
				t.Fatalf("did not expect an error, got %s", err)
			}
			if tt.expectErr {
				if err == nil {
					t.Fatal("expected an error but got none")
				} else {
					// got an expected error nothing more to do
					return
				}
			}

			expectedFiles := []string{"LICENSE", "main.go", "cmd/root.go"}
			for _, f := range expectedFiles {
				generatedFile := fmt.Sprintf("%s/%s", projectPath, f)
				goldenFile := fmt.Sprintf("testdata/%s.golden", filepath.Base(f))
				err := compareFiles(generatedFile, goldenFile)
				if err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}
