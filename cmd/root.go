package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/samakintunde/bujo-cli/internal/config"
	"github.com/samakintunde/bujo-cli/internal/storage"
	"github.com/samakintunde/bujo-cli/internal/sync"
	"github.com/samakintunde/bujo-cli/internal/tui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfg config.Config

var rootCmd = &cobra.Command{
	Use:   "bujo",
	Short: "A simple CLI tool for managing your daily tasks",
	Long:  `Bujo is a simple CLI tool for managing your daily tasks. It allows you to create, edit, and delete tasks, notes and events.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := initializeConfig(cmd)
		if err != nil {
			return err
		}
		db, err := storage.NewDBStore(cfg.GetDBPath())
		if err != nil {
			return err
		}
		fs, err := storage.NewFSStore(cfg.GetJournalPath())
		if err != nil {
			return err
		}
		syncer := sync.NewSyncer(cfg.GetJournalPath(), db)
		if err := syncer.Sync(); err != nil {
			return err
		}

		app := tui.NewApp(db, fs, syncer)
		p := tea.NewProgram(app, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			fmt.Printf("Error running TUI: %v", err)
			return err
		}
		return nil
	},
}

func Execute() error {
	return rootCmd.Execute()
}

var cfgFilePath string
var verbose bool

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFilePath, "config", "", "config file (default location: ./config.yaml, $HOME/.config/bujo/config.yaml)")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "verbose output")
}

func initializeConfig(cmd *cobra.Command) error {
	viper.SetEnvPrefix("BUJO")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "*", "-", "*"))
	viper.AutomaticEnv()

	if cfgFilePath != "" {
		viper.SetConfigFile(cfgFilePath)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)
		viper.AddConfigPath(".")
		viper.AddConfigPath(filepath.Join(home, ".config", "bujo"))
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")

		viper.SetDefault("path", filepath.Join(home, ".bujo"))
	}

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Printf("Error reading config file: %s\n", err)
			return err
		}
	}

	err := viper.Unmarshal(&cfg)
	if err != nil {
		fmt.Print("unable to decode config")
		return err
	}

	err = viper.BindPFlags(cmd.Flags())
	if err != nil {
		return err
	}

	// TODO: Add a verbose logger
	if verbose {
		fmt.Println("Configuration initialized. Using config file:", viper.ConfigFileUsed())
	}

	return nil
}
