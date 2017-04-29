package cmd

import (
	"os"
	"path/filepath"
	"strings"
)

type Project struct {
	absPath string
	cmdDir  string
	srcPath string
	license License
	name    string
}

func NewProject(projectName string) *Project {
	if projectName == "" {
		return nil
	}

	p := new(Project)
	p.name = projectName

	// 1. Find already created protect.
	p.absPath = findPackage(projectName)

	// 2. If there are no created project with this path and user in GOPATH,
	// then use GOPATH+projectName.
	if p.absPath == "" {
		wd, err := os.Getwd()
		if err != nil {
			er(err)
		}
		for _, goPath := range goPaths {
			if filepath.HasPrefix(wd, goPath) {
				p.absPath = filepath.Join(goPath, projectName)
				break
			}
		}
	}

	// 3. If user is not in GOPATH, then use (first GOPATH)+projectName.
	if p.absPath == "" {
		p.absPath = filepath.Join(srcPaths[0], projectName)
	}

	return p
}

// findPackage returns full path to go package. It supports multiple GOPATHs.
// findPackage returns "", if it can't find path.
// If packageName is "", findPackage returns "" too.
//
// For example, package "github.com/spf13/hugo"
// is located in /home/user/go/src/github.com/spf13/hugo,
// then `findPackage("github.com/spf13/hugo")`
// will return "/home/user/go/src/github.com/spf13/hugo"
func findPackage(packageName string) string {
	if packageName == "" {
		return ""
	}

	for _, srcPath := range srcPaths {
		packagePath := filepath.Join(srcPath, packageName)
		if exists(packagePath) {
			return packagePath
		}
	}

	return ""
}

func NewProjectFromPath(absPath string) *Project {
	if absPath == "" || !filepath.IsAbs(absPath) {
		return nil
	}

	p := new(Project)
	p.absPath = absPath
	p.absPath = strings.TrimSuffix(p.absPath, p.CmdDir())
	p.name = filepath.ToSlash(trimSrcPath(p.absPath, p.SrcPath()))
	return p
}

func trimSrcPath(absPath, srcPath string) string {
	relPath, err := filepath.Rel(srcPath, absPath)
	if err != nil {
		er("Cobra only supports project within $GOPATH")
	}
	return relPath
}

func (p *Project) License() License {
	if p.license.Text == "" { // check if license is not blank
		p.license = getLicense()
	}

	return p.license
}

func (p Project) Name() string {
	return p.name
}

func (p *Project) CmdDir() string {
	if p.absPath == "" {
		return ""
	}
	if p.cmdDir == "" {
		p.cmdDir = findCmdDir(p.absPath)
	}
	return p.cmdDir
}

func findCmdDir(absPath string) string {
	if !exists(absPath) || isEmpty(absPath) {
		return "cmd"
	}

	files, _ := filepath.Glob(filepath.Join(absPath, "c*"))
	for _, f := range files {
		for _, c := range cmdDirs {
			if f == c {
				return c
			}
		}
	}

	return "cmd"
}

func (p Project) AbsPath() string {
	return p.absPath
}

func (p *Project) SrcPath() string {
	if p.srcPath != "" {
		return p.srcPath
	}
	if p.absPath == "" {
		p.srcPath = srcPaths[0]
		return p.srcPath
	}

	for _, srcPath := range srcPaths {
		if strings.HasPrefix(p.absPath, srcPath) {
			p.srcPath = srcPath
			break
		}
	}

	return p.srcPath
}
