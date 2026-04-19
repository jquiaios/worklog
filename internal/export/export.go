package export

import (
	"fmt"
	"strings"

	"github.com/jquiaios/worklog/internal/entry"
)

// Single renders a flat export for one period (no quarter subheadings).
func Single(title string, highlights, lowlights []entry.Entry) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Work Log - %s\n", title)
	writeSection(&b, "Highlights", highlights)
	writeSection(&b, "Lowlights", lowlights)
	return b.String()
}

// Grouped renders entries grouped by quarter under each section - used for year exports.
func Grouped(title string, highlights, lowlights []entry.Entry) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Work Log - %s\n", title)
	writeSectionGrouped(&b, "Highlights", highlights)
	writeSectionGrouped(&b, "Lowlights", lowlights)
	return b.String()
}

// DefaultFilename returns a sensible export filename for a given period label.
// "Q1 2026" → "worklog-Q1-2026.md", "2026" → "worklog-2026.md", "All time" → "worklog-all.md"
func DefaultFilename(periodLabel string) string {
	label := strings.ToLower(strings.ReplaceAll(periodLabel, " ", "-"))
	if label == "all-time" {
		label = "all"
	}
	return "worklog-" + label + ".md"
}

func writeSection(b *strings.Builder, heading string, entries []entry.Entry) {
	fmt.Fprintf(b, "\n## %s\n\n", heading)
	if len(entries) == 0 {
		fmt.Fprintln(b, "_Nothing logged._")
		return
	}
	for _, e := range entries {
		fmt.Fprintf(b, "- %s _(%s)_\n", e.Body, e.CreatedAt.Local().Format("Jan 2"))
	}
}

type quarterGroup struct {
	q, year int
	entries []entry.Entry
}

func groupByQuarter(entries []entry.Entry) []quarterGroup {
	var groups []quarterGroup
	index := map[string]int{}

	for _, e := range entries {
		year := e.CreatedAt.Year()
		q := (int(e.CreatedAt.Month())-1)/3 + 1
		key := fmt.Sprintf("Q%d-%d", q, year)
		if i, ok := index[key]; ok {
			groups[i].entries = append(groups[i].entries, e)
		} else {
			index[key] = len(groups)
			groups = append(groups, quarterGroup{q: q, year: year, entries: []entry.Entry{e}})
		}
	}

	// Entries come newest-first from DB, so groups are newest-first too.
	// Reverse to chronological order for a readable review document.
	for i, j := 0, len(groups)-1; i < j; i, j = i+1, j-1 {
		groups[i], groups[j] = groups[j], groups[i]
	}
	return groups
}

func writeSectionGrouped(b *strings.Builder, heading string, entries []entry.Entry) {
	fmt.Fprintf(b, "\n## %s\n", heading)
	if len(entries) == 0 {
		fmt.Fprint(b, "\n_Nothing logged._\n")
		return
	}
	for _, g := range groupByQuarter(entries) {
		fmt.Fprintf(b, "\n### Q%d %d\n\n", g.q, g.year)
		for i := len(g.entries) - 1; i >= 0; i-- {
			e := g.entries[i]
			fmt.Fprintf(b, "- %s _(%s)_\n", e.Body, e.CreatedAt.Local().Format("Jan 2"))
		}
	}
}
