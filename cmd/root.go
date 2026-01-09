package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/samakintunde/bujo-cli/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfg config.Config

var rootCmd = &cobra.Command{
	Use:   "bujo",
	Short: "A simple CLI tool for managing your daily tasks",
	Long:  `Bujo is a simple CLI tool for managing your daily tasks. It allows you to create, edit, and delete tasks, notes and events.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return initializeConfig(cmd)
	},
}

func Execute() error {
	return rootCmd.Execute()
}

var cfgFilePath string

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFilePath, "config", "", "config file (default location: ./config.yaml, $HOME/.config/bujo/config.yaml)")
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
		viper.AddConfigPath(filepath.Join(home, ".config", ".bujo"))
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")

		viper.SetDefault("path", filepath.Join(home, ".bujo"))
		viper.SetDefault("db.path", filepath.Join(home, ".bujo", "db"))
		viper.SetDefault("journal.path", filepath.Join(home, ".bujo", "journal"))
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

	fmt.Println("Configuration initialized. Using config file:", viper.ConfigFileUsed())

	return nil
}
