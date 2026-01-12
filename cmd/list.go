package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/samakintunde/bujo-cli/internal/parser"
	"github.com/samakintunde/bujo-cli/internal/storage"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list <date>",
	Short: "List the day's entries",
	Long:  "List journal's entries for the day",
	RunE: func(cmd *cobra.Command, args []string) error {
		parsedDate := time.Now()
		date := parsedDate.Format(time.DateOnly)
		if len(args) > 0 && args[0] != "" {
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

		fs, err := storage.NewFSStore(cfg.GetJournalPath())
		if err != nil {
			return err
		}
		dayLog, err := fs.GetDayPath(date)
		if err != nil {
			return err
		}
		entries, err := parser.Parse(dayLog)
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
