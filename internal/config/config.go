package config

import "path/filepath"

type DBConfig struct {
}

type JournalConfig struct {
}

type Config struct {
	Path    string        `mapstructure:"path" yaml:"path"`
	DB      DBConfig      `mapstructure:"db" yaml:"db"`
	Journal JournalConfig `mapstructure:"journal" yaml:"journal"`
}

func (cfg *Config) GetDBPath() string {
	return filepath.Join(cfg.Path, "db")
}

func (cfg *Config) GetJournalPath() string {
	return filepath.Join(cfg.Path, "journal")
}
