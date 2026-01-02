package unit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"github.com/lidsol/sway-layout-manager/internal/config"
)

func TestNewConfig(t *testing.T) {
	// environment variables to restore after test
	originalDataHome := os.Getenv("XDG_DATA_HOME")
	originalConfigHome := os.Getenv("XDG_CONFIG_HOME")

	defer func() {
		os.Setenv("XDG_DATA_HOME", originalDataHome)
		os.Setenv("XDG_CONFIG_HOME", originalConfigHome)
	}()

	t.Run("with default XDG paths", func(t *testing.T) {
		// clear XDG environment variables to test defaults
		os.Unsetenv("XDG_DATA_HOME")
		os.Unsetenv("XDG_CONFIG_HOME")

		cfg, err := config.NewConfig()
		if err != nil {
			t.Fatalf("NewConfig() failed: %v", err)
		}

		if cfg.DefaultName != "layout-%s" {
			t.Errorf("Expected DefaultName to be 'layout-%%s', got %s", cfg.DefaultName)
		}

		// check if DataDir contains the expected path
		if !strings.Contains(cfg.DataDir, ".local/share/sway-layout-manager") {
			t.Errorf("DataDir should contain '.local/share/sway-layout-manager', got %s", cfg.DataDir)
		}

		// check if ConfigDir contains the expected path
		if !strings.Contains(cfg.ConfigDir, ".config/sway-layout-manager") {
			t.Errorf("ConfigDir should contain '.config/sway-layout-manager', got %s", cfg.ConfigDir)
		}
	})

	t.Run("with custom XDG paths", func(t *testing.T) {
		// set custom XDG paths
		customDataHome := "/tmp/test-data"
		customConfigHome := "/tmp/test-config"
		os.Setenv("XDG_DATA_HOME", customDataHome)
		os.Setenv("XDG_CONFIG_HOME", customConfigHome)

		cfg, err := config.NewConfig()
		if err != nil {
			t.Fatalf("NewConfig() failed: %v", err)
		}

		expectedDataDir := filepath.Join(customDataHome, "sway-layout-manager")
		if cfg.DataDir != expectedDataDir {
			t.Errorf("Expected DataDir to be %s, got %s", expectedDataDir, cfg.DataDir)
		}

		expectedConfigDir := filepath.Join(customConfigHome, "sway-layout-manager")
		if cfg.ConfigDir != expectedConfigDir {
			t.Errorf("Expected ConfigDir to be %s, got %s", expectedConfigDir, cfg.ConfigDir)
		}
	})
}

func TestConfigPresetPath(t *testing.T) {
	// create a temporary config for testing
	cfg := &config.Config{
		DataDir: "/tmp/test-presets",
	}

	tests := []struct {
		description string
		input string
		expected string
	}{
		{
			description: "name without extension",
			input: "my-layout",
			expected: "/tmp/test-presets/my-layout.json",
		},
		{
			description: "name with .json extension",
			input: "my-layout.json",
			expected: "/tmp/test-presets/my-layout.json",
		},
		{
			description: "name with other extension",
			input: "my-layout.txt",
			expected: "/tmp/test-presets/my-layout.txt.json",
		},
		{
			description: "empty name",
			input: "",
			expected: "/tmp/test-presets/.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			result := cfg.PresetPath(tt.input)
			if result != tt.expected {
				t.Errorf("PresetPath(%s) = %s, expected %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestConfigDefaultPresetName(t *testing.T) {
	cfg := &config.Config{
		DefaultName: "layout-%s",
	}

	result := cfg.DefaultPresetName()

	// verifies the format
	if !strings.HasPrefix(result, "layout-") {
		t.Errorf("DefaultPresetName() should start with 'layout-', got %s", result)
	}

	// verifies the PID part
	parts := strings.Split(result, "-")
	if len(parts) != 2 {
		t.Errorf("DefaultPresetName() should have format 'layout-<pid>', got %s", result)
	}

	if parts[1] == "" {
		t.Errorf("DefaultPresetName() should contain a PID, got %s", result)
	}
}