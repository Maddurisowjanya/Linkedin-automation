package config

import (
	"errors"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config is the root configuration structure for the PoC.
type Config struct {
	Browser  BrowserConfig  `yaml:"browser"`
	Database DatabaseConfig `yaml:"database"`
	Search   SearchConfig   `yaml:"search"`
	Connect  ConnectConfig  `yaml:"connect"`
	Messaging MessagingConfig `yaml:"messaging"`
}

type BrowserConfig struct {
	Headless      bool   `yaml:"headless"`
	ViewportWidth  int    `yaml:"viewport_width"`
	ViewportHeight int    `yaml:"viewport_height"`
	// If empty, a random realistic UA will be generated.
	UserAgent string `yaml:"user_agent"`
}

type DatabaseConfig struct {
	DSN string `yaml:"dsn"`
}

type SearchConfig struct {
	Keywords      []string      `yaml:"keywords"`
	MaxPages      int           `yaml:"max_pages"`
	PageDelayMin  time.Duration `yaml:"page_delay_min"`
	PageDelayMax  time.Duration `yaml:"page_delay_max"`
}

type ConnectConfig struct {
	DailyLimit        int           `yaml:"daily_limit"`
	NoteTemplate      string        `yaml:"note_template"`
	ActionDelayMin    time.Duration `yaml:"action_delay_min"`
	ActionDelayMax    time.Duration `yaml:"action_delay_max"`
}

type MessagingConfig struct {
	Templates        []string      `yaml:"templates"`
	CheckInterval    time.Duration `yaml:"check_interval"`
	DailyLimit       int           `yaml:"daily_limit"`
	ActionDelayMin   time.Duration `yaml:"action_delay_min"`
	ActionDelayMax   time.Duration `yaml:"action_delay_max"`
}

// Load reads YAML config from disk, applies environment overrides and
// sensible defaults. Time durations are parsed by the yaml library using the
// Go duration syntax such as "1s", "500ms", "2m".
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	applyDefaults(&cfg)
	if err := validate(&cfg); err != nil {
		return nil, err
	}

	// Example of simple environment override for DSN:
	if v := os.Getenv("SQLITE_DSN"); v != "" {
		cfg.Database.DSN = v
	}

	return &cfg, nil
}

func applyDefaults(cfg *Config) {
	if cfg.Browser.ViewportWidth == 0 {
		cfg.Browser.ViewportWidth = 1366
	}
	if cfg.Browser.ViewportHeight == 0 {
		cfg.Browser.ViewportHeight = 768
	}
	if cfg.Search.MaxPages == 0 {
		cfg.Search.MaxPages = 1
	}
	if cfg.Connect.DailyLimit == 0 {
		cfg.Connect.DailyLimit = 10
	}
	if cfg.Messaging.DailyLimit == 0 {
		cfg.Messaging.DailyLimit = 10
	}
	if cfg.Search.PageDelayMin == 0 {
		cfg.Search.PageDelayMin = 2 * time.Second
	}
	if cfg.Search.PageDelayMax == 0 {
		cfg.Search.PageDelayMax = 5 * time.Second
	}
	if cfg.Connect.ActionDelayMin == 0 {
		cfg.Connect.ActionDelayMin = 2 * time.Second
	}
	if cfg.Connect.ActionDelayMax == 0 {
		cfg.Connect.ActionDelayMax = 5 * time.Second
	}
	if cfg.Messaging.ActionDelayMin == 0 {
		cfg.Messaging.ActionDelayMin = 2 * time.Second
	}
	if cfg.Messaging.ActionDelayMax == 0 {
		cfg.Messaging.ActionDelayMax = 5 * time.Second
	}
	if cfg.Database.DSN == "" {
		cfg.Database.DSN = "file:linkedin_poc.db?_fk=1"
	}
}

func validate(cfg *Config) error {
	if len(cfg.Search.Keywords) == 0 {
		return errors.New("at least one search keyword must be configured")
	}
	if cfg.Connect.DailyLimit < 0 {
		return errors.New("connect.daily_limit cannot be negative")
	}
	if cfg.Messaging.DailyLimit < 0 {
		return errors.New("messaging.daily_limit cannot be negative")
	}
	return nil
}



