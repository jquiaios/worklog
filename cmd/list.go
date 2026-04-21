package cmd

import (
	"fmt"
	"time"

	"github.com/jquiaios/worklog/internal/db"
	"github.com/jquiaios/worklog/internal/entry"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List logged entries (newest first)",
	RunE: func(cmd *cobra.Command, args []string) error {
		typeFilter, _ := cmd.Flags().GetString("type")
		switch typeFilter {
		case "hl":
			typeFilter = string(entry.Highlight)
		case "ll":
			typeFilter = string(entry.Lowlight)
		case "highlight", "lowlight", "":
		default:
			return fmt.Errorf("unknown type %q: use highlight/hl or lowlight/ll", typeFilter)
		}

		store, err := db.Open(dbPath())
		if err != nil {
			return fmt.Errorf("opening database: %w", err)
		}
		defer func() { _ = store.Close() }()

		entries, err := store.List(typeFilter, time.Time{}, time.Time{})
		if err != nil {
			return fmt.Errorf("listing entries: %w", err)
		}

		if len(entries) == 0 {
			fmt.Println("no entries yet")
			return nil
		}

		for _, e := range entries {
			label, ansi := typeLabel(e.Type)
			fmt.Printf("#%-3d  %s  %s%s\033[0m  %s\n",
				e.ID,
				e.CreatedAt.Local().Format("2006-01-02"),
				ansi, label,
				e.Body,
			)
		}
		return nil
	},
}

func typeLabel(t entry.Type) (label, ansi string) {
	switch t {
	case entry.Highlight:
		return "[highlight]", "\033[32m"
	case entry.Lowlight:
		return "[lowlight] ", "\033[31m"
	default:
		return "[" + string(t) + "]", ""
	}
}

func init() {
	listCmd.Flags().StringP("type", "t", "", "filter by type: highlight/hl or lowlight/ll")
	rootCmd.AddCommand(listCmd)
}
