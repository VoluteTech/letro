package main

import (
	"html/template"
	"net/http"
	"path/filepath"
)

type Card struct {
	Title string
}

func renderTemplate(w http.ResponseWriter, filename string, data interface{}) {
	path := filepath.Join("templates", filename)
	tmpl, err := template.ParseFiles(path)
	if err != nil {
		http.Error(w, "Template not found: " +err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, data)
}

func main() {
}
