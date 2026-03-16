package templates

import (
	"embed"
	"html/template"
	"io"
	"strings"
	"time"
)

// BaseData is passed to all page templates
type BaseData struct {
	Title           string
	Theme           string
	IsAuthenticated bool
	Username        string
	IsAdmin         bool
	NavActive       string
	Error           string
	CurrentPath     string
}

//go:embed *.html
var FS embed.FS

// Execute renders base+page template. name is the page (e.g. "login", "servers").
func Execute(w io.Writer, name string, data interface{}) error {
	funcMap := template.FuncMap{
		"formatTime": func(t time.Time) string {
			if t.IsZero() {
				return "-"
			}
			return t.Format("2006-01-02 15:04")
		},
		"join": strings.Join,
	}
	tmpl, err := template.New("").Funcs(funcMap).ParseFS(FS, "base.html", name+".html")
	if err != nil {
		return err
	}
	return tmpl.ExecuteTemplate(w, "base", data)
}
