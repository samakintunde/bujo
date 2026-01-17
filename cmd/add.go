package cmd

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/samakintunde/bujo/internal/git"
	"github.com/samakintunde/bujo/internal/models"
	"github.com/samakintunde/bujo/internal/storage"
	"github.com/spf13/cobra"
)

type EntryTypeFlags struct {
	isTask  bool
	isEvent bool
	isNote  bool
}

var entryTypeFlags EntryTypeFlags

var addCmd = &cobra.Command{
	Use:   "add <text> [flags]",
	Short: "Add a task/event/note",
	Long:  "Add a task/event/note to the journal",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		err := initializeConfig(cmd)
		if err != nil {
			return err
		}

		if git.IsPresent() && !git.IsRepo(cfg.GetJournalPath()) {
			err = git.Init(cfg.GetJournalPath())
			if err != nil {
				return err
			}
		}

		fs, err := storage.NewFSStore(cfg.GetJournalPath())
		if err != nil {
			return err
		}

		entryFilePath, err := fs.EnsureDayPath(time.Now().Format(time.DateOnly))
		if err != nil {
			return err
		}

		entryType := inferEntryType(entryTypeFlags)
		entryContent := args[0]
		entry := models.NewEntry(entryType, entryContent)

		err = fs.AppendLine(entryFilePath, entry.RawString())
		if err != nil {
			return err
		}

		if git.IsPresent() {
			dir := filepath.Dir(entryFilePath)
			message := fmt.Sprintf("feat(bujo): add %s #%s\n", entryType, entry.ID)
			if err := git.Commit(dir, message); err != nil {
				return err
			}
		}

		fmt.Printf("Added %s #%s\n", entryType, entry.ID)
		return nil
	},
}

func inferEntryType(args EntryTypeFlags) models.EntryType {
	switch {
	case args.isTask:
		return models.EntryTypeTask
	case args.isEvent:
		return models.EntryTypeEvent
	case args.isNote:
		return models.EntryTypeNote
	default:
		return models.EntryTypeTask
	}
}

func init() {
	addCmd.Flags().BoolVar(&entryTypeFlags.isTask, "task", false, "Add a task")
	addCmd.Flags().BoolVar(&entryTypeFlags.isEvent, "event", false, "Add an event")
	addCmd.Flags().BoolVar(&entryTypeFlags.isNote, "note", false, "Add a note")

	addCmd.MarkFlagsMutuallyExclusive("task", "event", "note")

	rootCmd.AddCommand(addCmd)
}
