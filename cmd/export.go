package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jquiaios/worklog/internal/db"
	"github.com/jquiaios/worklog/internal/entry"
	"github.com/jquiaios/worklog/internal/export"
	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export entries as Markdown",
	Long: `Export work log entries as Markdown.

Defaults to the current quarter. Use -q for a specific quarter,
-y for a full year, or combine -q and -y for a specific quarter of a past year.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		quarterFlag, _ := cmd.Flags().GetString("quarter")
		yearFlag, _ := cmd.Flags().GetInt("year")
		outputFlag, _ := cmd.Flags().GetString("output")

		quarterChanged := cmd.Flags().Changed("quarter")
		yearChanged := cmd.Flags().Changed("year")

		now := time.Now()
		var from, to time.Time
		var title string
		isYear := false

		switch {
		case !quarterChanged && !yearChanged:
			q := (int(now.Month())-1)/3 + 1
			from, to, title = quarterBounds(q, now.Year())
		case quarterChanged:
			q, err := parseQuarter(quarterFlag)
			if err != nil {
				return err
			}
			year := now.Year()
			if yearChanged {
				year = yearFlag
			}
			from, to, title = quarterBounds(q, year)
		default:
			from = time.Date(yearFlag, time.January, 1, 0, 0, 0, 0, time.Local)
			to = time.Date(yearFlag+1, time.January, 1, 0, 0, 0, 0, time.Local)
			title = fmt.Sprintf("%d", yearFlag)
			isYear = true
		}

		store, err := db.Open(dbPath())
		if err != nil {
			return fmt.Errorf("opening database: %w", err)
		}
		defer store.Close()

		highlights, err := store.List(string(entry.Highlight), from, to)
		if err != nil {
			return err
		}
		lowlights, err := store.List(string(entry.Lowlight), from, to)
		if err != nil {
			return err
		}

		var md string
		if isYear {
			md = export.Grouped(title, highlights, lowlights)
		} else {
			md = export.Single(title, highlights, lowlights)
		}

		if outputFlag == "" {
			fmt.Print(md)
			return nil
		}

		if err := os.WriteFile(outputFlag, []byte(md), 0644); err != nil {
			return fmt.Errorf("writing file: %w", err)
		}
		fmt.Fprintf(os.Stderr, "exported to %s\n", outputFlag)
		return nil
	},
}

func parseQuarter(s string) (int, error) {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "Q1":
		return 1, nil
	case "Q2":
		return 2, nil
	case "Q3":
		return 3, nil
	case "Q4":
		return 4, nil
	default:
		return 0, fmt.Errorf("invalid quarter %q: use Q1, Q2, Q3, or Q4", s)
	}
}

func quarterBounds(q, year int) (from, to time.Time, title string) {
	month := time.Month((q-1)*3 + 1)
	from = time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
	to = from.AddDate(0, 3, 0)
	title = fmt.Sprintf("Q%d %d", q, year)
	return
}


func init() {
	exportCmd.Flags().StringP("quarter", "q", "", "quarter to export: Q1, Q2, Q3, Q4 (default: current quarter)")
	exportCmd.Flags().IntP("year", "y", 0, "year to export (default: current year)")
	exportCmd.Flags().StringP("output", "o", "", "write to file instead of stdout")
	rootCmd.AddCommand(exportCmd)
}
