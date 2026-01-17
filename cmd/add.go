package cmd

import (
	"fmt"
	"time"

	"github.com/samakintunde/bujo/internal/git"
	"github.com/samakintunde/bujo/internal/models"
	"github.com/samakintunde/bujo/internal/service"
	"github.com/samakintunde/bujo/internal/storage"
	"github.com/samakintunde/bujo/internal/sync"
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

		db, err := storage.NewDBStore(cfg.GetDBPath())
		if err != nil {
			return err
		}
		defer db.Close()

		fs, err := storage.NewFSStore(cfg.GetJournalPath())
		if err != nil {
			return err
		}

		syncer := sync.NewSyncer(cfg.GetJournalPath(), db)
		svc := service.NewJournalService(fs, db, syncer)

		entryType := inferEntryType(entryTypeFlags)
		entryContent := args[0]

		entry, err := svc.AddEntry(entryContent, entryType, time.Now())
		if err != nil {
			return err
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
