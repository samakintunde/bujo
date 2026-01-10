package cmd

import (
	"fmt"

	"github.com/samakintunde/bujo-cli/internal/storage"
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
		err := initializeConfig(cmd)
		if err != nil {
			return err
		}
		entryType := "task"
		bullet := "- [ ]"
		if isEvent {
			entryType = "event"
			bullet = "- *"
		} else if isNote {
			entryType = "note"
			bullet = "-"
		}

		fs, err := storage.NewFSStore(cfg.Journal)
		if err != nil {
			return err
		}

		dayLog, err := fs.GetTodayPath()
		if err != nil {
			return err
		}

		err = fs.AppendLine(dayLog, fmt.Sprintf("%s %s", bullet, args[0]))
		if err != nil {
			return err
		}

		fmt.Printf("Added %s: %v\n", entryType, args)
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
