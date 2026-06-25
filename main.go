package main

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"sync"
)

type Card struct {
	ID          int
	Title       string
	Description string
}

type TemplateData struct {
	Column string
	Card   Card
}

var (
	cards   = make(map[int]Card)
	nextID  int
	cardsMu sync.Mutex
)

func renderTemplate(w http.ResponseWriter, filename string, data interface{}) {
	path := filepath.Join("templates", filename)
	tmpl, err := template.ParseFiles(path)
	if err != nil {
		http.Error(w, "Template not found: "+err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, data)
}

func main() {
	http.HandleFunc("GET /", handleHome)
	http.HandleFunc("GET /cards/new", handleNewCardForm)
	http.HandleFunc("GET /cards/cancel", handleFormCancel)
	http.HandleFunc("POST /cards", handleCreateCard)
	http.HandleFunc("GET /cards/edit/{id}", handleEditCard)
	http.HandleFunc("PUT /cards/{id}", handleUpdateCard)
	http.HandleFunc("GET /cards/{id}", handleGetCard)

	log.Println("Server starting on http://localhost:42069")
	log.Fatal(http.ListenAndServe(":42069", nil))
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "home.html", nil)
}

func handleNewCardForm(w http.ResponseWriter, r *http.Request) {
	column := r.URL.Query().Get("column")
	if column == "" {
		column = "todo"
	}
	renderTemplate(w, "form.html", TemplateData{Column: column})
}

func handleFormCancel(w http.ResponseWriter, r *http.Request) {
	column := r.URL.Query().Get("column")
	if column == "" {
		column = "todo"
	}
	renderTemplate(w, "button.html", TemplateData{Column: column})
}

func handleCreateCard(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("title")
	desc := r.FormValue("description")
	column := r.FormValue("column")
	if column == "" {
		column = "todo"
	}
	if title == "" {
		http.Error(w, "No title provided", http.StatusBadRequest)
		return
	}

	cardsMu.Lock()
	nextID++
	card := Card{ID: nextID, Title: title, Description: desc}
	cards[nextID] = card
	cardsMu.Unlock()

	w.Header().Set("HX-Trigger", "restore-form-"+column)
	renderTemplate(w, "response.html", TemplateData{Column: column, Card: card})
}

func handleEditCard(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid card ID", http.StatusBadRequest)
		return
	}

	cardsMu.Lock()
	card, ok := cards[id]
	cardsMu.Unlock()
	if !ok {
		http.Error(w, "Card not found", http.StatusNotFound)
		return
	}

	column := r.URL.Query().Get("column")
	if column == "" {
		column = "todo"
	}

	renderTemplate(w, "edit.html", TemplateData{Column: column, Card: card})
}

func handleUpdateCard(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid card ID", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	desc := r.FormValue("description")
	column := r.FormValue("column")
	if column == "" {
		column = "todo"
	}
	if title == "" {
		http.Error(w, "No title provided", http.StatusBadRequest)
		return
	}

	cardsMu.Lock()
	if _, ok := cards[id]; !ok {
		cardsMu.Unlock()
		http.Error(w, "Card not found", http.StatusNotFound)
		return
	}
	card := Card{ID: id, Title: title, Description: desc}
	cards[id] = card
	cardsMu.Unlock()

	renderTemplate(w, "response.html", TemplateData{Column: column, Card: card})
}

func handleGetCard(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid card ID", http.StatusBadRequest)
		return
	}

	cardsMu.Lock()
	card, ok := cards[id]
	cardsMu.Unlock()
	if !ok {
		http.Error(w, "Card not found", http.StatusNotFound)
		return
	}

	column := r.URL.Query().Get("column")
	if column == "" {
		column = "todo"
	}

	renderTemplate(w, "response.html", TemplateData{Column: column, Card: card})
}
