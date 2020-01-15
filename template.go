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
		Messages []string
	}{
		Data:     data,
		LoggedIn: b.userLoggedIn,
		Messages: b.Messages,
	}

	if err := tmpl.t.Execute(w, d); err != nil {
		b.addMessage("Could not load page, please try again!")
		http.Redirect(w, r, "/books", http.StatusFound)
		return b.appErrorf(r, err, "could not write template: %v", err)
	}

	b.Messages = nil

	return nil
}
