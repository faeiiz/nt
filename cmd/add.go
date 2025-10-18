package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <title> <body>",
	Short: "add a new note",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		title, body := args[0], args[1]
		n, err := store.Add(title, body)
		if err != nil {
			return err
		}
		fmt.Printf("Added note %d: %s\n", n.ID, n.Title)
		return nil
	},
}
