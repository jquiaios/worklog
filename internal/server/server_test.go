package server_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/jquiaios/worklog/internal/db"
	"github.com/jquiaios/worklog/internal/entry"
	"github.com/jquiaios/worklog/internal/server"
)

func newTestServer(t *testing.T) (http.Handler, *db.DB) {
	t.Helper()
	store, err := db.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	return server.New(store), store
}

func insertTestEntry(t *testing.T, store *db.DB, typ entry.Type, body string, at time.Time) int64 {
	t.Helper()
	id, err := store.Insert(entry.Entry{Type: typ, Body: body, CreatedAt: at})
	if err != nil {
		t.Fatalf("Insert: %v", err)
	}
	return id
}

func TestListEntries_Empty(t *testing.T) {
	h, _ := newTestServer(t)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/entries", nil))

	if rec.Code != http.StatusOK {
		t.Errorf("status %d, want 200", rec.Code)
	}
	var entries []any
	if err := json.NewDecoder(rec.Body).Decode(&entries); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected empty array, got %d entries", len(entries))
	}
}

func TestListEntries_TypeFilter(t *testing.T) {
	h, store := newTestServer(t)
	now := time.Now()
	insertTestEntry(t, store, entry.Highlight, "hl entry", now)
	insertTestEntry(t, store, entry.Lowlight, "ll entry", now)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/entries?type=highlight", nil))

	var entries []map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&entries); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(entries) != 1 || entries[0]["type"] != "highlight" {
		t.Errorf("expected 1 highlight, got %v", entries)
	}
}

func TestListEntries_QuarterFilter(t *testing.T) {
	h, store := newTestServer(t)
	insertTestEntry(t, store, entry.Highlight, "in q1", time.Date(2026, 2, 1, 12, 0, 0, 0, time.Local))
	insertTestEntry(t, store, entry.Highlight, "in q2", time.Date(2026, 5, 1, 12, 0, 0, 0, time.Local))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/entries?quarter=1&year=2026", nil))

	var entries []map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&entries); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(entries) != 1 || entries[0]["body"] != "in q1" {
		t.Errorf("expected only q1 entry, got %v", entries)
	}
}

func TestListEntries_InvalidQuarter(t *testing.T) {
	h, store := newTestServer(t)
	insertTestEntry(t, store, entry.Highlight, "any", time.Now())

	// Quarter 5 is out of range — server skips the filter and returns all entries
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/entries?quarter=5&year=2026", nil))

	if rec.Code != http.StatusOK {
		t.Errorf("status %d, want 200", rec.Code)
	}
	var entries []map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&entries); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("expected all entries for invalid quarter, got %d", len(entries))
	}
}

func TestCreateEntry_Valid(t *testing.T) {
	h, _ := newTestServer(t)
	body := `{"type":"highlight","body":"great work"}`
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/entries", bytes.NewBufferString(body)))

	if rec.Code != http.StatusCreated {
		t.Errorf("status %d, want 201", rec.Code)
	}
	var resp map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if id, _ := resp["id"].(float64); id <= 0 {
		t.Errorf("expected positive id, got %v", resp["id"])
	}
}

func TestCreateEntry_EmptyBody(t *testing.T) {
	h, _ := newTestServer(t)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/entries", bytes.NewBufferString(`{"type":"highlight","body":""}`)))

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status %d, want 400", rec.Code)
	}
}

func TestCreateEntry_InvalidType(t *testing.T) {
	h, _ := newTestServer(t)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/entries", bytes.NewBufferString(`{"type":"unknown","body":"test"}`)))

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status %d, want 400", rec.Code)
	}
}

func TestCreateEntry_MalformedJSON(t *testing.T) {
	h, _ := newTestServer(t)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/entries", bytes.NewBufferString("{bad json")))

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status %d, want 400", rec.Code)
	}
}

func TestUpdateEntry_Valid(t *testing.T) {
	h, store := newTestServer(t)
	id := insertTestEntry(t, store, entry.Highlight, "original", time.Now())

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/entries/%d", id), bytes.NewBufferString(`{"body":"updated"}`))
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("status %d, want 204", rec.Code)
	}
}

func TestUpdateEntry_InvalidID(t *testing.T) {
	h, _ := newTestServer(t)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/api/entries/notanid", bytes.NewBufferString(`{"body":"x"}`)))

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status %d, want 400", rec.Code)
	}
}

func TestDeleteEntry_Valid(t *testing.T) {
	h, store := newTestServer(t)
	id := insertTestEntry(t, store, entry.Highlight, "to delete", time.Now())

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/entries/%d", id), nil))

	if rec.Code != http.StatusNoContent {
		t.Errorf("status %d, want 204", rec.Code)
	}
}

func TestDeleteEntry_InvalidID(t *testing.T) {
	h, _ := newTestServer(t)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/api/entries/notanid", nil))

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status %d, want 400", rec.Code)
	}
}
