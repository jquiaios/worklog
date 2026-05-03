package db_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/jquiaios/worklog/internal/db"
	"github.com/jquiaios/worklog/internal/entry"
)

func openTestDB(t *testing.T) *db.DB {
	t.Helper()
	store, err := db.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	return store
}

func insertEntry(t *testing.T, store *db.DB, typ entry.Type, body string, at time.Time) int64 {
	t.Helper()
	id, err := store.Insert(entry.Entry{Type: typ, Body: body, CreatedAt: at})
	if err != nil {
		t.Fatalf("Insert %q: %v", body, err)
	}
	return id
}

func TestOpen(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "worklog.db")

	s1, err := db.Open(path)
	if err != nil {
		t.Fatalf("first open: %v", err)
	}
	_ = s1.Close()

	// Second open on same path must not error (schema is idempotent via CREATE IF NOT EXISTS)
	s2, err := db.Open(path)
	if err != nil {
		t.Fatalf("second open: %v", err)
	}
	_ = s2.Close()
}

func TestInsert(t *testing.T) {
	store := openTestDB(t)
	id, err := store.Insert(entry.Entry{
		Type:      entry.Highlight,
		Body:      "shipped the feature",
		CreatedAt: time.Now(),
	})
	if err != nil {
		t.Fatalf("Insert: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestList_NoFilter(t *testing.T) {
	store := openTestDB(t)
	now := time.Now()

	insertEntry(t, store, entry.Highlight, "win one", now.Add(-2*time.Hour))
	insertEntry(t, store, entry.Lowlight, "miss one", now.Add(-1*time.Hour))
	insertEntry(t, store, entry.Highlight, "win two", now)

	got, err := store.List("", time.Time{}, time.Time{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(got))
	}
	if got[0].Body != "win two" {
		t.Errorf("expected newest first, got %q", got[0].Body)
	}
}

func TestList_TypeFilter(t *testing.T) {
	store := openTestDB(t)
	now := time.Now()

	insertEntry(t, store, entry.Highlight, "hl", now)
	insertEntry(t, store, entry.Lowlight, "ll", now)

	hls, err := store.List(string(entry.Highlight), time.Time{}, time.Time{})
	if err != nil {
		t.Fatalf("List highlights: %v", err)
	}
	if len(hls) != 1 || hls[0].Body != "hl" {
		t.Errorf("unexpected highlights: %v", hls)
	}

	lls, err := store.List(string(entry.Lowlight), time.Time{}, time.Time{})
	if err != nil {
		t.Fatalf("List lowlights: %v", err)
	}
	if len(lls) != 1 || lls[0].Body != "ll" {
		t.Errorf("unexpected lowlights: %v", lls)
	}
}

func TestList_DateRange(t *testing.T) {
	store := openTestDB(t)
	base := time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)

	insertEntry(t, store, entry.Highlight, "before", base.Add(-24*time.Hour))
	insertEntry(t, store, entry.Highlight, "inside", base)
	insertEntry(t, store, entry.Highlight, "after", base.Add(24*time.Hour))

	from := base.Add(-1 * time.Hour)
	to := base.Add(1 * time.Hour)

	got, err := store.List("", from, to)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != 1 || got[0].Body != "inside" {
		t.Errorf("expected only 'inside', got %v", got)
	}
}

func TestUpdate(t *testing.T) {
	store := openTestDB(t)
	id := insertEntry(t, store, entry.Highlight, "original", time.Now())

	if err := store.Update(id, "updated"); err != nil {
		t.Fatalf("Update: %v", err)
	}

	entries, err := store.List("", time.Time{}, time.Time{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if entries[0].Body != "updated" {
		t.Errorf("expected 'updated', got %q", entries[0].Body)
	}
}

func TestDelete_Exists(t *testing.T) {
	store := openTestDB(t)
	id := insertEntry(t, store, entry.Highlight, "to delete", time.Now())

	deleted, err := store.Delete(id)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if !deleted {
		t.Error("expected deleted=true")
	}

	entries, err := store.List("", time.Time{}, time.Time{})
	if err != nil {
		t.Fatalf("List after delete: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected empty list after delete, got %d entries", len(entries))
	}
}

func TestDelete_NotFound(t *testing.T) {
	store := openTestDB(t)
	deleted, err := store.Delete(9999)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if deleted {
		t.Error("expected deleted=false for non-existent ID")
	}
}
