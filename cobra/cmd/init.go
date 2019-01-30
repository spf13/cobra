// Copyright © 2015 Steve Francia <spf@spf13.com>.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"path"
)

var (
	pkgName string

	initCmd = &cobra.Command{
		Use:     "init [name]",
		Aliases: []string{"initialize", "initialise", "create"},
		Short:   "Initialize a Cobra Application",
		Long: `Initialize (cobra init) will create a new application, with a license
and the appropriate structure for a Cobra-based CLI application.

  * If a name is provided, it will be created in the current directory;
  * If no name is provided, the current directory will be assumed;
  * If a relative path is provided, it will be created inside $GOPATH
    (e.g. github.com/spf13/hugo);
  * If an absolute path is provided, it will be created;
  * If the directory already exists but is empty, it will be used.

Init will not use an existing directory with contents.`,

		Run: func(cmd *cobra.Command, args []string) {

			wd, err := os.Getwd()
			if err != nil {
				er(err)
			}

			if len(args) > 0 {
				if args[0] != "." {
					wd = fmt.Sprintf("%s/%s", wd, args[0])
				}
			}

			project := &Project{
				AbsolutePath: wd,
				PkgName:      pkgName,
				Legal:        getLicense(),
				Copyright:    copyrightLine(),
				Viper:        viper.GetBool("useViper"),
				AppName:      path.Base(pkgName),
			}

			if err := project.Create(); err != nil {
				er(err)
			}

			fmt.Printf("Your Cobra applicaton is ready at\n%s\n", project.AbsolutePath)
		},
	}
)

func init() {
	initCmd.Flags().StringVar(&pkgName, "pkg-name", "", "fully qualified pkg name")
	initCmd.MarkFlagRequired("pkg-name")
}
