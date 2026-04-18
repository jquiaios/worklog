package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jquiaios/worklog/internal/db"
	"github.com/jquiaios/worklog/internal/entry"
	"github.com/jquiaios/worklog/internal/tui"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "worklog",
	Short: "Work log — capture highlights and lowlights fast",
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := db.Open(dbPath())
		if err != nil {
			return fmt.Errorf("opening database: %w", err)
		}
		defer store.Close()

		highlights, err := store.List(string(entry.Highlight))
		if err != nil {
			return err
		}
		lowlights, err := store.List(string(entry.Lowlight))
		if err != nil {
			return err
		}

		m := tui.New(store, highlights, lowlights)
		_, err = tea.NewProgram(m, tea.WithAltScreen()).Run()
		return err
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func dbPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error: cannot determine home directory:", err)
		os.Exit(1)
	}
	return filepath.Join(home, ".worklog", "worklog.db")
}
