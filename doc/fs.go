package doc

import "github.com/spf13/afero"

// This can be used by any code in current package that need to access filesystem
// It helps softwares using cobra's documentation generation to have better tests
var docFs *afero.Afero

func init() {
	// By default, make the docFS a OS one, allowing real FS operations
	docFs = &afero.Afero{Fs: afero.NewOsFs()}
}

// GetFS returns the current Afero FS used by the package
// Useful to backup the current FS, before setting a temp FS with SetFS
func GetFS() *afero.Afero {
	return docFs
}

// SetFS defines the Afero FS that will be used by the package
func SetFS(fs *afero.Afero) {
	docFs = fs
}
