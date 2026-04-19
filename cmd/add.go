package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jquiaios/worklog/internal/db"
	"github.com/jquiaios/worklog/internal/entry"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Log a new entry",
	Long:  "Log a new highlight or lowlight.\n\nExamples:\n  worklog add hl \"Won the sprint demo\"\n  worklog add ll \"Missed the deploy window\"",
}

var addHighlightCmd = &cobra.Command{
	Use:     "highlight <text>",
	Aliases: []string{"hl"},
	Short:   "[hl] Log a highlight",
	Args:    cobra.MinimumNArgs(1),
	Run:     addEntry(entry.Highlight),
}

var addLowlightCmd = &cobra.Command{
	Use:     "lowlight <text>",
	Aliases: []string{"ll"},
	Short:   "[ll] Log a lowlight",
	Args:    cobra.MinimumNArgs(1),
	Run:     addEntry(entry.Lowlight),
}

// Hidden top-level shortcuts so muscle memory still works.
var hlCmd = &cobra.Command{
	Use:    "hl <text>",
	Hidden: true,
	Args:   cobra.MinimumNArgs(1),
	Run:    addEntry(entry.Highlight),
}

var llCmd = &cobra.Command{
	Use:    "ll <text>",
	Hidden: true,
	Args:   cobra.MinimumNArgs(1),
	Run:    addEntry(entry.Lowlight),
}

func addEntry(t entry.Type) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		body := strings.TrimSpace(strings.Join(args, " "))
		if body == "" {
			fmt.Fprintln(os.Stderr, "error: entry body cannot be empty")
			os.Exit(1)
		}
		store, err := db.Open(dbPath())
		if err != nil {
			fmt.Fprintln(os.Stderr, "error opening database:", err)
			os.Exit(1)
		}
		defer store.Close()

		e := entry.Entry{Type: t, Body: body, CreatedAt: time.Now()}
		if _, err := store.Insert(e); err != nil {
			fmt.Fprintln(os.Stderr, "error saving entry:", err)
			os.Exit(1)
		}
		fmt.Printf("[%s] %s\n", t, body)
	}
}

func init() {
	addCmd.AddCommand(addHighlightCmd, addLowlightCmd)
	rootCmd.AddCommand(addCmd, hlCmd, llCmd)
}
