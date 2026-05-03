package export_test

import (
	"strings"
	"testing"
	"time"

	"github.com/jquiaios/worklog/internal/entry"
	"github.com/jquiaios/worklog/internal/export"
)

func makeEntry(body string, year int, month time.Month, day int) entry.Entry {
	return entry.Entry{
		Type:      entry.Highlight,
		Body:      body,
		CreatedAt: time.Date(year, month, day, 12, 0, 0, 0, time.UTC),
	}
}

func TestDefaultFilename(t *testing.T) {
	cases := []struct {
		label, want string
	}{
		{"Q1 2026", "worklog-q1-2026.md"},
		{"Q4 2025", "worklog-q4-2025.md"},
		{"2026", "worklog-2026.md"},
		{"All time", "worklog-all.md"},
	}
	for _, tc := range cases {
		got := export.DefaultFilename(tc.label)
		if got != tc.want {
			t.Errorf("DefaultFilename(%q) = %q, want %q", tc.label, got, tc.want)
		}
	}
}

func TestSingle_Empty(t *testing.T) {
	out := export.Single("Q1 2026", nil, nil)
	if !strings.Contains(out, "# Work Log - Q1 2026") {
		t.Error("missing title")
	}
	if strings.Count(out, "_Nothing logged._") != 2 {
		t.Errorf("expected 2 × '_Nothing logged._', got %d", strings.Count(out, "_Nothing logged._"))
	}
}

func TestSingle_WithEntries(t *testing.T) {
	hls := []entry.Entry{makeEntry("shipped auth", 2026, time.January, 5)}
	lls := []entry.Entry{makeEntry("missed deploy", 2026, time.February, 10)}
	out := export.Single("Q1 2026", hls, lls)

	for _, want := range []string{
		"## Highlights",
		"## Lowlights",
		"shipped auth",
		"missed deploy",
		"Jan 5",
		"Feb 10",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q", want)
		}
	}
}

func TestGrouped_SingleQuarter(t *testing.T) {
	hls := []entry.Entry{
		makeEntry("win a", 2026, time.January, 1),
		makeEntry("win b", 2026, time.February, 1),
	}
	out := export.Grouped("2026", hls, nil)

	if !strings.Contains(out, "### Q1 2026") {
		t.Error("missing '### Q1 2026' subheading")
	}
	if strings.Count(out, "### Q") != 1 {
		t.Errorf("expected exactly 1 quarter subheading, got %d", strings.Count(out, "### Q"))
	}
}

func TestGrouped_MultipleQuarters(t *testing.T) {
	// Simulates DB order: newest first
	hls := []entry.Entry{
		makeEntry("Q3 win", 2026, time.July, 1),
		makeEntry("Q2 win", 2026, time.April, 1),
		makeEntry("Q1 win", 2026, time.January, 1),
	}
	out := export.Grouped("2026", hls, nil)

	q1 := strings.Index(out, "Q1")
	q2 := strings.Index(out, "Q2")
	q3 := strings.Index(out, "Q3")
	if q1 == -1 || q2 == -1 || q3 == -1 {
		t.Fatal("missing quarter headings in output")
	}
	if q1 >= q2 || q2 >= q3 {
		t.Error("quarters not in chronological order (expected Q1 < Q2 < Q3)")
	}
}

func TestGrouped_CrossYear(t *testing.T) {
	// Newest first, as returned by DB
	hls := []entry.Entry{
		makeEntry("2026 win", 2026, time.January, 1),
		makeEntry("2025 win", 2025, time.October, 1),
	}
	out := export.Grouped("All time", hls, nil)

	if !strings.Contains(out, "Q1 2026") {
		t.Error("missing 'Q1 2026'")
	}
	if !strings.Contains(out, "Q4 2025") {
		t.Error("missing 'Q4 2025'")
	}
	// After reversal, 2025 Q4 must appear before 2026 Q1
	if strings.Index(out, "Q4 2025") > strings.Index(out, "Q1 2026") {
		t.Error("expected Q4 2025 before Q1 2026 in chronological order")
	}
}
