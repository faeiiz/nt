package cmd

import (
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

var viewCmd = &cobra.Command{
	Use:   "view <id>",
	Short: "view a note",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id64, err := strconv.ParseUint(args[0], 10, 64)
		if err != nil {
			return err
		}
		n, err := store.Get(id64)
		if err != nil {
			return err
		}
		fmt.Printf("ID: %d\nTitle: %s\nDate: %s\n\n%s\n", n.ID, n.Title, n.CreatedAt.Format(time.RFC3339), n.Body)
		return nil
	},
}
