package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	isTask  bool
	isEvent bool
	isNote  bool
)

var addCmd = &cobra.Command{
	Use:   "add <text> [flags]",
	Short: "Add a task/event/note",
	Long:  "Add a task/event/note to the journal",
	RunE: func(cmd *cobra.Command, args []string) error {
		entryType := "task"
		if isEvent {
			entryType = "event"
		} else if isNote {
			entryType = "note"
		}

		fmt.Printf("Adding %s: %v\n", entryType, args)
		return nil
	},
}

func init() {
	addCmd.Flags().BoolVar(&isTask, "task", false, "Add a task")
	addCmd.Flags().BoolVar(&isEvent, "event", false, "Add an event")
	addCmd.Flags().BoolVar(&isNote, "note", false, "Add a note")

	addCmd.MarkFlagsMutuallyExclusive("task", "event", "note")

	rootCmd.AddCommand(addCmd)
}
