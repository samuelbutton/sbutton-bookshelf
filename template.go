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
	// put this template in the body of the template already created in the
	// base.html step
	template.Must(tmpl.New("body").Parse(string(b)))

	// returns pointer to base, which now contains the requisite template
	return &appTemplate{tmpl.Lookup("base.html")}
}

type appTemplate struct {
	t *template.Template
}

func (tmpl *appTemplate) Execute(b *Bookshelf, w http.ResponseWriter, r *http.Request, data interface{}) *appError {
	d := struct {
		Data interface{}
	}{
		Data: data,
	}

	if err := tmpl.t.Execute(w, d); err != nil {
		return b.appErrorf(r, err, "could not write template: %v", err)
	}

	return nil
}
