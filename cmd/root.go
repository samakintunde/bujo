package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "bujo",
	Short: "A simple CLI tool for managing your daily tasks",
	Long:  `Bujo is a simple CLI tool for managing your daily tasks. It allows you to create, edit, and delete tasks, notes and events.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func Execute() error {
	return rootCmd.Execute()
}
