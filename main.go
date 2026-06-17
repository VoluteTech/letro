package main

import (
	"html/template"
	"log"
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
	http.HandleFunc("GET /", handleHome)
	http.HandleFunc("GET /cards/new", handleNewCardForm)
	http.HandleFunc("GET /cards/cancel", handleFormCancel)
	http.HandleFunc("POST /cards", handleCreateCard)

	log.Println("Server starting on http://localhost:42069")
	log.Fatal(http.ListenAndServe(":42069", nil))
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "home.html", nil)
}

func handleNewCardForm(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "form.html", nil)
}

func handleFormCancel(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "button.html", nil)
}

func handleCreateCard(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("title")
	if title == "" {
		http.Error(w, "No title provided", http.StatusBadRequest)
		return
	}

	card := Card{ Title: title }
	renderTemplate(w, "response.html", card)
}


