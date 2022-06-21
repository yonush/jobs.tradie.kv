// Basic template page regsitration and rendering
package main

//TODO
// - Nothing
import (
	"net/http"
	"path/filepath"
	"text/template"
)

//template handlers for the page views
type View struct {
	Template *template.Template
	Layout   string
}

//register new page view
func NewView(layout string, files ...string) *View {
	files = append(layoutFiles(), files...)
	t, err := template.ParseFiles(files...)
	if err != nil {
		panic(err)
	}

	return &View{
		Template: t,
		Layout:   layout,
	}
}

//render view with/without page data
func (v *View) Render(w http.ResponseWriter, data interface{}) error {
	return v.Template.ExecuteTemplate(w, v.Layout, data)
}

func layoutFiles() []string {
	files, err := filepath.Glob("./views/*.gohtml")
	if err != nil {
		panic(err)
	}
	return files
}
