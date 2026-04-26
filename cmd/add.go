package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/jquiaios/worklog/internal/db"
	"github.com/jquiaios/worklog/internal/entry"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Log a new entry",
	Long:  "Log a new highlight or lowlight.",
	Example: `  worklog add highlight "Won the sprint demo"
  worklog add hl "Won the sprint demo"
  worklog highlight "Won the sprint demo"
  worklog hl "Won the sprint demo"

  worklog add lowlight "Missed the deploy window"
  worklog add ll "Missed the deploy window"
  worklog lowlight "Missed the deploy window"
  worklog ll "Missed the deploy window"`,
}

var addHighlightCmd = &cobra.Command{
	Use:     "highlight <text>",
	Aliases: []string{"hl"},
	Short:   "[hl] Log a highlight",
	Args:    cobra.MinimumNArgs(1),
	RunE:    addEntry(entry.Highlight),
}

var addLowlightCmd = &cobra.Command{
	Use:     "lowlight <text>",
	Aliases: []string{"ll"},
	Short:   "[ll] Log a lowlight",
	Args:    cobra.MinimumNArgs(1),
	RunE:    addEntry(entry.Lowlight),
}

// Hidden top-level shortcuts: worklog highlight / worklog hl, worklog lowlight / worklog ll.
var hlCmd = &cobra.Command{
	Use:     "highlight <text>",
	Aliases: []string{"hl"},
	Hidden:  true,
	Args:    cobra.MinimumNArgs(1),
	RunE:    addEntry(entry.Highlight),
}

var llCmd = &cobra.Command{
	Use:     "lowlight <text>",
	Aliases: []string{"ll"},
	Hidden:  true,
	Args:    cobra.MinimumNArgs(1),
	RunE:    addEntry(entry.Lowlight),
}

func addEntry(t entry.Type) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		body := strings.TrimSpace(strings.Join(args, " "))
		if body == "" {
			return fmt.Errorf("entry body cannot be empty")
		}
		store, err := db.Open(dbPath())
		if err != nil {
			return fmt.Errorf("opening database: %w", err)
		}
		defer func() { _ = store.Close() }()

		e := entry.Entry{Type: t, Body: body, CreatedAt: time.Now()}
		if _, err := store.Insert(e); err != nil {
			return fmt.Errorf("saving entry: %w", err)
		}
		_, ansi := typeLabel(t)
		fmt.Printf("Added %s[%s]\033[0m \033[2m%s\033[0m\n", ansi, t, body)
		return nil
	}
}

func init() {
	addCmd.AddCommand(addHighlightCmd, addLowlightCmd)
	rootCmd.AddCommand(addCmd, hlCmd, llCmd)
}
