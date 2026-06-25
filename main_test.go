package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
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
