package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/you/nt/storage"
)

var dbPath string
var store *storage.BoltStore

var rootCmd = &cobra.Command{
	Use:   "nt",
	Short: "nt — terminal notes",
	Long:  "nt is a small offline terminal note-taking app using bbolt, cobra and bubbletea.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if dbPath == "" {
			p, err := storage.DefaultDBPath()
			if err != nil {
				return err
			}
			dbPath = p
		} else {
			// expand tilde
			if dbPath[:2] == "~/" {
				dbPath = filepath.Join(os.Getenv("HOME"), dbPath[2:])
			}
		}
		var err error
		store, err = storage.NewBoltStore(dbPath)
		return err
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if store != nil {
			_ = store.Close()
		}
	},
}

func Execute() {
	rootCmd.PersistentFlags().StringVar(&dbPath, "db", "", "path to notes.db")
	// add subcommands
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(viewCmd)
	rootCmd.AddCommand(tuiCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
