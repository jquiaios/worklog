package cmd

import (
	"fmt"
	"strconv"

	"github.com/jquiaios/worklog/internal/db"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:     "delete <id>",
	Aliases: []string{"rm"},
	Short:   "Delete an entry by ID",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil || id <= 0 {
			return fmt.Errorf("invalid id %q: must be a positive integer", args[0])
		}

		store, err := db.Open(dbPath())
		if err != nil {
			return fmt.Errorf("opening database: %w", err)
		}
		defer func() { _ = store.Close() }()

		found, err := store.Delete(id)
		if err != nil {
			return fmt.Errorf("deleting entry: %w", err)
		}
		if !found {
			return fmt.Errorf("no entry with id #%d", id)
		}
		fmt.Printf("deleted #%d\n", id)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
