package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"gopkg.in/yaml.v3"
)

func LoadSettings(configPath string) (*Settings, error) {
	// If file doesn't exist, return defaults silently
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return DefaultSettings(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Start with defaults and unmarshal on top
	settings := DefaultSettings()
	if err := yaml.Unmarshal(data, settings); err != nil {
		return nil, formatYAMLError(err, configPath)
	}

	// Validate the loaded settings
	if err := ValidateSettings(settings); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return settings, nil
}

// creates a default configuration file at the specified path
func GenerateDefaultConfig(configPath string) error {
	// Ensure the directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write the default template
	template := DefaultConfigTemplate()
	if err := os.WriteFile(configPath, []byte(template), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// attempts to extract line number information from YAML errors
func formatYAMLError(err error, configPath string) error {
	errStr := err.Error()

	// Try to extract line information from yaml.v3 errors
	if strings.Contains(errStr, "line") {
		return fmt.Errorf("invalid YAML syntax in %s: %s", configPath, errStr)
	}

	return fmt.Errorf("failed to parse YAML in %s: %w", configPath, err)
}

// returns the path to the configuration file
func (c *Config) ConfigPath() string {
	return filepath.Join(c.ConfigDir, "config.yaml")
}

// loads configuration from the standard location or returns defaults
func (c *Config) LoadOrDefault() (*Settings, error) {
	return LoadSettings(c.ConfigPath())
}
