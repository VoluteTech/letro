package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sort"
	"strconv"
	"sync"
)

type Card struct {
	ID          int
	Title       string
	Description string
	Column      string
	Position    int
}

type HomeData struct {
	TodoCards []Card
	WipCards  []Card
	DoneCards []Card
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

// dict builds a map from alternating key-value pairs for use in templates.
func dict(values ...interface{}) map[string]interface{} {
	m := make(map[string]interface{})
	for i := 0; i < len(values); i += 2 {
		if i+1 < len(values) {
			key := fmt.Sprintf("%v", values[i])
			m[key] = values[i+1]
		}
	}
	return m
}

var templates = template.Must(
	template.New("").
		Funcs(template.FuncMap{"dict": dict}).
		ParseGlob("templates/*.html"),
)

func renderTemplate(w http.ResponseWriter, name string, data interface{}) {
	err := templates.ExecuteTemplate(w, name, data)
	if err != nil {
		log.Printf("Template error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func nextPositionInColumn(col string) int {
	maxPos := 0
	for _, c := range cards {
		if c.Column == col && c.Position > maxPos {
			maxPos = c.Position
		}
	}
	return maxPos + 1
}

func collectCardsByColumn() (todo, wip, done []Card) {
	for _, c := range cards {
		switch c.Column {
		case "todo":
			todo = append(todo, c)
		case "wip":
			wip = append(wip, c)
		case "done":
			done = append(done, c)
		}
	}
	sort.Slice(todo, func(i, j int) bool { return todo[i].Position < todo[j].Position })
	sort.Slice(wip, func(i, j int) bool { return wip[i].Position < wip[j].Position })
	sort.Slice(done, func(i, j int) bool { return done[i].Position < done[j].Position })
	return
}

// ---------------------------------------------------------------------------
// Handlers
// ---------------------------------------------------------------------------

func main() {
	http.HandleFunc("GET /", handleHome)
	http.HandleFunc("GET /cards/new", handleNewCardForm)
	http.HandleFunc("GET /cards/cancel", handleFormCancel)
	http.HandleFunc("POST /cards", handleCreateCard)
	http.HandleFunc("GET /cards/edit/{id}", handleEditCard)
	http.HandleFunc("PUT /cards/{id}", handleUpdateCard)
	http.HandleFunc("GET /cards/{id}", handleGetCard)
	http.HandleFunc("DELETE /cards/{id}", handleDeleteCard)
	http.HandleFunc("PATCH /cards/{id}/move", handleMoveCard)

	log.Println("Server starting on http://localhost:42069")
	log.Fatal(http.ListenAndServe(":42069", nil))
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	cardsMu.Lock()
	todo, wip, done := collectCardsByColumn()
	cardsMu.Unlock()

	renderTemplate(w, "home.html", HomeData{
		TodoCards: todo,
		WipCards:  wip,
		DoneCards: done,
	})
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
	card := Card{
		ID:          nextID,
		Title:       title,
		Description: desc,
		Column:      column,
		Position:    nextPositionInColumn(column),
	}
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
	existing, ok := cards[id]
	if !ok {
		cardsMu.Unlock()
		http.Error(w, "Card not found", http.StatusNotFound)
		return
	}
	card := Card{
		ID:          id,
		Title:       title,
		Description: desc,
		Column:      column,
		Position:    existing.Position,
	}
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

func handleDeleteCard(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid card ID", http.StatusBadRequest)
		return
	}

	cardsMu.Lock()
	if _, ok := cards[id]; !ok {
		cardsMu.Unlock()
		http.Error(w, "Card not found", http.StatusNotFound)
		return
	}
	delete(cards, id)
	cardsMu.Unlock()

	w.WriteHeader(http.StatusOK)
}

func handleMoveCard(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid card ID", http.StatusBadRequest)
		return
	}

	newColumn := r.FormValue("column")
	if newColumn == "" {
		http.Error(w, "No column provided", http.StatusBadRequest)
		return
	}

	cardsMu.Lock()
	card, ok := cards[id]
	if !ok {
		cardsMu.Unlock()
		http.Error(w, "Card not found", http.StatusNotFound)
		return
	}

	card.Column = newColumn
	card.Position = nextPositionInColumn(newColumn)
	cards[id] = card
	cardsMu.Unlock()

	renderTemplate(w, "response.html", TemplateData{Column: newColumn, Card: card})
}
