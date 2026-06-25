package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
)

func TestHome(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handleHome(w, r)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
	body := w.Body.String()
	if !strings.Contains(body, "card-list-todo") {
		t.Error("expected card-list-todo in response")
	}
}

func TestNewCardForm(t *testing.T) {
	tests := []struct {
		name       string
		column     string
		wantID     string
		wantStatus int
	}{
		{"todo column", "todo", "form-container-todo", http.StatusOK},
		{"wip column", "wip", "form-container-wip", http.StatusOK},
		{"done column", "done", "form-container-done", http.StatusOK},
		{"defaults to todo", "", "form-container-todo", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := "/cards/new"
			if tt.column != "" {
				u += "?column=" + tt.column
			}
			r := httptest.NewRequest("GET", u, nil)
			w := httptest.NewRecorder()
			handleNewCardForm(w, r)

			res := w.Result()
			if res.StatusCode != tt.wantStatus {
				t.Fatalf("expected %d, got %d", tt.wantStatus, res.StatusCode)
			}
			body := w.Body.String()
			if !strings.Contains(body, tt.wantID) {
				t.Errorf("expected %q in response\n%s", tt.wantID, body)
			}
			if !strings.Contains(body, "hx-post") {
				t.Error("expected form element with hx-post")
			}
		})
	}
}

func TestCancelForm(t *testing.T) {
	tests := []struct {
		name       string
		column     string
		wantID     string
		wantStatus int
	}{
		{"todo column", "todo", "form-container-todo", http.StatusOK},
		{"wip column", "wip", "form-container-wip", http.StatusOK},
		{"done column", "done", "form-container-done", http.StatusOK},
		{"defaults to todo", "", "form-container-todo", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := "/cards/cancel"
			if tt.column != "" {
				u += "?column=" + tt.column
			}
			r := httptest.NewRequest("GET", u, nil)
			w := httptest.NewRecorder()
			handleFormCancel(w, r)

			res := w.Result()
			if res.StatusCode != tt.wantStatus {
				t.Fatalf("expected %d, got %d", tt.wantStatus, res.StatusCode)
			}
			body := w.Body.String()
			if !strings.Contains(body, tt.wantID) {
				t.Errorf("expected %q in response\n%s", tt.wantID, body)
			}
			if !strings.Contains(body, "Add a card") {
				t.Error("expected 'Add a card' button in response")
			}
		})
	}
}

func TestCreateCard(t *testing.T) {
	tests := []struct {
		name       string
		title      string
		desc       string
		column     string
		wantTitle  string
		wantDesc   string
		wantStatus int
		wantErr    bool
	}{
		{"with title and description", "Test Card", "A test description", "todo", "Test Card", "A test description", http.StatusOK, false},
		{"with title only", "Minimal", "", "wip", "Minimal", "", http.StatusOK, false},
		{"default column", "Default", "", "", "Default", "", http.StatusOK, false},
		{"missing title", "", "", "todo", "", "", http.StatusBadRequest, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := url.Values{}
			form.Set("title", tt.title)
			form.Set("description", tt.desc)
			form.Set("column", tt.column)

			r := httptest.NewRequest("POST", "/cards", strings.NewReader(form.Encode()))
			r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			w := httptest.NewRecorder()
			handleCreateCard(w, r)

			res := w.Result()
			if res.StatusCode != tt.wantStatus {
				t.Fatalf("expected %d, got %d", tt.wantStatus, res.StatusCode)
			}

			if tt.wantErr {
				return
			}

			body := w.Body.String()
			if !strings.Contains(body, tt.wantTitle) {
				t.Errorf("expected title %q in response\n%s", tt.wantTitle, body)
			}
			if tt.wantDesc != "" && !strings.Contains(body, tt.wantDesc) {
				t.Errorf("expected description %q in response\n%s", tt.wantDesc, body)
			}
			if !strings.Contains(body, "/cards/edit/") {
				t.Error("expected edit button with card ID in response")
			}
			if !strings.Contains(body, "closest .bg-white") {
				t.Error("expected edit button to target closest .bg-white")
			}

			hxTrigger := res.Header.Get("HX-Trigger")
			wantTrigger := "restore-form-" + tt.column
			if tt.column == "" {
				wantTrigger = "restore-form-todo"
			}
			if hxTrigger != wantTrigger {
				t.Errorf("expected HX-Trigger header %q, got %q", wantTrigger, hxTrigger)
			}
		})
	}
}

func TestEditCard(t *testing.T) {
	form := url.Values{}
	form.Set("title", "Edit Me")
	form.Set("description", "Original Desc")
	form.Set("column", "wip")

	r := httptest.NewRequest("POST", "/cards", strings.NewReader(form.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	handleCreateCard(w, r)

	var cardID int
	for id, c := range cards {
		if c.Title == "Edit Me" {
			cardID = id
			break
		}
	}
	if cardID == 0 {
		t.Fatal("card not found in store")
	}

	r2 := httptest.NewRequest("GET", "/cards/edit/"+strconv.Itoa(cardID)+"?column=wip", nil)
	r2.SetPathValue("id", strconv.Itoa(cardID))
	w2 := httptest.NewRecorder()
	handleEditCard(w2, r2)

	res := w2.Result()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}

	body := w2.Body.String()
	if !strings.Contains(body, "Edit Me") {
		t.Error("expected pre-filled title in edit form")
	}
	if !strings.Contains(body, "Original Desc") {
		t.Error("expected pre-filled description in edit form")
	}
	if !strings.Contains(body, "Save") {
		t.Error("expected Save button in edit form")
	}
	if !strings.Contains(body, "Cancel") {
		t.Error("expected Cancel button in edit form")
	}
	if !strings.Contains(body, "/cards/"+strconv.Itoa(cardID)) {
		t.Error("expected form action with card ID")
	}
}

func TestUpdateCard(t *testing.T) {
	form := url.Values{}
	form.Set("title", "Update Me")
	form.Set("description", "Old Desc")
	form.Set("column", "done")

	r := httptest.NewRequest("POST", "/cards", strings.NewReader(form.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	handleCreateCard(w, r)

	var cardID int
	for id, c := range cards {
		if c.Title == "Update Me" {
			cardID = id
			break
		}
	}
	if cardID == 0 {
		t.Fatal("card not found in store")
	}

	updateForm := url.Values{}
	updateForm.Set("title", "Updated Title")
	updateForm.Set("description", "New Desc")
	updateForm.Set("column", "done")

	r2 := httptest.NewRequest("PUT", "/cards/"+strconv.Itoa(cardID), strings.NewReader(updateForm.Encode()))
	r2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r2.SetPathValue("id", strconv.Itoa(cardID))
	w2 := httptest.NewRecorder()
	handleUpdateCard(w2, r2)

	res := w2.Result()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}

	body := w2.Body.String()
	if !strings.Contains(body, "Updated Title") {
		t.Error("expected updated title in response")
	}
	if !strings.Contains(body, "New Desc") {
		t.Error("expected updated description in response")
	}

	cardsMu.Lock()
	updated, ok := cards[cardID]
	cardsMu.Unlock()
	if !ok {
		t.Fatal("card not found in store")
	}
	if updated.Title != "Updated Title" {
		t.Errorf("expected stored title 'Updated Title', got %q", updated.Title)
	}
	if updated.Description != "New Desc" {
		t.Errorf("expected stored description 'New Desc', got %q", updated.Description)
	}
}

func TestGetCard(t *testing.T) {
	form := url.Values{}
	form.Set("title", "Get Me")
	form.Set("description", "Get Desc")
	form.Set("column", "todo")

	r := httptest.NewRequest("POST", "/cards", strings.NewReader(form.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	handleCreateCard(w, r)

	var cardID int
	for id, c := range cards {
		if c.Title == "Get Me" {
			cardID = id
			break
		}
	}
	if cardID == 0 {
		t.Fatal("card not found in store")
	}

	r2 := httptest.NewRequest("GET", "/cards/"+strconv.Itoa(cardID)+"?column=todo", nil)
	r2.SetPathValue("id", strconv.Itoa(cardID))
	w2 := httptest.NewRecorder()
	handleGetCard(w2, r2)

	res := w2.Result()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}

	body := w2.Body.String()
	if !strings.Contains(body, "Get Me") {
		t.Error("expected card title in response")
	}
}

func TestEditCard_NotFound(t *testing.T) {
	r := httptest.NewRequest("GET", "/cards/edit/99999?column=todo", nil)
	r.SetPathValue("id", "99999")
	w := httptest.NewRecorder()
	handleEditCard(w, r)

	res := w.Result()
	if res.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", res.StatusCode)
	}
}

func TestUpdateCard_NotFound(t *testing.T) {
	form := url.Values{}
	form.Set("title", "Orphan")
	form.Set("description", "")
	form.Set("column", "todo")

	r := httptest.NewRequest("PUT", "/cards/99999", strings.NewReader(form.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.SetPathValue("id", "99999")
	w := httptest.NewRecorder()
	handleUpdateCard(w, r)

	res := w.Result()
	if res.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", res.StatusCode)
	}
}

func TestUpdateCard_MissingTitle(t *testing.T) {
	form := url.Values{}
	form.Set("title", "Needs Title")
	form.Set("description", "")
	form.Set("column", "todo")

	r := httptest.NewRequest("POST", "/cards", strings.NewReader(form.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	handleCreateCard(w, r)

	var cardID int
	for id, c := range cards {
		if c.Title == "Needs Title" {
			cardID = id
			break
		}
	}

	updateForm := url.Values{}
	updateForm.Set("title", "")
	updateForm.Set("description", "")
	updateForm.Set("column", "todo")

	r2 := httptest.NewRequest("PUT", "/cards/"+strconv.Itoa(cardID), strings.NewReader(updateForm.Encode()))
	r2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r2.SetPathValue("id", strconv.Itoa(cardID))
	w2 := httptest.NewRecorder()
	handleUpdateCard(w2, r2)

	res := w2.Result()
	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", res.StatusCode)
	}
}

func TestGetCard_NotFound(t *testing.T) {
	r := httptest.NewRequest("GET", "/cards/99999?column=todo", nil)
	r.SetPathValue("id", "99999")
	w := httptest.NewRecorder()
	handleGetCard(w, r)

	res := w.Result()
	if res.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", res.StatusCode)
	}
}
