package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	dbStorage "github.com/samakintunde/bujo-cli/internal/db"
	fsStorage "github.com/samakintunde/bujo-cli/internal/storage"
	"github.com/samakintunde/bujo-cli/internal/sync"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list <date>",
	Short: "List the day's entries",
	Long:  "List journal's entries for the day",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) > 1 {
			return fmt.Errorf("too many arguments")
		}
		if len(args) == 1 && args[0] != "" {
			if _, err := time.Parse(time.DateOnly, args[0]); err != nil {
				return fmt.Errorf("invalid date format")
			}
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		parsedDate := time.Now()
		date := parsedDate.Format(time.DateOnly)
		if len(args) > 0 {
			parsed, err := time.Parse(time.DateOnly, args[0])
			if err != nil {
				return err
			}
			parsedDate = parsed
			date = parsed.Format(time.DateOnly)
		}
		err := initializeConfig(cmd)
		if err != nil {
			return err
		}

		if err := os.MkdirAll(cfg.GetDBPath(), 0755); err != nil {
			return err
		}

		dbStore, err := dbStorage.New(cfg.GetDBPath())
		if err != nil {
			return err
		}
		defer dbStore.Close()

		fsStore, err := fsStorage.NewFSStore(cfg.GetJournalPath())
		if err != nil {
			return err
		}

		syncer := sync.NewSyncer(cfg.GetJournalPath(), dbStore)
		if err := syncer.Sync(); err != nil {
			return err
		}

		dayLog, err := fsStore.GetDayPath(date)
		if err != nil {
			return err
		}

		entries, err := dbStore.GetEntriesByFile(dayLog)
		if err != nil {
			return err
		}

		header := fmt.Sprintf("Entries (%s):\n", parsedDate.Format("2 January, 2006"))
		border := strings.Repeat("-", len(header))
		var body strings.Builder
		for _, entry := range entries {
			body.WriteString(entry.DisplayString())
			body.WriteString("\n")
		}
		fmt.Printf("%s%s\n%s", header, border, body.String())
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
