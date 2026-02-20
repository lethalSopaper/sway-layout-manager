package integration

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"github.com/lidsol/sway-layout-manager/internal/config"
	"github.com/lidsol/sway-layout-manager/internal/layout"
	"github.com/lidsol/sway-layout-manager/pkg/preset"
)

// tests that XDG directories are created correctly
func TestXDGDirectoryCreation(t *testing.T) {
	testDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// create config and verify directories creation
	cfg, err := config.NewConfig()
	if err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	// XDG data directory
	expectedDataDir := filepath.Join(testDir, "data", "sway-layout-manager")

	// check the preset directory gets created whensomething is saved
	manager := preset.NewManager(cfg)

	// create a simple test preset
	testPreset := &layout.Preset{
		Name: "test-xdg-creation",
		CreatedAt: time.Now(),
		SwayVersion: "1.8.1",
		Outputs: []layout.OutputLayout{},
		Workspaces: []layout.WorkspaceLayout{},
	}

	err = manager.Save(testPreset)
	if err != nil {
		t.Fatalf("Failed to save test preset: %v", err)
	}

	// verify directory was created
	if _, err := os.Stat(expectedDataDir); os.IsNotExist(err) {
		t.Errorf("XDG data directory was not created: %s", expectedDataDir)
	}

	// verify file was created
	expectedFile := filepath.Join(expectedDataDir, "test-xdg-creation.json")
	if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
		t.Errorf("Preset file was not created: %s", expectedFile)
	}

	t.Logf("Successfully created XDG directory structure in: %s", expectedDataDir)
}

// tests that preset files persist correctly across operations
func TestPresetFilePersistence(t *testing.T) {
	testDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	cfg, err := config.NewConfig()
	if err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	manager := preset.NewManager(cfg)

	// create multiple test presets
	presets := map[string]*layout.Preset{
		"workspace-dev": {
			Name: "workspace-dev",
			CreatedAt: time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC),
			SwayVersion: "1.8.1",
			Outputs: []layout.OutputLayout{
				{
					Name: "HDMI-A-1",
					Active: true,
					Rect: layout.Rect{
						X: 0, Y: 0, Width: 1920, Height: 1080,
					},
				},
			},
			Workspaces: []layout.WorkspaceLayout{
				{
					Num: 1,
					Name: "dev",
					Output: "HDMI-A-1",
					Layout: "splith",
					Containers: []layout.ContainerLayout{},
				},
			},
		},
		"workspace-web": {
			Name: "workspace-web",
			CreatedAt: time.Date(2026, 1, 2, 15, 30, 0, 0, time.UTC),
			SwayVersion: "1.8.1",
			Outputs: []layout.OutputLayout{
				{
					Name: "DP-1",
					Active: true,
					Rect: layout.Rect{
						X: 1920, Y: 0, Width: 2560, Height: 1440,
					},
				},
			},
			Workspaces: []layout.WorkspaceLayout{
				{
					Num: 2,
					Name: "web",
					Output: "DP-1",
					Layout: "splitv",
					Containers: []layout.ContainerLayout{},
				},
			},
		},
	}

	// save all presets
	for name, preset := range presets {
		if err := manager.Save(preset); err != nil {
			t.Fatalf("Failed to save preset %s: %v", name, err)
		}
	}

	// verify all presets exist
	for name := range presets {
		exists := manager.Exists(name)
		if !exists {
			t.Errorf("Preset %s should exist after saving", name)
		}
	}

	// new manager instance
	manager2 := preset.NewManager(cfg)

	// verify all presets still exist and can be loaded
	for name, originalPreset := range presets {
		exists := manager2.Exists(name)
		if !exists {
			t.Errorf("Preset %s should persist after program restart", name)
			continue
		}

		loadedPreset, err := manager2.Load(name)
		if err != nil {
			t.Errorf("Failed to load persisted preset %s: %v", name, err)
			continue
		}

		// verify key fields match
		if loadedPreset.Name != originalPreset.Name {
			t.Errorf("Name mismatch for %s: expected %s, got %s", name, originalPreset.Name, loadedPreset.Name)
		}
		if loadedPreset.SwayVersion != originalPreset.SwayVersion {
			t.Errorf("SwayVersion mismatch for %s: expected %s, got %s", name, originalPreset.SwayVersion, loadedPreset.SwayVersion)
		}
		if len(loadedPreset.Outputs) != len(originalPreset.Outputs) {
			t.Errorf("Outputs count mismatch for %s: expected %d, got %d", name, len(originalPreset.Outputs), len(loadedPreset.Outputs))
		}
		if len(loadedPreset.Workspaces) != len(originalPreset.Workspaces) {
			t.Errorf("Workspaces count mismatch for %s: expected %d, got %d", name, len(originalPreset.Workspaces), len(loadedPreset.Workspaces))
		}
	}

	// list all presets
	allPresets, err := manager2.List()
	if err != nil {
		t.Fatalf("Failed to list presets: %v", err)
	}

	if len(allPresets) != len(presets) {
		t.Errorf("Expected %d presets in list, got %d", len(presets), len(allPresets))
	}

	t.Logf("Successfully verified persistence of %d presets in: %s", len(presets), testDir)
}

// tests that files are created with correct permissions
func TestFilePermissions(t *testing.T) {
	_, cleanup := setupTestEnvironment(t)
	defer cleanup()

	cfg, err := config.NewConfig()
	if err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	manager := preset.NewManager(cfg)

	// create test preset
	testPreset := &layout.Preset{
		Name: "permission-test",
		CreatedAt: time.Now(),
		SwayVersion: "1.8.1",
		Outputs: []layout.OutputLayout{},
		Workspaces: []layout.WorkspaceLayout{},
	}

	err = manager.Save(testPreset)
	if err != nil {
		t.Fatalf("Failed to save test preset: %v", err)
	}

	// check file permissions
	presetPath := cfg.PresetPath("permission-test")
	fileInfo, err := os.Stat(presetPath)
	if err != nil {
		t.Fatalf("Failed to stat preset file: %v", err)
	}

	mode := fileInfo.Mode()
	if mode.Perm()&0o400 == 0 {
		t.Error("Preset file should be readable by owner")
	}

	if mode.Perm()&0o200 == 0 {
		t.Error("Preset file should be writable by owner")
	}

	// check directory permissions
	dirPath := filepath.Dir(presetPath)
	dirInfo, err := os.Stat(dirPath)
	if err != nil {
		t.Fatalf("Failed to stat preset directory: %v", err)
	}

	dirMode := dirInfo.Mode()
	if !dirMode.IsDir() {
		t.Error("Preset path should be a directory")
	}

	if dirMode.Perm()&0o100 == 0 {
		t.Error("Preset directory should be accessible by owner")
	}

	t.Logf("File permissions: %v, Directory permissions: %v", mode.Perm(), dirMode.Perm())
}

// tests that saved files are valid JSON
func TestJSONFormatValidation(t *testing.T) {
	_, cleanup := setupTestEnvironment(t)
	defer cleanup()

	cfg, err := config.NewConfig()
	if err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	manager := preset.NewManager(cfg)

	// create complex test preset with various data types
	testPreset := &layout.Preset{
		Name: "json-validation-test",
		CreatedAt: time.Date(2026, 1, 2, 12, 30, 45, 123456789, time.UTC),
		SwayVersion: "sway version 1.8.1-abc123",
		Outputs: []layout.OutputLayout{
			{
				Name: "HDMI-A-1",
				Active: true,
				Rect: layout.Rect{X: 0, Y: 0, Width: 1920, Height: 1080},
			},
			{
				Name: "DP-1",
				Active: false,
				Rect: layout.Rect{X: -1920, Y: 0, Width: 1920, Height: 1080},
			},
		},
		Workspaces: []layout.WorkspaceLayout{
			{
				Num: 1,
				Name: "workspace with spaces",
				Output: "HDMI-A-1",
				Layout: "splith",
				Containers: []layout.ContainerLayout{
					{
						Type: "con",
						Layout: "splitv",
						Rect: layout.Rect{X: 0, Y: 25, Width: 960, Height: 1055},
					},
					{
						Type: "con",
						Layout: "none",
						Rect: layout.Rect{X: 960, Y: 25, Width: 960, Height: 1055},
					},
				},
			},
		},
	}

	err = manager.Save(testPreset)
	if err != nil {
		t.Fatalf("Failed to save test preset: %v", err)
	}

	// read file directly and verify JSON
	presetPath := cfg.PresetPath("json-validation-test")
	fileData, err := os.ReadFile(presetPath)
	if err != nil {
		t.Fatalf("Failed to read preset file: %v", err)
	}

	// verify it's valid JSON by unmarshaling
	var jsonData map[string]interface{}
	err = json.Unmarshal(fileData, &jsonData)
	if err != nil {
		t.Fatalf("Preset file contains invalid JSON: %v", err)
	}

	expectedFields := []string{"name", "created_at", "sway_version", "outputs", "workspaces"}
	for _, field := range expectedFields {
		if _, exists := jsonData[field]; !exists {
			t.Errorf("JSON missing required field: %s", field)
		}
	}

	// verify JSON is properly formatted
	jsonStr := string(fileData)
	if !strings.Contains(jsonStr, "\n") {
		t.Error("JSON should be formatted with newlines for readability")
	}

	t.Logf("JSON file is valid and properly formatted (%d bytes)", len(fileData))
}

// tests behavior when multiple operations happen simultaneously
func TestConcurrentAccess(t *testing.T) {
	testDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	cfg, err := config.NewConfig()
	if err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	numOperations := 5
	errChan := make(chan error, numOperations)

	// launch concurrent save operations
	for i := 0; i < numOperations; i++ {
		go func(index int) {
			manager := preset.NewManager(cfg)
			presetName := fmt.Sprintf("concurrent-test-%d", index)

			testPreset := &layout.Preset{
				Name: presetName,
				CreatedAt: time.Now(),
				SwayVersion: fmt.Sprintf("1.8.%d", index),
				Outputs: []layout.OutputLayout{},
				Workspaces: []layout.WorkspaceLayout{},
			}

			err := manager.Save(testPreset)
			errChan <- err
		}(i)
	}

	// wait for all operations to complete
	var errors []error
	for i := 0; i < numOperations; i++ {
		if err := <-errChan; err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		t.Errorf("Concurrent operations failed: %v", errors)
	}

	// verify all presets were created
	manager := preset.NewManager(cfg)
	allPresets, err := manager.List()
	if err != nil {
		t.Fatalf("Failed to list presets after concurrent operations: %v", err)
	}

	if len(allPresets) != numOperations {
		t.Errorf("Expected %d presets after concurrent operations, got %d", numOperations, len(allPresets))
	}

	t.Logf("Successfully completed %d concurrent operations in: %s", numOperations, testDir)
}

// tests that delete operations actually remove files
func TestDeleteFileRemoval(t *testing.T) {
	_, cleanup := setupTestEnvironment(t)
	defer cleanup()

	cfg, err := config.NewConfig()
	if err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	manager := preset.NewManager(cfg)

	// test preset
	testPreset := &layout.Preset{
		Name: "delete-test",
		CreatedAt: time.Now(),
		SwayVersion: "1.8.1",
		Outputs: []layout.OutputLayout{},
		Workspaces: []layout.WorkspaceLayout{},
	}

	err = manager.Save(testPreset)
	if err != nil {
		t.Fatalf("Failed to save test preset: %v", err)
	}

	// verify file exists
	presetPath := cfg.PresetPath("delete-test")
	if _, err := os.Stat(presetPath); os.IsNotExist(err) {
		t.Fatalf("Preset file should exist before deletion: %s", presetPath)
	}

	// delete preset
	err = manager.Delete("delete-test")
	if err != nil {
		t.Fatalf("Failed to delete preset: %v", err)
	}

	// verify file no longer exists
	if _, err := os.Stat(presetPath); !os.IsNotExist(err) {
		t.Errorf("Preset file should not exist after deletion: %s", presetPath)
	}

	exists := manager.Exists("delete-test")
	if exists {
		t.Error("Deleted preset should not exist")
	}

	t.Logf("Successfully verified file deletion: %s", presetPath)
}