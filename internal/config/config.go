package config

type DBConfig struct {
	Path string `mapstructure:"path" yaml:"path"`
}

type JournalConfig struct {
	Path string `mapstructure:"path" yaml:"path"`
}

type Config struct {
	Path    string        `mapstructure:"path" yaml:"path"`
	DB      DBConfig      `mapstructure:"db" yaml:"db"`
	Journal JournalConfig `mapstructure:"journal" yaml:"journal"`
}
