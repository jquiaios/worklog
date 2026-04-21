package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jquiaios/worklog/internal/db"
	"github.com/jquiaios/worklog/internal/tui"
	"github.com/spf13/cobra"
)

// Version is the build-time application version, injected via ldflags.
// Falls back to "dev" for local (non-released) builds.
var Version = "dev"

var rootCmd = &cobra.Command{
	Use:     "worklog",
	Short:   "Work log - capture highlights and lowlights fast",
	Version: Version,
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := db.Open(dbPath())
		if err != nil {
			return fmt.Errorf("opening database: %w", err)
		}
		defer func() { _ = store.Close() }()

		m, err := tui.New(store)
		if err != nil {
			return err
		}
		_, err = tea.NewProgram(m, tea.WithAltScreen()).Run()
		return err
	},
}

func Execute() {
	rootCmd.Version = Version
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
