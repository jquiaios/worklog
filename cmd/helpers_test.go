package cmd

import (
	"testing"
	"time"

	"github.com/jquiaios/worklog/internal/entry"
)

func TestParseQuarter_Valid(t *testing.T) {
	cases := []struct {
		input string
		want  int
	}{
		{"Q1", 1},
		{"Q2", 2},
		{"Q3", 3},
		{"Q4", 4},
		{"q1", 1},
		{"q4", 4},
		{" Q2 ", 2},
	}
	for _, tc := range cases {
		got, err := parseQuarter(tc.input)
		if err != nil {
			t.Errorf("parseQuarter(%q): unexpected error: %v", tc.input, err)
		}
		if got != tc.want {
			t.Errorf("parseQuarter(%q) = %d, want %d", tc.input, got, tc.want)
		}
	}
}

func TestParseQuarter_Invalid(t *testing.T) {
	for _, input := range []string{"Q5", "Q0", "1", "", "quarter1", "q"} {
		if _, err := parseQuarter(input); err == nil {
			t.Errorf("parseQuarter(%q): expected error, got nil", input)
		}
	}
}

func TestQuarterBounds_Q1(t *testing.T) {
	from, to, title := quarterBounds(1, 2026)
	wantFrom := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.Local)
	wantTo := time.Date(2026, time.April, 1, 0, 0, 0, 0, time.Local)

	if !from.Equal(wantFrom) {
		t.Errorf("from = %v, want %v", from, wantFrom)
	}
	if !to.Equal(wantTo) {
		t.Errorf("to = %v, want %v", to, wantTo)
	}
	if title != "Q1 2026" {
		t.Errorf("title = %q, want 'Q1 2026'", title)
	}
}

func TestQuarterBounds_Q4_YearBoundary(t *testing.T) {
	from, to, title := quarterBounds(4, 2025)
	wantFrom := time.Date(2025, time.October, 1, 0, 0, 0, 0, time.Local)
	wantTo := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.Local)

	if !from.Equal(wantFrom) {
		t.Errorf("from = %v, want %v", from, wantFrom)
	}
	if !to.Equal(wantTo) {
		t.Errorf("to = %v, want %v", to, wantTo)
	}
	if title != "Q4 2025" {
		t.Errorf("title = %q, want 'Q4 2025'", title)
	}
}

func TestTypeLabel_Highlight(t *testing.T) {
	label, ansi := typeLabel(entry.Highlight)
	if label != "[highlight]" {
		t.Errorf("label = %q, want '[highlight]'", label)
	}
	if ansi != "\033[32m" {
		t.Errorf("ansi = %q, want green escape", ansi)
	}
}

func TestTypeLabel_Lowlight(t *testing.T) {
	label, ansi := typeLabel(entry.Lowlight)
	if label != "[lowlight] " {
		t.Errorf("label = %q, want '[lowlight] '", label)
	}
	if ansi != "\033[31m" {
		t.Errorf("ansi = %q, want red escape", ansi)
	}
}

func TestTypeLabel_Unknown(t *testing.T) {
	label, ansi := typeLabel(entry.Type("custom"))
	if label != "[custom]" {
		t.Errorf("label = %q, want '[custom]'", label)
	}
	if ansi != "" {
		t.Errorf("ansi = %q, want empty for unknown type", ansi)
	}
}
