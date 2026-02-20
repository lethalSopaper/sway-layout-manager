package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lidsol/sway-layout-manager/internal/config"
)

// tests that respects XDG Base Directory environment variables
func TestXDGEnvironmentVariables(t *testing.T) {
	t.Run("respects_XDG_CONFIG_HOME", func(t *testing.T) {
		tmpDir := t.TempDir()
		configDir := filepath.Join(tmpDir, "config")
		if err := os.MkdirAll(configDir, 0755); err != nil {
			t.Fatalf("Failed to create temp config directory: %v", err)
		}

		// set XDG_CONFIG_HOME
		originalConfigHome := os.Getenv("XDG_CONFIG_HOME")
		defer os.Setenv("XDG_CONFIG_HOME", originalConfigHome)
		os.Setenv("XDG_CONFIG_HOME", configDir)

		// create config - should use our XDG_CONFIG_HOME
		cfg, err := config.NewConfig()
		if err != nil {
			t.Fatalf("Failed to create config: %v", err)
		}
		if cfg == nil {
			t.Fatal("Config should not be nil")
		}

		expectedPath := filepath.Join(configDir, "sway-layout-manager")
		if cfg.ConfigDir != expectedPath {
			t.Errorf("Expected config directory %s, got %s", expectedPath, cfg.ConfigDir)
		}
	})

	t.Run("respects_XDG_DATA_HOME", func(t *testing.T) {
		tmpDir := t.TempDir()
		dataDir := filepath.Join(tmpDir, "data")
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			t.Fatalf("Failed to create temp data directory: %v", err)
		}

		// set XDG_DATA_HOME
		originalDataHome := os.Getenv("XDG_DATA_HOME")
		defer os.Setenv("XDG_DATA_HOME", originalDataHome)
		os.Setenv("XDG_DATA_HOME", dataDir)

		// create config - should use our XDG_DATA_HOME
		cfg, err := config.NewConfig()
		if err != nil {
			t.Fatalf("Failed to create config: %v", err)
		}
		if cfg == nil {
			t.Fatal("Config should not be nil")
		}

		expectedPath := filepath.Join(dataDir, "sway-layout-manager")
		if cfg.DataDir != expectedPath {
			t.Errorf("Expected data directory %s, got %s", expectedPath, cfg.DataDir)
		}
	})

	t.Run("falls_back_to_home_directory", func(t *testing.T) {
		// clear XDG variables
		originalConfigHome := os.Getenv("XDG_CONFIG_HOME")
		originalDataHome := os.Getenv("XDG_DATA_HOME")
		defer func() {
			os.Setenv("XDG_CONFIG_HOME", originalConfigHome)
			os.Setenv("XDG_DATA_HOME", originalDataHome)
		}()
		os.Unsetenv("XDG_CONFIG_HOME")
		os.Unsetenv("XDG_DATA_HOME")

		cfg, err := config.NewConfig()
		if err != nil {
			t.Fatalf("Failed to create config: %v", err)
		}
		if cfg == nil {
			t.Fatal("Config should not be nil")
		}

		homeDir, err := os.UserHomeDir()
		if err != nil {
			t.Fatalf("Failed to get home directory: %v", err)
		}

		expectedConfigDir := filepath.Join(homeDir, ".config", "sway-layout-manager")
		expectedDataDir := filepath.Join(homeDir, ".local", "share", "sway-layout-manager")

		if cfg.ConfigDir != expectedConfigDir {
			t.Errorf("Expected config directory %s, got %s", expectedConfigDir, cfg.ConfigDir)
		}
		if cfg.DataDir != expectedDataDir {
			t.Errorf("Expected data directory %s, got %s", expectedDataDir, cfg.DataDir)
		}
	})
}

// tests that configuration can be saved and loaded correctly
func TestConfigurationPersistence(t *testing.T) {
	t.Run("creates_directories_correctly", func(t *testing.T) {
		// setup isolated environment
		tmpDir := t.TempDir()
		configDir := filepath.Join(tmpDir, "config")
		dataDir := filepath.Join(tmpDir, "data")

		originalConfigHome := os.Getenv("XDG_CONFIG_HOME")
		originalDataHome := os.Getenv("XDG_DATA_HOME")
		defer func() {
			os.Setenv("XDG_CONFIG_HOME", originalConfigHome)
			os.Setenv("XDG_DATA_HOME", originalDataHome)
		}()

		os.Setenv("XDG_CONFIG_HOME", configDir)
		os.Setenv("XDG_DATA_HOME", dataDir)

		// create configuration
		cfg, err := config.NewConfig()
		if err != nil {
			t.Fatalf("Failed to create config: %v", err)
		}
		if cfg == nil {
			t.Fatal("Config should not be nil")
		}

		// verify paths
		expectedConfigDir := filepath.Join(configDir, "sway-layout-manager")
		expectedDataDir := filepath.Join(dataDir, "sway-layout-manager")

		if cfg.ConfigDir != expectedConfigDir {
			t.Errorf("Expected config directory %s, got %s", expectedConfigDir, cfg.ConfigDir)
		}
		if cfg.DataDir != expectedDataDir {
			t.Errorf("Expected data directory %s, got %s", expectedDataDir, cfg.DataDir)
		}

		// check that directories exist
		if _, err := os.Stat(cfg.ConfigDir); os.IsNotExist(err) {
			t.Errorf("Config directory was not created: %s", cfg.ConfigDir)
		}
		if _, err := os.Stat(cfg.DataDir); os.IsNotExist(err) {
			t.Errorf("Data directory was not created: %s", cfg.DataDir)
		}
	})

	t.Run("succeeds_without_existing_files", func(t *testing.T) {
		tmpDir := t.TempDir()
		configDir := filepath.Join(tmpDir, "config")
		dataDir := filepath.Join(tmpDir, "data")

		originalConfigHome := os.Getenv("XDG_CONFIG_HOME")
		originalDataHome := os.Getenv("XDG_DATA_HOME")
		defer func() {
			os.Setenv("XDG_CONFIG_HOME", originalConfigHome)
			os.Setenv("XDG_DATA_HOME", originalDataHome)
		}()

		os.Setenv("XDG_CONFIG_HOME", configDir)
		os.Setenv("XDG_DATA_HOME", dataDir)

		// don't create any files, test directory creation
		cfg, err := config.NewConfig()
		if err != nil {
			t.Fatalf("Failed to create config: %v", err)
		}
		if cfg == nil {
			t.Fatal("Config should not be nil even without existing directories")
		}

		// should have valid paths set
		if cfg.ConfigDir == "" {
			t.Error("ConfigDir should not be empty")
		}
		if cfg.DataDir == "" {
			t.Error("DataDir should not be empty")
		}

		// directories should be created
		if _, err := os.Stat(cfg.ConfigDir); os.IsNotExist(err) {
			t.Errorf("Config directory should be created: %s", cfg.ConfigDir)
		}
		if _, err := os.Stat(cfg.DataDir); os.IsNotExist(err) {
			t.Errorf("Data directory should be created: %s", cfg.DataDir)
		}
	})
}

// tests that the application provides sensible defaults when no config exists
func TestDefaultConfiguration(t *testing.T) {
	t.Run("provides_default_values", func(t *testing.T) {
		// setup isolated environment
		tmpDir := t.TempDir()
		configDir := filepath.Join(tmpDir, "config")
		dataDir := filepath.Join(tmpDir, "data")

		originalConfigHome := os.Getenv("XDG_CONFIG_HOME")
		originalDataHome := os.Getenv("XDG_DATA_HOME")
		defer func() {
			os.Setenv("XDG_CONFIG_HOME", originalConfigHome)
			os.Setenv("XDG_DATA_HOME", originalDataHome)
		}()

		os.Setenv("XDG_CONFIG_HOME", configDir)
		os.Setenv("XDG_DATA_HOME", dataDir)

		cfg, err := config.NewConfig()
		if err != nil {
			t.Fatalf("Failed to create config: %v", err)
		}
		if cfg == nil {
			t.Fatal("Config should not be nil")
		}

		// check that essential paths are set
		if cfg.ConfigDir == "" {
			t.Error("Default ConfigDir should not be empty")
		}
		if cfg.DataDir == "" {
			t.Error("Default DataDir should not be empty")
		}
		if cfg.DefaultName == "" {
			t.Error("Default DefaultName should not be empty")
		}

		expectedConfigDir := filepath.Join(configDir, "sway-layout-manager")
		expectedDataDir := filepath.Join(dataDir, "sway-layout-manager")

		if cfg.ConfigDir != expectedConfigDir {
			t.Errorf("Expected default config directory %s, got %s", expectedConfigDir, cfg.ConfigDir)
		}
		if cfg.DataDir != expectedDataDir {
			t.Errorf("Expected default data directory %s, got %s", expectedDataDir, cfg.DataDir)
		}
	})

	t.Run("creates_necessary_directories", func(t *testing.T) {
		tmpDir := t.TempDir()
		configDir := filepath.Join(tmpDir, "config")
		dataDir := filepath.Join(tmpDir, "data")

		originalConfigHome := os.Getenv("XDG_CONFIG_HOME")
		originalDataHome := os.Getenv("XDG_DATA_HOME")
		defer func() {
			os.Setenv("XDG_CONFIG_HOME", originalConfigHome)
			os.Setenv("XDG_DATA_HOME", originalDataHome)
		}()

		os.Setenv("XDG_CONFIG_HOME", configDir)
		os.Setenv("XDG_DATA_HOME", dataDir)

		cfg, err := config.NewConfig()
		if err != nil {
			t.Fatalf("Failed to create config: %v", err)
		}
		if cfg == nil {
			t.Fatal("Config should not be nil")
		}

		// check that data directory was created
		dataDirectory := cfg.DataDir
		stat, err := os.Stat(dataDirectory)
		if err != nil {
			t.Fatalf("Data directory should be created: %v", err)
		}
		if !stat.IsDir() {
			t.Error("DataDir should be a directory")
		}

		// check permissions
		mode := stat.Mode()
		expectedPerm := os.FileMode(0755)
		if mode.Perm() != expectedPerm {
			t.Errorf("Expected directory permissions %o, got %o", expectedPerm, mode.Perm())
		}

		// check that config directory also exists
		configDirectory := cfg.ConfigDir
		if _, err := os.Stat(configDirectory); os.IsNotExist(err) {
			t.Errorf("Config directory should be created: %s", configDirectory)
		}
	})
}

// tests the complete directory structure created by the application
func TestConfigurationDirectoryStructure(t *testing.T) {
	t.Run("creates_complete_directory_structure", func(t *testing.T) {
		tmpDir := t.TempDir()
		configDir := filepath.Join(tmpDir, "config")
		dataDir := filepath.Join(tmpDir, "data")

		originalConfigHome := os.Getenv("XDG_CONFIG_HOME")
		originalDataHome := os.Getenv("XDG_DATA_HOME")
		defer func() {
			os.Setenv("XDG_CONFIG_HOME", originalConfigHome)
			os.Setenv("XDG_DATA_HOME", originalDataHome)
		}()

		os.Setenv("XDG_CONFIG_HOME", configDir)
		os.Setenv("XDG_DATA_HOME", dataDir)

		// create config
		cfg, err := config.NewConfig()
		if err != nil {
			t.Fatalf("Failed to create config: %v", err)
		}
		if cfg == nil {
			t.Fatal("Config should not be nil")
		}

		expectedDirs := []string{
			filepath.Join(configDir, "sway-layout-manager"),
			filepath.Join(dataDir, "sway-layout-manager"),
		}

		for _, dir := range expectedDirs {
			stat, err := os.Stat(dir)
			if err != nil {
				t.Errorf("Directory should exist: %s - %v", dir, err)
				continue
			}
			if !stat.IsDir() {
				t.Errorf("Path should be a directory: %s", dir)
			}
		}
	})

	t.Run("directory_permissions_are_correct", func(t *testing.T) {
		tmpDir := t.TempDir()
		configDir := filepath.Join(tmpDir, "config")
		dataDir := filepath.Join(tmpDir, "data")

		originalConfigHome := os.Getenv("XDG_CONFIG_HOME")
		originalDataHome := os.Getenv("XDG_DATA_HOME")
		defer func() {
			os.Setenv("XDG_CONFIG_HOME", originalConfigHome)
			os.Setenv("XDG_DATA_HOME", originalDataHome)
		}()

		os.Setenv("XDG_CONFIG_HOME", configDir)
		os.Setenv("XDG_DATA_HOME", dataDir)

		// create config
		cfg, err := config.NewConfig()
		if err != nil {
			t.Fatalf("Failed to create config: %v", err)
		}
		if cfg == nil {
			t.Fatal("Config should not be nil")
		}

		// check data directory permissions
		dataDirectory := cfg.DataDir
		stat, err := os.Stat(dataDirectory)
		if err != nil {
			t.Fatalf("Failed to stat data directory: %v", err)
		}

		expectedPerm := os.FileMode(0755)
		if stat.Mode().Perm() != expectedPerm {
			t.Errorf("Expected data directory permissions %o, got %o", expectedPerm, stat.Mode().Perm())
		}
	})
}

// tests that multiple config instances don't interfere with each other
func TestConfigurationIsolation(t *testing.T) {
	t.Run("multiple_configs_with_different_environments", func(t *testing.T) {
		tmpDir1 := t.TempDir()
		tmpDir2 := t.TempDir()

		// first environment
		configDir1 := filepath.Join(tmpDir1, "config")
		dataDir1 := filepath.Join(tmpDir1, "data")

		// second environment
		configDir2 := filepath.Join(tmpDir2, "config")
		dataDir2 := filepath.Join(tmpDir2, "data")

		originalConfigHome := os.Getenv("XDG_CONFIG_HOME")
		originalDataHome := os.Getenv("XDG_DATA_HOME")
		defer func() {
			os.Setenv("XDG_CONFIG_HOME", originalConfigHome)
			os.Setenv("XDG_DATA_HOME", originalDataHome)
		}()

		// create config from first environment
		os.Setenv("XDG_CONFIG_HOME", configDir1)
		os.Setenv("XDG_DATA_HOME", dataDir1)
		cfg1, err := config.NewConfig()
		if err != nil {
			t.Fatalf("Failed to create config 1: %v", err)
		}

		// create config from second environment
		os.Setenv("XDG_CONFIG_HOME", configDir2)
		os.Setenv("XDG_DATA_HOME", dataDir2)
		cfg2, err := config.NewConfig()
		if err != nil {
			t.Fatalf("Failed to create config 2: %v", err)
		}

		if cfg1.ConfigDir == cfg2.ConfigDir {
			t.Error("Configs should have different ConfigDir values")
		}
		if cfg1.DataDir == cfg2.DataDir {
			t.Error("Configs should have different DataDir values")
		}

		// verify first environment paths
		expectedConfigDir1 := filepath.Join(configDir1, "sway-layout-manager")
		expectedDataDir1 := filepath.Join(dataDir1, "sway-layout-manager")
		if cfg1.ConfigDir != expectedConfigDir1 {
			t.Errorf("Config 1: Expected directory %s, got %s", expectedConfigDir1, cfg1.ConfigDir)
		}
		if cfg1.DataDir != expectedDataDir1 {
			t.Errorf("Config 1: Expected data dir %s, got %s", expectedDataDir1, cfg1.DataDir)
		}

		// verify second environment paths
		expectedConfigDir2 := filepath.Join(configDir2, "sway-layout-manager")
		expectedDataDir2 := filepath.Join(dataDir2, "sway-layout-manager")
		if cfg2.ConfigDir != expectedConfigDir2 {
			t.Errorf("Config 2: Expected directory %s, got %s", expectedConfigDir2, cfg2.ConfigDir)
		}
		if cfg2.DataDir != expectedDataDir2 {
			t.Errorf("Config 2: Expected data dir %s, got %s", expectedDataDir2, cfg2.DataDir)
		}

		// verify both data directories were created independently
		if _, err := os.Stat(cfg1.DataDir); os.IsNotExist(err) {
			t.Error("Config 1 data directory should exist")
		}
		if _, err := os.Stat(cfg2.DataDir); os.IsNotExist(err) {
			t.Error("Config 2 data directory should exist")
		}
	})
}
