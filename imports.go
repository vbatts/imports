/*
  go build imports.go
  ./import -r ./src/docker/ | less
*/
package main

import (
	"flag"
	"fmt"
	"go/build"
	"log"
	"os"
	"path/filepath"
	"text/template"
)

var (
	DefaultPath  string = "."
	ShouldWalk   bool   = false
	PathsToCheck []Path = []Path{}
)

func main() {
	flag.BoolVar(&ShouldWalk, "r", ShouldWalk, "whether to recurse/walk the file system for packages")
	flag.Parse()

	if flag.NArg() > 0 {
		for _, arg := range flag.Args() {
			if IsDir(arg) {
				if a_path, err := filepath.Abs(arg); err == nil {
					PathsToCheck = append(PathsToCheck, Path{Base: a_path})
				} else {
					fmt.Fprintln(os.Stderr, err)
				}
			} else {
				if pkg, err := build.Default.Import(arg, "", build.FindOnly); err == nil {
					PathsToCheck = append(PathsToCheck, Path{Base: pkg.SrcRoot, Rel: pkg.ImportPath})
				} else {
					fmt.Fprintln(os.Stderr, err)
				}
			}
		}
	} else {
		if a_path, err := filepath.Abs(DefaultPath); err == nil {
			PathsToCheck = append(PathsToCheck, Path{Base: a_path})
		} else {
			fmt.Fprintln(os.Stderr, err)
		}
	}

	if ShouldWalk {
		paths := []Path{}
		for _, path := range PathsToCheck {
			p, err := GetSourceDirs(path)
			if err != nil {
				log.Fatal(err)
			}
			paths = append(paths, p...)
		}
		PathsToCheck = paths
	}

	for _, path := range PathsToCheck {
		pkg, err := build.ImportDir(path.String(), build.AllowBinary)
		if err != nil {
			log.Fatal(err)
		}
		pkg.Name = filepath.Join(filepath.Base(path.Base), path.Rel)
		PrintPackage(defaultTemplate, pkg)

	}
}

type Path struct {
	Base, Rel string
}

func (p Path) String() string {
	return filepath.Join(p.Base, p.Rel)
}

/*
This will search the path.Base, and return a list of paths with the same base
and their corresponding path.Rel for their relative path.
*/
func GetSourceDirs(path Path) (paths []Path, err error) {
	found := func(thisPath string) bool {
		for _, p := range paths {
			if thisPath == p.String() {
				return true
			}
		}
		return false
	}
	isSource := func(f string, i os.FileInfo, e error) error {
		if e != nil {
			return e
		}
		if i.Mode().IsRegular() && filepath.Ext(f) == ".go" {
			dir := filepath.Dir(f)
			if !found(dir) {
				rel, err := filepath.Rel(path.Base, dir)
				if err != nil {
					return err
				}
				paths = append(paths, Path{Base: path.Base, Rel: rel})
			}
		}
		return nil
	}
	err = filepath.Walk(path.Base, isSource)
	return paths, err
}

func PrintPackage(t *template.Template, pkg *build.Package) error {
	return t.Execute(os.Stdout, pkg)
}

var defaultTemplate = template.Must(template.New("default").Parse(defaultOutput))
var defaultOutput = `
Package: {{.Name}}
ImportPath: {{.ImportPath}}
Dir: {{.Dir}}{{if .Imports}}
Imports: {{range .Imports}}
 {{.}} {{end}}
{{end}}
`

func IsDir(path string) bool {
	if fi, err := os.Stat(path); err == nil && fi.IsDir() {
		return true
	}
	return false
}
