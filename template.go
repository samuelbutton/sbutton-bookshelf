package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"path/filepath"
)

func parseTemplate(filename string) *appTemplate {
	tmpl := template.Must(template.ParseFiles("templates/base.html"))

	path := filepath.Join("templates", filename)
	b, err := ioutil.ReadFile(path)

	if err != nil {
		panic(fmt.Errorf("could not read template: %v", err))
	}

	template.Must(tmpl.New("body").Parse(string(b)))

	return &appTemplate{tmpl.Lookup("base.html")}
}

type appTemplate struct {
	t *template.Template
}

func (tmpl *appTemplate) Execute(b *Bookshelf, w http.ResponseWriter, r *http.Request, data interface{}) *appError {
	d := struct {
		Data     interface{}
		LoggedIn bool
	}{
		Data:     data,
		LoggedIn: b.userLoggedIn,
	}

	if err := tmpl.t.Execute(w, d); err != nil {
		return b.appErrorf(r, err, "could not write template: %v", err)
	}

	return nil
}
