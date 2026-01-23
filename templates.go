// Package baseproject is the main package of the golang-base-project which defines all routes and database connections and settings, the glue to the entire application
package baseproject

import (
	"html/template"
	"io"
	"io/fs"
	"strings"
)

func loadTemplates() (*template.Template, error) {
	// var err4 error
	templ := template.New("")
	err := fs.WalkDir(staticFS, ".", func(path string, dirEntry fs.DirEntry, err error) error {
		if dirEntry.IsDir() {
			return nil
		}
		file, err2 := staticFS.Open(path)
		if err2 != nil {
			return err2
		}
		contents, err3 := io.ReadAll(file)
		if err3 != nil {
			return err3
		}
		parts := strings.Split(path, "/")
		if len(parts) > 0 && strings.HasSuffix(parts[len(parts)-1], ".gohtml") {
			lastPart := parts[len(parts)-1]
			templ = template.Must(templ.New(lastPart).Parse(string(contents)))
			//templ, err4 = templ.New(lastPart).Parse(string(contents))
			//if err4 != nil {
			//	return err4
			//}
		}
		return nil
	})
	return templ, err
}
