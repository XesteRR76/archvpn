package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// DefaultSubscriptions provides an initial placeholder or empty list
var DefaultSubscriptions = []string{}

type Config struct {
	ConfigDir          string   `json:"config_dir,omitempty"`
	SubscriptionURLs   []string `json:"subscription_urls"`
	LastSelectedNodeID string   `json:"last_selected_node_id,omitempty"`
	TunInterface       string   `json:"tun_interface"`
	DNSServers         []string `json:"dns_servers"`
}

// DefaultConfig returns standard default configuration
func DefaultConfig() *Config {
	home, _ := os.UserHomeDir()
	configDir := filepath.Join(home, ".config", "archvpn")
	return &Config{
		ConfigDir:        configDir,
		SubscriptionURLs: DefaultSubscriptions,
		TunInterface:     "tun0",
		DNSServers:       []string{"1.1.1.1", "8.8.8.8"},
	}
}

// LoadConfig loads configuration from disk or returns defaults
func LoadConfig() (*Config, error) {
	cfg := DefaultConfig()
	configFile := filepath.Join(cfg.ConfigDir, "config.json")

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		_ = cfg.Save()
		return cfg, nil
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return cfg, err
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return cfg, err
	}

	if len(cfg.SubscriptionURLs) == 0 {
		cfg.SubscriptionURLs = DefaultSubscriptions
	}
	if cfg.TunInterface == "" {
		cfg.TunInterface = "tun0"
	}
	if len(cfg.DNSServers) == 0 {
		cfg.DNSServers = []string{"1.1.1.1", "8.8.8.8"}
	}

	return cfg, nil
}

// Save writes current config to config.json
func (c *Config) Save() error {
	if c.ConfigDir == "" {
		home, _ := os.UserHomeDir()
		c.ConfigDir = filepath.Join(home, ".config", "archvpn")
	}

	if err := os.MkdirAll(c.ConfigDir, 0755); err != nil {
		return err
	}

	configFile := filepath.Join(c.ConfigDir, "config.json")
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configFile, data, 0644)
}
