package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// Config contains configuration paths and settings
type Config struct {
	DataDir     string
	ConfigDir   string
	DefaultName string
	Settings    *Settings // loaded configuration settings
}

// constructor
func NewConfig() (*Config, error) {
	config := &Config{
		DefaultName: "layout-%s",
	}
	// directory where presets are stored
	if err := config.setupDataDir(); err != nil {
		return nil, fmt.Errorf("failed to setup data directory: %w", err)
	}
	// directory where config files are stored
	if err := config.setupConfigDir(); err != nil {
		return nil, fmt.Errorf("failed to setup config directory: %w", err)
	}

	// load configuration settings
	settings, err := config.LoadOrDefault()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}
	config.Settings = settings

	// override data directory if specified in config
	if settings.General.DataDir != "" {
		config.DataDir = settings.General.DataDir
		// ensure custom data directory exists
		if err := os.MkdirAll(config.DataDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create custom data directory %s: %w", config.DataDir, err)
		}
	}

	return config, nil
}

// creates and sets the data directory path
func (c *Config) setupDataDir() error {
	// use XDG_DATA_HOME if set, otherwise ~/.local/share
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		dataHome = filepath.Join(homeDir, ".local", "share")
	}

	c.DataDir = filepath.Join(dataHome, "sway-layout-manager")
	// create the directory if it doesn't exist
	if err := os.MkdirAll(c.DataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory %s: %w", c.DataDir, err)
	}

	return nil
}

// creates and sets the config directory path
func (c *Config) setupConfigDir() error {
	// use XDG_CONFIG_HOME if set, otherwise ~/.config
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		configHome = filepath.Join(homeDir, ".config")
	}

	c.ConfigDir = filepath.Join(configHome, "sway-layout-manager")

	// create the directory if it doesn't exist
	if err := os.MkdirAll(c.ConfigDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory %s: %w", c.ConfigDir, err)
	}

	return nil
}

// returns the full path for a preset file
func (c *Config) PresetPath(name string) string {
	// Ensure the name has .json extension
	if filepath.Ext(name) != ".json" {
		name += ".json"
	}
	return filepath.Join(c.DataDir, name)
}

// returns a default name with current timestamp
func (c *Config) DefaultPresetName() string {
	return fmt.Sprintf(c.DefaultName, fmt.Sprintf("%d", os.Getpid()))
}
