package cmd

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list notes",
	RunE: func(cmd *cobra.Command, args []string) error {
		notes, err := store.List()
		if err != nil {
			return err
		}
		// sort by ID asc
		sort.Slice(notes, func(i, j int) bool { return notes[i].ID < notes[j].ID })
		for _, n := range notes {
			fmt.Printf("%d\t%s\t%s\n", n.ID, n.CreatedAt.Format("2006-01-02"), n.Title)
		}
		return nil
	},
}
