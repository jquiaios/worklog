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

var highlightCmd = &cobra.Command{
	Use:     "highlight <text>",
	Aliases: []string{"hl"},
	Short:   "Log a highlight",
	Args:    cobra.MinimumNArgs(1),
	Run:     addEntry(entry.Highlight),
}

var lowlightCmd = &cobra.Command{
	Use:     "lowlight <text>",
	Aliases: []string{"ll"},
	Short:   "Log a lowlight",
	Args:    cobra.MinimumNArgs(1),
	Run:     addEntry(entry.Lowlight),
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
		if err := store.Insert(e); err != nil {
			fmt.Fprintln(os.Stderr, "error saving entry:", err)
			os.Exit(1)
		}
		fmt.Printf("[%s] %s\n", t, body)
	}
}

func init() {
	rootCmd.AddCommand(highlightCmd, lowlightCmd)
}
